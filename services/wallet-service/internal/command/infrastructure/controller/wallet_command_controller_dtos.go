package controller

import (
	"time"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg"
	valueObject "wallet/wallet-service/pkg/domain/value_object"
)

type CreateWalletDTO struct {
	WalletID string `json:"wallet_id"`
	HolderID string `json:"holder_id"`
}

func (dto CreateWalletDTO) CreateWalletCommand() (domain.CreateWalletCommand, error) {
	walletID, err := valueObject.NewID(dto.WalletID)
	if err != nil {
		return domain.CreateWalletCommand{}, pkg.ErrInvalidWalletID
	}
	holderID, err := valueObject.NewID(dto.HolderID)
	if err != nil {
		return domain.CreateWalletCommand{}, pkg.ErrInvalidHolderID
	}
	return domain.CreateWalletCommand{
		WalletID:  walletID,
		HolderID:  holderID,
		Timestamp: time.Now(),
	}, nil
}

type TransferFundsDTO struct {
	AmountInCents int    `json:"amount_in_cents"`
	TransferID    string `json:"transfer_id"`
	ToWalletID    string `json:"to_wallet_id"`
}

func (dto TransferFundsDTO) TransferFundsCommand() (domain.TransferFundsCommand, error) {
	transferID, err := valueObject.NewID(dto.TransferID)
	if err != nil {
		return domain.TransferFundsCommand{}, pkg.ErrInvalidTransferID
	}
	toWalletID, err := valueObject.NewID(dto.ToWalletID)
	if err != nil {
		return domain.TransferFundsCommand{}, pkg.ErrInvalidToWalletID
	}
	amount, err := valueObject.NewMoney(dto.AmountInCents)
	if err != nil {
		return domain.TransferFundsCommand{}, err
	}
	if amount.Amount() == 0 {
		return domain.TransferFundsCommand{}, pkg.ErrNonPositiveAmount
	}
	return domain.TransferFundsCommand{
		Amount:     amount,
		TransferID: transferID,
		ToWalletID: toWalletID,
		Timestamp:  time.Now(),
	}, nil
}

type MockTransferDTO struct {
	AmountInCents int    `json:"amount_in_cents"`
	TransferID    string `json:"transfer_id"`
	FromWalletID  string `json:"from_wallet_id"`
}

func (dto MockTransferDTO) ReceiveFundsTransferCommand() (domain.ReceiveFundsTransferCommand, error) {
	transferID, err := valueObject.NewID(dto.TransferID)
	if err != nil {
		return domain.ReceiveFundsTransferCommand{}, pkg.ErrInvalidTransferID
	}
	fromWalletID, err := valueObject.NewID(dto.FromWalletID)
	if err != nil {
		return domain.ReceiveFundsTransferCommand{}, pkg.ErrInvalidFromWalletID
	}
	amount, err := valueObject.NewMoney(dto.AmountInCents)
	if err != nil {
		return domain.ReceiveFundsTransferCommand{}, err
	}
	if amount.Amount() == 0 {
		return domain.ReceiveFundsTransferCommand{}, pkg.ErrNonPositiveAmount
	}
	return domain.ReceiveFundsTransferCommand{
		Amount:       amount,
		TransferID:   transferID,
		FromWalletID: fromWalletID,
		Timestamp:    time.Now(),
	}, nil
}
