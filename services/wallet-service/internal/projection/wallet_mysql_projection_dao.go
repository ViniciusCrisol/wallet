package projection

import (
	"context"
	"database/sql"
	"log/slog"

	"wallet/wallet-service/pkg/apperr"
	"wallet/wallet-service/pkg/integrationevent"
)

type WalletMySQLProjectionDAO struct {
	db *sql.DB
}

func NewWalletMySQLProjectionDAO(db *sql.DB) *WalletMySQLProjectionDAO {
	return &WalletMySQLProjectionDAO{
		db: db,
	}
}

func (dao *WalletMySQLProjectionDAO) CreateWallet(ctx context.Context, event integrationevent.WalletCreatedEvent) error {
	result, err := dao.db.ExecContext(
		ctx,
		"INSERT IGNORE INTO wallet_projections (wallet_id, holder_id, balance_in_cents, created_at, updated_at) VALUES (?, ?, 0, ?, ?)",
		event.WalletID,
		event.HolderID,
		event.CreatedAt,
		event.UpdatedAt,
	)
	if err != nil {
		slog.Error("failed to insert wallet projection", slog.String("wallet_id", event.WalletID), slog.String("error", err.Error()))
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(
			"failed to get rows affected for wallet projection insert",
			slog.String("wallet_id", event.WalletID),
			slog.String("error", err.Error()),
		)
		return err
	}
	if rowsAffected == 0 {
		slog.Info("wallet projection already exists, skipping", slog.String("wallet_id", event.WalletID))
	}
	return nil
}

func (dao *WalletMySQLProjectionDAO) ApplyFundsTransferred(ctx context.Context, event integrationevent.FundsTransferredEvent) error {
	tx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error(
			"failed to begin transaction for funds transferred projection",
			slog.String("transfer_id", event.TransferID),
			slog.String("error", err.Error()),
		)
		return err
	}
	defer tx.Rollback()

	insertResult, err := tx.ExecContext(
		ctx,
		"INSERT IGNORE INTO processed_events (event_id, event_type) VALUES (?, ?)",
		event.TransferID, integrationevent.FundsTransferredEventName,
	)
	if err != nil {
		slog.Error(
			"failed to insert processed event for funds transferred",
			slog.String("transfer_id", event.TransferID),
			slog.String("error", err.Error()),
		)
		return err
	}
	inserted, err := insertResult.RowsAffected()
	if err != nil {
		slog.Error(
			"failed to get rows affected for processed event insert",
			slog.String("transfer_id", event.TransferID),
			slog.String("error", err.Error()),
		)
		return err
	}
	if inserted == 0 {
		slog.Info(
			"event already processed, skipping",
			slog.String("transfer_id", event.TransferID),
			slog.String("event_type", integrationevent.FundsTransferredEventName),
		)
		return nil
	}

	result, err := tx.ExecContext(ctx,
		"UPDATE wallet_projections SET balance_in_cents = balance_in_cents - ?, updated_at = ? WHERE wallet_id = ?",
		event.AmountInCents,
		event.Timestamp,
		event.FromWalletID,
	)
	if err != nil {
		slog.Error(
			"failed to update wallet projection for funds transferred",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.FromWalletID),
			slog.String("error", err.Error()),
		)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(
			"failed to get rows affected for funds transferred update",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.FromWalletID),
			slog.String("error", err.Error()),
		)
		return err
	}
	if rowsAffected == 0 {
		slog.Warn(
			"wallet projection not found for funds transferred",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.FromWalletID),
		)
		return apperr.ErrWalletProjectionNotFound
	}

	return tx.Commit()
}

func (dao *WalletMySQLProjectionDAO) ApplyFundsTransferReceived(ctx context.Context, event integrationevent.FundsTransferReceivedEvent) error {
	tx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		slog.Error(
			"failed to begin transaction for funds transfer received projection",
			slog.String("transfer_id", event.TransferID),
			slog.String("error", err.Error()),
		)
		return err
	}
	defer tx.Rollback()

	insertResult, err := tx.ExecContext(
		ctx,
		"INSERT IGNORE INTO processed_events (event_id, event_type) VALUES (?, ?)",
		event.TransferID, integrationevent.FundsTransferReceivedEventName,
	)
	if err != nil {
		slog.Error(
			"failed to insert processed event for funds transfer received",
			slog.String("transfer_id", event.TransferID),
			slog.String("error", err.Error()),
		)
		return err
	}
	inserted, err := insertResult.RowsAffected()
	if err != nil {
		slog.Error(
			"failed to get rows affected for processed event insert",
			slog.String("transfer_id", event.TransferID),
			slog.String("error", err.Error()),
		)
		return err
	}
	if inserted == 0 {
		slog.Info(
			"event already processed, skipping",
			slog.String("transfer_id", event.TransferID),
			slog.String("event_type", integrationevent.FundsTransferReceivedEventName),
		)
		return nil
	}

	result, err := tx.ExecContext(ctx,
		"UPDATE wallet_projections SET balance_in_cents = balance_in_cents + ?, updated_at = ? WHERE wallet_id = ?",
		event.AmountInCents,
		event.Timestamp,
		event.WalletID,
	)
	if err != nil {
		slog.Error(
			"failed to update wallet projection for funds transfer received",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.WalletID),
			slog.String("error", err.Error()),
		)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(
			"failed to get rows affected for funds transfer received update",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.WalletID),
			slog.String("error", err.Error()),
		)
		return err
	}
	if rowsAffected == 0 {
		slog.Warn(
			"wallet projection not found for funds transfer received",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.WalletID),
		)
		return apperr.ErrWalletProjectionNotFound
	}

	return tx.Commit()
}
