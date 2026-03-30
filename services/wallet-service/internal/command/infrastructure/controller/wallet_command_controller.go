package controller

import (
	"encoding/json"
	"net/http"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/internal/command/infrastructure/persistence"
	"wallet/wallet-service/pkg/apperr"
	"wallet/wallet-service/pkg/valueobject"
	"wallet/wallet-service/pkg/web"
)

type WalletCommandController struct {
	esHandler *persistence.WalletKurrentDBESHandler
}

func NewWalletCommandController(esHandler *persistence.WalletKurrentDBESHandler) *WalletCommandController {
	return &WalletCommandController{
		esHandler: esHandler,
	}
}

const maxRequestBodySize = 1 << 20 // 1 MB

func (controller *WalletCommandController) Create(response http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(response, request.Body, maxRequestBodySize)
	var dto CreateWalletDTO
	if err := json.NewDecoder(request.Body).Decode(&dto); err != nil {
		web.RespondWithError(response, apperr.ErrUnprocessableEntity)
		return
	}
	command, err := dto.CreateWalletCommand()
	if err != nil {
		web.RespondWithError(response, err)
		return
	}

	wallet := domain.NewWallet(command)
	if err := controller.esHandler.Save(request.Context(), wallet); err != nil {
		web.RespondWithError(response, err)
		return
	}

	response.Header().Set("Location", "/wallets/"+wallet.ID().String())
	response.WriteHeader(http.StatusCreated)
}

func (controller *WalletCommandController) TransferFunds(response http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(response, request.Body, maxRequestBodySize)
	var dto TransferFundsDTO
	if err := json.NewDecoder(request.Body).Decode(&dto); err != nil {
		web.RespondWithError(response, apperr.ErrUnprocessableEntity)
		return
	}
	walletID, err := valueobject.NewID(request.PathValue("id"))
	if err != nil {
		web.RespondWithError(response, err)
		return
	}
	command, err := dto.TransferFundsCommand()
	if err != nil {
		web.RespondWithError(response, err)
		return
	}

	wallet, found, err := controller.esHandler.Find(request.Context(), walletID)
	if err != nil {
		web.RespondWithError(response, err)
		return
	}
	if !found {
		web.RespondWithError(response, apperr.ErrWalletNotFound)
		return
	}
	if err := wallet.TransferFunds(command); err != nil {
		web.RespondWithError(response, err)
		return
	}
	if err := controller.esHandler.Save(request.Context(), wallet); err != nil {
		web.RespondWithError(response, err)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}

func (controller *WalletCommandController) MockTransfer(response http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(response, request.Body, maxRequestBodySize)
	var dto MockTransferDTO
	if err := json.NewDecoder(request.Body).Decode(&dto); err != nil {
		web.RespondWithError(response, apperr.ErrUnprocessableEntity)
		return
	}
	walletID, err := valueobject.NewID(request.PathValue("id"))
	if err != nil {
		web.RespondWithError(response, err)
		return
	}
	command, err := dto.ReceiveFundsTransferCommand()
	if err != nil {
		web.RespondWithError(response, err)
		return
	}

	wallet, found, err := controller.esHandler.Find(request.Context(), walletID)
	if err != nil {
		web.RespondWithError(response, err)
		return
	}
	if !found {
		web.RespondWithError(response, apperr.ErrWalletNotFound)
		return
	}
	if err := wallet.ReceiveFundsTransfer(command); err != nil {
		web.RespondWithError(response, err)
		return
	}
	if err := controller.esHandler.Save(request.Context(), wallet); err != nil {
		web.RespondWithError(response, err)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
