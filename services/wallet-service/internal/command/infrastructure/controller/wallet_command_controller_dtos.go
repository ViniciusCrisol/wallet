package controller

import (
	"time"

	"wallet/wallet-service/internal/command/domain"
	appErr "wallet/wallet-service/pkg/app_err"
	valueObject "wallet/wallet-service/pkg/domain/value_object"
)

type CreateWalletDTO struct {
	WalletID  string `json:"wallet_id"`
	HolderID  string `json:"holder_id"`
	CreatedAt string `json:"created_at"`
}

func (dto CreateWalletDTO) CreateWalletCommand() (domain.CreateWalletCommand, error) {
	walletID, err := valueObject.NewID(dto.WalletID)
	if err != nil {
		return domain.CreateWalletCommand{}, appErr.ErrInvalidWalletID
	}
	holderID, err := valueObject.NewID(dto.HolderID)
	if err != nil {
		return domain.CreateWalletCommand{}, appErr.ErrInvalidHolderID
	}
	createdAt, err := time.Parse(time.RFC3339, dto.CreatedAt)
	if err != nil {
		return domain.CreateWalletCommand{}, appErr.ErrInvalidCreatedAt
	}
	return domain.CreateWalletCommand{
		WalletID:  walletID,
		HolderID:  holderID,
		Timestamp: createdAt,
	}, nil
}

type TransferFundsDTO struct {
	AmountInCents int    `json:"amount_in_cents"`
	TransferID    string `json:"transfer_id"`
	ToWalletID    string `json:"to_wallet_id"`
	Category      string `json:"category"`
	TransferredAt string `json:"transferred_at"`
}

func (dto TransferFundsDTO) TransferFundsCommand() (domain.TransferFundsCommand, error) {
	transferID, err := valueObject.NewID(dto.TransferID)
	if err != nil {
		return domain.TransferFundsCommand{}, appErr.ErrInvalidTransferID
	}
	toWalletID, err := valueObject.NewID(dto.ToWalletID)
	if err != nil {
		return domain.TransferFundsCommand{}, appErr.ErrInvalidToWalletID
	}
	amount, err := valueObject.NewMoney(dto.AmountInCents)
	if err != nil {
		return domain.TransferFundsCommand{}, err
	}
	if amount.Amount() == 0 {
		return domain.TransferFundsCommand{}, appErr.ErrNonPositiveAmount
	}
	category := dto.Category
	if category == "" {
		category = domain.CategoryUnclassified
	}
	if !domain.IsValidCategory(category) {
		return domain.TransferFundsCommand{}, appErr.ErrInvalidCategory
	}
	transferredAt, err := time.Parse(time.RFC3339, dto.TransferredAt)
	if err != nil {
		return domain.TransferFundsCommand{}, appErr.ErrInvalidTransferredAt
	}
	return domain.TransferFundsCommand{
		Amount:     amount,
		TransferID: transferID,
		ToWalletID: toWalletID,
		Category:   category,
		Timestamp:  transferredAt,
	}, nil
}

type MockTransferDTO struct {
	AmountInCents int    `json:"amount_in_cents"`
	TransferID    string `json:"transfer_id"`
	FromWalletID  string `json:"from_wallet_id"`
	Category      string `json:"category"`
	TransferredAt string `json:"transferred_at"`
}

func (dto MockTransferDTO) ReceiveFundsTransferCommand() (domain.ReceiveFundsTransferCommand, error) {
	transferID, err := valueObject.NewID(dto.TransferID)
	if err != nil {
		return domain.ReceiveFundsTransferCommand{}, appErr.ErrInvalidTransferID
	}
	fromWalletID, err := valueObject.NewID(dto.FromWalletID)
	if err != nil {
		return domain.ReceiveFundsTransferCommand{}, appErr.ErrInvalidFromWalletID
	}
	amount, err := valueObject.NewMoney(dto.AmountInCents)
	if err != nil {
		return domain.ReceiveFundsTransferCommand{}, err
	}
	if amount.Amount() == 0 {
		return domain.ReceiveFundsTransferCommand{}, appErr.ErrNonPositiveAmount
	}
	category := dto.Category
	if category == "" {
		category = domain.CategoryUnclassified
	}
	if !domain.IsValidCategory(category) {
		return domain.ReceiveFundsTransferCommand{}, appErr.ErrInvalidCategory
	}
	transferredAt, err := time.Parse(time.RFC3339, dto.TransferredAt)
	if err != nil {
		return domain.ReceiveFundsTransferCommand{}, appErr.ErrInvalidTransferredAt
	}
	return domain.ReceiveFundsTransferCommand{
		Amount:       amount,
		TransferID:   transferID,
		FromWalletID: fromWalletID,
		Category:     category,
		Timestamp:    transferredAt,
	}, nil
}
