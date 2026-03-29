package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"wallet/wallet-service/internal/commandside/domain"
	"wallet/wallet-service/internal/commandside/infrastructure/persistence"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"
)

type WalletController struct {
	esHandler *persistence.WalletKurrentDBESHandler
}

func NewWalletController(esHandler *persistence.WalletKurrentDBESHandler) *WalletController {
	return &WalletController{
		esHandler: esHandler,
	}
}

func (controller *WalletController) Create(response http.ResponseWriter, request *http.Request) {
	var dto CreateWalletDTO
	if err := json.NewDecoder(request.Body).Decode(&dto); err != nil {
		slog.Warn("failed to decode create wallet request", slog.String("error", err.Error()))
		pkg.RespondWithError(response, pkg.ErrUnprocessableEntity)
		return
	}
	command, err := dto.CreateWalletCommand()
	if err != nil {
		pkg.RespondWithError(response, err)
		return
	}

	wallet := domain.NewWallet(command)
	if err := controller.esHandler.Save(request.Context(), wallet); err != nil {
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
	command, err := dto.TransferFundsCommand()
	if err != nil {
		pkg.RespondWithError(response, err)
		return
	}

	wallet, found, err := controller.esHandler.Find(request.Context(), walletID)
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
	if err := controller.esHandler.Save(request.Context(), wallet); err != nil {
		pkg.RespondWithError(response, err)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}

func (controller *WalletController) MockTransfer(response http.ResponseWriter, request *http.Request) {
	var dto MockTransferDTO
	if err := json.NewDecoder(request.Body).Decode(&dto); err != nil {
		slog.Warn("failed to decode mock transfer request", slog.String("error", err.Error()))
		pkg.RespondWithError(response, pkg.ErrUnprocessableEntity)
		return
	}
	walletID, err := valueobject.NewID(request.PathValue("id"))
	if err != nil {
		pkg.RespondWithError(response, err)
		return
	}
	command, err := dto.ReceiveFundsTransferCommand()
	if err != nil {
		pkg.RespondWithError(response, err)
		return
	}

	wallet, found, err := controller.esHandler.Find(request.Context(), walletID)
	if err != nil {
		pkg.RespondWithError(response, err)
		return
	}
	if !found {
		pkg.RespondWithError(response, pkg.ErrWalletNotFound)
		return
	}
	if err := wallet.ReceiveFundsTransfer(command); err != nil {
		pkg.RespondWithError(response, err)
		return
	}
	if err := controller.esHandler.Save(request.Context(), wallet); err != nil {
		pkg.RespondWithError(response, err)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
