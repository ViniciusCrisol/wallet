package controller

import (
	"encoding/json"
	"net/http"

	"wallet/wallet-service/internal/domain"
	"wallet/wallet-service/internal/infrastructure/persistence"
	"wallet/wallet-service/pkg"
	"wallet/wallet-service/pkg/valueobject"
)

type WalletController struct {
	wallletKurrentdbDAO *persistence.WalletKurrentdbDAO
}

func (controller *WalletController) Create(response http.ResponseWriter, request *http.Request) {
	var dto CreateWalletDTO
	if err := json.NewDecoder(request.Body).Decode(&dto); err != nil {
		pkg.RespondWithError(pkg.ErrUnprocessableEntity, response)
		return
	}
	command, err := dto.ToCreateWalletCommand()
	if err != nil {
		pkg.RespondWithError(err, response)
		return
	}

	wallet, err := domain.NewWallet(command)
	if err != nil {
		pkg.RespondWithError(err, response)
		return
	}
	if err := controller.wallletKurrentdbDAO.Save(wallet); err != nil {
		pkg.RespondWithError(err, response)
		return
	}

	response.WriteHeader(http.StatusCreated)
}

func (controller *WalletController) TransferFunds(response http.ResponseWriter, request *http.Request) {
	var dto TransferFundsDTO
	if err := json.NewDecoder(request.Body).Decode(&dto); err != nil {
		pkg.RespondWithError(pkg.ErrUnprocessableEntity, response)
		return
	}
	walletID, err := valueobject.NewID(request.PathValue("id"))
	if err != nil {
		pkg.RespondWithError(err, response)
		return
	}
	command, err := dto.ToTransferFundsCommand()
	if err != nil {
		pkg.RespondWithError(err, response)
		return
	}

	wallet, found, err := controller.wallletKurrentdbDAO.Find(walletID)
	if err != nil {
		pkg.RespondWithError(err, response)
		return
	}
	if !found {
		pkg.RespondWithError(pkg.ErrWalletNotFound, response)
		return
	}
	if err := wallet.TransferFunds(command); err != nil {
		pkg.RespondWithError(err, response)
		return
	}
	if err := controller.wallletKurrentdbDAO.Save(wallet); err != nil {
		pkg.RespondWithError(err, response)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
