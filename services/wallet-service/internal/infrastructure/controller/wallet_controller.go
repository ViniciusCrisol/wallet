package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"wallet/wallet-service/internal/domain"
	"wallet/wallet-service/internal/infrastructure/persistence"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"
)

type WalletController struct {
	walletKurrentDBDAO *persistence.WalletKurrentDBDAO
}

func NewWalletController(dao *persistence.WalletKurrentDBDAO) *WalletController {
	return &WalletController{
		walletKurrentDBDAO: dao,
	}
}

func (controller *WalletController) Create(response http.ResponseWriter, request *http.Request) {
	var dto CreateWalletDTO
	if err := json.NewDecoder(request.Body).Decode(&dto); err != nil {
		slog.Warn("failed to decode create wallet request", slog.String("error", err.Error()))
		pkg.RespondWithError(response, pkg.ErrUnprocessableEntity)
		return
	}
	command, err := dto.ToCreateWalletCommand()
	if err != nil {
		pkg.RespondWithError(response, err)
		return
	}

	wallet := domain.NewWallet(command)
	if err := controller.walletKurrentDBDAO.Save(wallet); err != nil {
		pkg.RespondWithError(response, err)
		return
	}

	response.WriteHeader(http.StatusCreated)
}

func (controller *WalletController) TransferFunds(response http.ResponseWriter, request *http.Request) {
	var dto TransferFundsDTO
	if err := json.NewDecoder(request.Body).Decode(&dto); err != nil {
		slog.Warn("failed to decode transfer funds request", slog.String("error", err.Error()))
		pkg.RespondWithError(response, pkg.ErrUnprocessableEntity)
		return
	}
	walletID, err := valueobject.NewID(request.PathValue("id"))
	if err != nil {
		pkg.RespondWithError(response, err)
		return
	}
	command, err := dto.ToTransferFundsCommand()
	if err != nil {
		pkg.RespondWithError(response, err)
		return
	}

	wallet, found, err := controller.walletKurrentDBDAO.Find(walletID)
	if err != nil {
		pkg.RespondWithError(response, err)
		return
	}
	if !found {
		pkg.RespondWithError(response, pkg.ErrWalletNotFound)
		return
	}
	if err := wallet.TransferFunds(command); err != nil {
		pkg.RespondWithError(response, err)
		return
	}
	if err := controller.walletKurrentDBDAO.Save(wallet); err != nil {
		pkg.RespondWithError(response, err)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
