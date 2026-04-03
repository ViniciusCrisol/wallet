package postgresqlprojectordao

import (
	"context"
	"database/sql"
	"log/slog"

	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/platform/integrationevent"
)

type WalletPostgreSQLProjectionDAO struct {
	db *sql.DB
}

func NewWalletPostgreSQLProjectionDAO(db *sql.DB) *WalletPostgreSQLProjectionDAO {
	return &WalletPostgreSQLProjectionDAO{
		db: db,
	}
}

func (dao *WalletPostgreSQLProjectionDAO) CreateWallet(ctx context.Context, event integrationevent.WalletCreatedEvent) error {
	_, err := dao.db.ExecContext(
		ctx,
		"INSERT INTO wallet_projections (wallet_id, holder_id, balance_in_cents, created_at, updated_at) VALUES ($1, $2, 0, $3, $4)",
		event.WalletID,
		event.HolderID,
		event.CreatedAt,
		event.UpdatedAt,
	)
	if err != nil {
		slog.Error("failed to insert wallet projection",
			slog.String("wallet_id", event.WalletID),
			slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (dao *WalletPostgreSQLProjectionDAO) ApplyFundsTransferred(ctx context.Context, event integrationevent.FundsTransferredEvent) error {
	tx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("failed to begin transaction for funds transferred",
			slog.String("transfer_id", event.TransferID),
			slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		"UPDATE wallet_projections SET balance_in_cents = balance_in_cents - $1, updated_at = $2 WHERE wallet_id = $3",
		event.AmountInCents,
		event.Timestamp,
		event.FromWalletID,
	)
	if err != nil {
		slog.Error("failed to update wallet projection for funds transferred",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.FromWalletID),
			slog.String("error", err.Error()))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("failed to get rows affected for funds transferred update",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.FromWalletID),
			slog.String("error", err.Error()))
		return err
	}
	if rowsAffected == 0 {
		slog.Warn("wallet projection not found for funds transferred",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.FromWalletID))
		return pkg.ErrWalletProjectionNotFound
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO transfer_projections (wallet_id, transfer_id, counterpart_wallet_id, direction, amount_in_cents, transferred_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		event.FromWalletID,
		event.TransferID,
		event.ToWalletID,
		"outgoing",
		event.AmountInCents,
		event.Timestamp,
	)
	if err != nil {
		slog.Error("failed to insert transfer projection",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.FromWalletID),
			slog.String("error", err.Error()))
		return err
	}
	return tx.Commit()
}

func (dao *WalletPostgreSQLProjectionDAO) ApplyFundsTransferReceived(ctx context.Context, event integrationevent.FundsTransferReceivedEvent) error {
	tx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error("failed to begin transaction for funds transfer received",
			slog.String("transfer_id", event.TransferID),
			slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		"UPDATE wallet_projections SET balance_in_cents = balance_in_cents + $1, updated_at = $2 WHERE wallet_id = $3",
		event.AmountInCents,
		event.Timestamp,
		event.WalletID,
	)
	if err != nil {
		slog.Error("failed to update wallet projection for funds transfer received",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.WalletID),
			slog.String("error", err.Error()))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("failed to get rows affected for funds transfer received update",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.WalletID),
			slog.String("error", err.Error()))
		return err
	}
	if rowsAffected == 0 {
		slog.Warn("wallet projection not found for funds transfer received",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.WalletID))
		return pkg.ErrWalletProjectionNotFound
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO transfer_projections (wallet_id, transfer_id, counterpart_wallet_id, direction, amount_in_cents, transferred_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		event.WalletID,
		event.TransferID,
		event.FromWalletID,
		"incoming",
		event.AmountInCents,
		event.Timestamp,
	)
	if err != nil {
		slog.Error("failed to insert transfer projection",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.WalletID),
			slog.String("error", err.Error()))
		return err
	}
	return tx.Commit()
}
