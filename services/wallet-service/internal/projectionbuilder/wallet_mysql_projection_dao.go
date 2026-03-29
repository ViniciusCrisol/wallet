package projectionbuilder

import (
	"context"
	"database/sql"
	"log/slog"

	"wallet/wallet-service/pkg"
)

type WalletMySQLProjectionDAO struct {
	db *sql.DB
}

func NewWalletMySQLProjectionDAO(db *sql.DB) *WalletMySQLProjectionDAO {
	return &WalletMySQLProjectionDAO{
		db: db,
	}
}

func (dao *WalletMySQLProjectionDAO) CreateWallet(ctx context.Context, event pkg.WalletCreatedEvent) error {
	_, err := dao.db.ExecContext(ctx,
		"INSERT INTO wallet_projections (wallet_id, holder_id, balance_in_cents, created_at, updated_at) VALUES (?, ?, 0, ?, ?)",
		event.WalletID,
		event.HolderID,
		event.CreatedAt,
		event.UpdatedAt,
	)
	if err != nil {
		slog.Error("failed to insert wallet projection", slog.String("wallet_id", event.WalletID), slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (dao *WalletMySQLProjectionDAO) ApplyFundsTransferred(ctx context.Context, event pkg.FundsTransferredEvent) error {
	_, err := dao.db.ExecContext(ctx,
		"UPDATE wallet_projections SET balance_in_cents = balance_in_cents - ?, updated_at = ? WHERE wallet_id = ?",
		event.AmountInCents,
		event.Timestamp,
		event.FromWalletID,
	)
	if err != nil {
		slog.Error(
			"failed to debit wallet projection",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.FromWalletID),
			slog.Int("amount", event.AmountInCents),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (dao *WalletMySQLProjectionDAO) ApplyFundsTransferReceived(ctx context.Context, event pkg.FundsTransferReceivedEvent) error {
	_, err := dao.db.ExecContext(ctx,
		"UPDATE wallet_projections SET balance_in_cents = balance_in_cents + ?, updated_at = ? WHERE wallet_id = ?",
		event.AmountInCents,
		event.Timestamp,
		event.WalletID,
	)
	if err != nil {
		slog.Error(
			"failed to credit wallet projection",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.WalletID),
			slog.Int("amount", event.AmountInCents),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}
