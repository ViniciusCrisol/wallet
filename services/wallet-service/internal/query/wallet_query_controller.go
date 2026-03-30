package query

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"wallet/wallet-service/pkg/apperr"
	"wallet/wallet-service/pkg/valueobject"
	"wallet/wallet-service/pkg/web"
)

type WalletResponse struct {
	WalletID       string    `json:"wallet_id"`
	HolderID       string    `json:"holder_id"`
	BalanceInCents int       `json:"balance_in_cents"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	walletID, err := valueobject.NewID(request.PathValue("id"))
	if err != nil {
		web.RespondWithError(response, err)
		return
	}

	var wallet WalletResponse
	err = controller.db.QueryRowContext(
		request.Context(),
		"SELECT wallet_id, holder_id, balance_in_cents, created_at, updated_at FROM wallet_projections WHERE wallet_id = ?",
		walletID.String(),
	).Scan(
		&wallet.WalletID,
		&wallet.HolderID,
		&wallet.BalanceInCents,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		web.RespondWithError(response, apperr.ErrWalletNotFound)
		return
	}
	if err != nil {
		slog.Error("failed to query wallet projection",
			slog.String("wallet_id", walletID.String()),
			slog.String("error", err.Error()))
		web.RespondWithError(response, err)
		return
	}

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(wallet); err != nil {
		slog.Error("failed to encode wallet response",
			slog.String("wallet_id", walletID.String()),
			slog.String("error", err.Error()))
	}
}

func (controller *WalletQueryController) FindByHolderID(response http.ResponseWriter, request *http.Request) {
	holderID, err := valueobject.NewID(request.URL.Query().Get("holder_id"))
	if err != nil {
		web.RespondWithError(response, err)
		return
	}

	rows, err := controller.db.QueryContext(
		request.Context(),
		"SELECT wallet_id, holder_id, balance_in_cents, created_at, updated_at FROM wallet_projections WHERE holder_id = ?",
		holderID.String(),
	)
	if err != nil {
		slog.Error("failed to query wallet projections by holder",
			slog.String("holder_id", holderID.String()),
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

	response.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(response).Encode(wallets); err != nil {
		slog.Error("failed to encode wallets response",
			slog.String("holder_id", holderID.String()),
			slog.String("error", err.Error()))
	}
}
