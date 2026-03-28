package persistence

import (
	"database/sql"
	"log/slog"

	"wallet/wallet-service/pkg"
)

const projectionName = "wallet_projection"

type WalletMysqlDAO struct {
	db *sql.DB
}

func NewWalletMysqlDAO(db *sql.DB) *WalletMysqlDAO {
	return &WalletMysqlDAO{
		db: db,
	}
}

func (dao *WalletMysqlDAO) CreateWallet(event pkg.WalletCreatedEvent) error {
	_, err := dao.db.Exec(
		"INSERT INTO wallet_projections (wallet_id, holder_id, balance_in_cents, created_at, updated_at) VALUES (?, ?, 0, ?, ?)",
		event.WalletID,
		event.HolderID,
		event.CreatedAt,
		event.UpdatedAt,
	)
	if err != nil {
		slog.Error("failed to insert wallet", slog.String("wallet_id", event.WalletID), slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (dao *WalletMysqlDAO) ApplyFundsTransferred(event pkg.FundsTransferredEvent) error {
	_, err := dao.db.Exec(
		"UPDATE wallet_projections SET balance_in_cents = balance_in_cents - ?, updated_at = ? WHERE wallet_id = ?",
		event.AmountInCents,
		event.Timestamp,
		event.FromWalletID,
	)
	if err != nil {
		slog.Error(
			"failed to update balance",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.ToWalletID),
			slog.Int("amount", event.AmountInCents),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}

func (dao *WalletMysqlDAO) ApplyFundsTransferReceived(event pkg.FundsTransferReceivedEvent) error {
	_, err := dao.db.Exec(
		"UPDATE wallet_projections SET balance_in_cents = balance_in_cents + ?, updated_at = ? WHERE wallet_id = ?",
		event.AmountInCents,
		event.Timestamp,
		event.WalletID,
	)
	if err != nil {
		slog.Error(
			"failed to update balance",
			slog.String("transfer_id", event.TransferID),
			slog.String("wallet_id", event.WalletID),
			slog.Int("amount", event.AmountInCents),
			slog.String("error", err.Error()),
		)
		return err
	}
	return nil
}
