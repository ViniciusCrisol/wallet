package projection

import (
	"context"
	"database/sql"
	"fmt"
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
		return fmt.Errorf("inserting wallet projection %s: %w", event.WalletID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected for wallet projection %s: %w", event.WalletID, err)
	}
	if rowsAffected == 0 {
		slog.Info("wallet projection already exists, skipping", slog.String("wallet_id", event.WalletID))
	}
	return nil
}

func (dao *WalletMySQLProjectionDAO) ApplyFundsTransferred(ctx context.Context, event integrationevent.FundsTransferredEvent) error {
	tx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction for transfer %s: %w", event.TransferID, err)
	}
	defer tx.Rollback()

	insertResult, err := tx.ExecContext(
		ctx,
		"INSERT IGNORE INTO processed_events (event_id, event_type) VALUES (?, ?)",
		event.TransferID, integrationevent.FundsTransferredEventName,
	)
	if err != nil {
		return fmt.Errorf("recording processed event for transfer %s: %w", event.TransferID, err)
	}
	inserted, err := insertResult.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking processed event for transfer %s: %w", event.TransferID, err)
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
		return fmt.Errorf("debiting wallet %s for transfer %s: %w", event.FromWalletID, event.TransferID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected for transfer %s: %w", event.TransferID, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: wallet_id=%s, transfer_id=%s", apperr.ErrWalletProjectionNotFound, event.FromWalletID, event.TransferID)
	}

	return tx.Commit()
}

func (dao *WalletMySQLProjectionDAO) ApplyFundsTransferReceived(ctx context.Context, event integrationevent.FundsTransferReceivedEvent) error {
	tx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction for transfer %s: %w", event.TransferID, err)
	}
	defer tx.Rollback()

	insertResult, err := tx.ExecContext(
		ctx,
		"INSERT IGNORE INTO processed_events (event_id, event_type) VALUES (?, ?)",
		event.TransferID, integrationevent.FundsTransferReceivedEventName,
	)
	if err != nil {
		return fmt.Errorf("recording processed event for transfer %s: %w", event.TransferID, err)
	}
	inserted, err := insertResult.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking processed event for transfer %s: %w", event.TransferID, err)
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
		return fmt.Errorf("crediting wallet %s for transfer %s: %w", event.WalletID, event.TransferID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected for transfer %s: %w", event.TransferID, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: wallet_id=%s, transfer_id=%s", apperr.ErrWalletProjectionNotFound, event.WalletID, event.TransferID)
	}

	return tx.Commit()
}
