package mysql_projector_dao

import (
	"context"
	"database/sql"
	"log/slog"

	appErr "wallet/wallet-service/pkg/app_err"
	integrationEvent "wallet/wallet-service/pkg/platform/integration_event"
)

type WalletMySQLProjectionDAO struct {
	db *sql.DB
}

func NewWalletMySQLProjectionDAO(db *sql.DB) *WalletMySQLProjectionDAO {
	return &WalletMySQLProjectionDAO{
		db: db,
	}
}

func (dao *WalletMySQLProjectionDAO) CreateWallet(ctx context.Context, event integrationEvent.WalletCreatedEvent) error {
	_, err := dao.db.ExecContext(
		ctx,
		"INSERT INTO wallet_projections (wallet_id, holder_id, balance_in_cents, created_at, updated_at) VALUES (?, ?, 0, ?, ?)",
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

func (dao *WalletMySQLProjectionDAO) ApplyFundsTransferred(ctx context.Context, event integrationEvent.FundsTransferredEvent) error {
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
		"UPDATE wallet_projections SET balance_in_cents = balance_in_cents - ?, updated_at = ? WHERE wallet_id = ?",
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
		return appErr.ErrWalletProjectionNotFound
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO transfer_projections (wallet_id, transfer_id, counterpart_wallet_id, direction, category, amount_in_cents, transferred_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.FromWalletID,
		event.TransferID,
		event.ToWalletID,
		"outgoing",
		event.Category,
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

func (dao *WalletMySQLProjectionDAO) ApplyFundsTransferReceived(ctx context.Context, event integrationEvent.FundsTransferReceivedEvent) error {
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
		"UPDATE wallet_projections SET balance_in_cents = balance_in_cents + ?, updated_at = ? WHERE wallet_id = ?",
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
		return appErr.ErrWalletProjectionNotFound
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO transfer_projections (wallet_id, transfer_id, counterpart_wallet_id, direction, category, amount_in_cents, transferred_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.WalletID,
		event.TransferID,
		event.FromWalletID,
		"incoming",
		event.Category,
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
