package query

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	appErr "wallet/wallet-service/pkg/app_err"
	"wallet/wallet-service/pkg/platform/uuid"
	"wallet/wallet-service/pkg/platform/web"
)

type WalletResponse struct {
	WalletID       string    `json:"wallet_id"`
	HolderID       string    `json:"holder_id"`
	BalanceInCents int       `json:"balance_in_cents"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TransferResponse struct {
	TransferID    string    `json:"transfer_id"`
	Direction     string    `json:"direction"`
	Category      string    `json:"category"`
	AmountInCents int       `json:"amount_in_cents"`
	TransferredAt time.Time `json:"transferred_at"`
}

type WalletQueryController struct {
	db *sql.DB
}

func NewWalletQueryController(db *sql.DB) *WalletQueryController {
	return &WalletQueryController{
		db: db,
	}
}

func (controller *WalletQueryController) FindByID(response http.ResponseWriter, request *http.Request) {
	walletID := request.PathValue("id")
	if !uuid.IsValid(walletID) {
		web.RespondWithError(response, appErr.ErrInvalidWalletID)
		return
	}

	var wallet WalletResponse
	err := controller.db.QueryRowContext(
		request.Context(),
		"SELECT wallet_id, holder_id, balance_in_cents, created_at, updated_at FROM wallet_projections WHERE wallet_id = ?",
		walletID,
	).Scan(
		&wallet.WalletID,
		&wallet.HolderID,
		&wallet.BalanceInCents,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		web.RespondWithError(response, appErr.ErrWalletNotFound)
		return
	}
	if err != nil {
		slog.Error("failed to query wallet projection",
			slog.String("wallet_id", walletID),
			slog.String("error", err.Error()))
		web.RespondWithError(response, err)
		return
	}
	web.RespondWithJSON(response, http.StatusOK, wallet)
}

func (controller *WalletQueryController) FindByHolderID(response http.ResponseWriter, request *http.Request) {
	holderID := request.URL.Query().Get("holder_id")
	if !uuid.IsValid(holderID) {
		web.RespondWithError(response, appErr.ErrInvalidHolderID)
		return
	}

	rows, err := controller.db.QueryContext(
		request.Context(),
		"SELECT wallet_id, holder_id, balance_in_cents, created_at, updated_at FROM wallet_projections WHERE holder_id = ?",
		holderID,
	)
	if err != nil {
		slog.Error("failed to query wallet projections by holder",
			slog.String("holder_id", holderID),
			slog.String("error", err.Error()))
		web.RespondWithError(response, err)
		return
	}
	defer rows.Close()

	wallets := []WalletResponse{}
	for rows.Next() {
		var wallet WalletResponse
		if err := rows.Scan(
			&wallet.WalletID,
			&wallet.HolderID,
			&wallet.BalanceInCents,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			slog.Error("failed to scan wallet projection row", slog.String("error", err.Error()))
			web.RespondWithError(response, err)
			return
		}
		wallets = append(wallets, wallet)
	}
	if err := rows.Err(); err != nil {
		slog.Error("error iterating wallet projection rows", slog.String("error", err.Error()))
		web.RespondWithError(response, err)
		return
	}
	web.RespondWithJSON(response, http.StatusOK, wallets)
}

func (controller *WalletQueryController) FindTransfersByWalletID(response http.ResponseWriter, request *http.Request) {
	walletID := request.PathValue("id")
	if !uuid.IsValid(walletID) {
		web.RespondWithError(response, appErr.ErrInvalidWalletID)
		return
	}
	limit, ok := web.GetIntQueryParam(request, "limit")
	if !ok || limit <= 0 || limit > 100 {
		web.RespondWithError(response, appErr.ErrInvalidPaginationLimit)
		return
	}
	offset, ok := web.GetIntQueryParam(request, "offset")
	if !ok || offset < 0 {
		web.RespondWithError(response, appErr.ErrInvalidPaginationOffset)
		return
	}

	rows, err := controller.db.QueryContext(
		request.Context(),
		`
			SELECT
				transfer_id, direction, category, amount_in_cents, transferred_at
			FROM
				transfer_projections
			WHERE
				wallet_id = ?
			ORDER BY
				transferred_at DESC
			LIMIT ?
			OFFSET ?
		`,
		walletID, limit, offset,
	)
	if err != nil {
		slog.Error("failed to query transfer projections",
			slog.String("wallet_id", walletID),
			slog.String("error", err.Error()))
		web.RespondWithError(response, err)
		return
	}
	defer rows.Close()

	transfers := []TransferResponse{}
	for rows.Next() {
		var transfer TransferResponse
		if err := rows.Scan(
			&transfer.TransferID,
			&transfer.Direction,
			&transfer.Category,
			&transfer.AmountInCents,
			&transfer.TransferredAt,
		); err != nil {
			slog.Error("failed to scan transfer projection row", slog.String("error", err.Error()))
			web.RespondWithError(response, err)
			return
		}
		transfers = append(transfers, transfer)
	}
	if err := rows.Err(); err != nil {
		slog.Error("error iterating transfer projection rows", slog.String("error", err.Error()))
		web.RespondWithError(response, err)
		return
	}
	web.RespondWithJSON(response, http.StatusOK, transfers)
}
