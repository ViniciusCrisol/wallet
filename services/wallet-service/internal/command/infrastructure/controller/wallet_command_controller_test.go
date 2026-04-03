package controller

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wallet/wallet-service/internal/command/domain"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/stretchr/testify/assert"
)

func TestWalletCommandController_Create(t *testing.T) {
	t.Parallel()

	t.Run("It should return 201 when valid wallet data is provided", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		body := marshalBody(t, CreateWalletDTO{
			WalletID: valueobject.GenerateID().String(),
			HolderID: valueobject.GenerateID().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets", body)
		rec := httptest.NewRecorder()
		newCreateMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("It should return 422 when request body is invalid JSON", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString("invalid"))
		rec := httptest.NewRecorder()
		newCreateMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("It should return 400 when wallet_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		body := marshalBody(t, CreateWalletDTO{
			WalletID: "not-a-uuid",
			HolderID: valueobject.GenerateID().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets", body)
		rec := httptest.NewRecorder()
		newCreateMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when holder_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		body := marshalBody(t, CreateWalletDTO{
			WalletID: valueobject.GenerateID().String(),
			HolderID: "not-a-uuid",
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets", body)
		rec := httptest.NewRecorder()
		newCreateMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 409 when a wallet with the same ID already exists", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		dto := CreateWalletDTO{
			WalletID: valueobject.GenerateID().String(),
			HolderID: valueobject.GenerateID().String(),
		}

		req1 := httptest.NewRequest(http.MethodPost, "/wallets", marshalBody(t, dto))
		rec1 := httptest.NewRecorder()
		newCreateMux(controller).ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusCreated, rec1.Code)

		req2 := httptest.NewRequest(http.MethodPost, "/wallets", marshalBody(t, dto))
		rec2 := httptest.NewRecorder()
		newCreateMux(controller).ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusConflict, rec2.Code)
	})
}

func TestWalletCommandController_TransferFunds(t *testing.T) {
	t.Parallel()

	t.Run("It should return 204 when wallet has sufficient balance for the transfer", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		wallet := domain.NewWallet(domain.CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})

		amount, err := valueobject.NewMoney(1000)
		assert.NoError(t, err)

		err = wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})
		assert.NoError(t, err)
		assert.NoError(t, controller.esHandler.Save(context.Background(), wallet))

		body := marshalBody(t, TransferFundsDTO{
			AmountInCents: 500,
			TransferID:    valueobject.GenerateID().String(),
			ToWalletID:    valueobject.GenerateID().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID().String()+"/transfer", body)
		rec := httptest.NewRecorder()
		newTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("It should return 422 when request body is invalid JSON", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		nonExistentID := valueobject.GenerateID().String()

		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/transfer", bytes.NewBufferString("invalid"))
		rec := httptest.NewRecorder()
		newTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("It should return 400 when path wallet id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		body := marshalBody(t, TransferFundsDTO{
			AmountInCents: 100,
			TransferID:    valueobject.GenerateID().String(),
			ToWalletID:    valueobject.GenerateID().String(),
		})

		req := httptest.NewRequest(http.MethodPost, "/wallets/not-a-uuid/transfer", body)
		rec := httptest.NewRecorder()
		newTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when transfer amount is invalid", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		nonExistentID := valueobject.GenerateID().String()
		body := marshalBody(t, TransferFundsDTO{
			AmountInCents: 0,
			TransferID:    valueobject.GenerateID().String(),
			ToWalletID:    valueobject.GenerateID().String(),
		})

		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/transfer", body)
		rec := httptest.NewRecorder()
		newTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 404 when wallet does not exist", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		nonExistentID := valueobject.GenerateID().String()
		body := marshalBody(t, TransferFundsDTO{
			AmountInCents: 100,
			TransferID:    valueobject.GenerateID().String(),
			ToWalletID:    valueobject.GenerateID().String(),
		})

		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/transfer", body)
		rec := httptest.NewRecorder()
		newTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("It should return 400 when wallet has insufficient balance", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		wallet := domain.NewWallet(domain.CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		assert.NoError(t, controller.esHandler.Save(context.Background(), wallet))

		body := marshalBody(t, TransferFundsDTO{
			AmountInCents: 500,
			TransferID:    valueobject.GenerateID().String(),
			ToWalletID:    valueobject.GenerateID().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID().String()+"/transfer", body)
		rec := httptest.NewRecorder()
		newTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 409 when the same transfer is submitted twice", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		wallet := domain.NewWallet(domain.CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})

		amount, err := valueobject.NewMoney(1000)
		assert.NoError(t, err)
		err = wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})
		assert.NoError(t, err)
		assert.NoError(t, controller.esHandler.Save(context.Background(), wallet))

		transferID := valueobject.GenerateID().String()
		dto := TransferFundsDTO{
			AmountInCents: 100,
			TransferID:    transferID,
			ToWalletID:    valueobject.GenerateID().String(),
		}

		req1 := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID().String()+"/transfer", marshalBody(t, dto))
		rec1 := httptest.NewRecorder()
		newTransferMux(controller).ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusNoContent, rec1.Code)

		req2 := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID().String()+"/transfer", marshalBody(t, dto))
		rec2 := httptest.NewRecorder()
		newTransferMux(controller).ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusConflict, rec2.Code)
	})
}

func TestWalletCommandController_MockTransfer(t *testing.T) {
	t.Parallel()

	t.Run("It should return 204 when wallet exists and balance limit is not exceeded", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		wallet := domain.NewWallet(domain.CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		assert.NoError(t, controller.esHandler.Save(context.Background(), wallet))

		body := marshalBody(t, MockTransferDTO{
			AmountInCents: 500,
			TransferID:    valueobject.GenerateID().String(),
			FromWalletID:  valueobject.GenerateID().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID().String()+"/mock-transfer", body)
		rec := httptest.NewRecorder()
		newMockTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("It should return 422 when request body is invalid JSON", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		nonExistentID := valueobject.GenerateID().String()
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/mock-transfer", bytes.NewBufferString("invalid"))
		rec := httptest.NewRecorder()
		newMockTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("It should return 400 when path wallet id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		body := marshalBody(t, MockTransferDTO{
			AmountInCents: 100,
			TransferID:    valueobject.GenerateID().String(),
			FromWalletID:  valueobject.GenerateID().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/not-a-uuid/mock-transfer", body)
		rec := httptest.NewRecorder()
		newMockTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when transfer_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		nonExistentID := valueobject.GenerateID().String()
		body := marshalBody(t, MockTransferDTO{
			AmountInCents: 100,
			TransferID:    "not-a-uuid",
			FromWalletID:  valueobject.GenerateID().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/mock-transfer", body)
		rec := httptest.NewRecorder()
		newMockTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when from_wallet_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		nonExistentID := valueobject.GenerateID().String()
		body := marshalBody(t, MockTransferDTO{
			AmountInCents: 100,
			TransferID:    valueobject.GenerateID().String(),
			FromWalletID:  "not-a-uuid",
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/mock-transfer", body)
		rec := httptest.NewRecorder()
		newMockTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when amount_in_cents is zero", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		nonExistentID := valueobject.GenerateID().String()
		body := marshalBody(t, MockTransferDTO{
			AmountInCents: 0,
			TransferID:    valueobject.GenerateID().String(),
			FromWalletID:  valueobject.GenerateID().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/mock-transfer", body)
		rec := httptest.NewRecorder()
		newMockTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 404 when wallet does not exist", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		nonExistentID := valueobject.GenerateID().String()
		body := marshalBody(t, MockTransferDTO{
			AmountInCents: 100,
			TransferID:    valueobject.GenerateID().String(),
			FromWalletID:  valueobject.GenerateID().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/mock-transfer", body)
		rec := httptest.NewRecorder()
		newMockTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("It should return 400 when receiving funds would exceed the balance limit", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		wallet := domain.NewWallet(domain.CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})

		nearMaxAmount, err := valueobject.NewMoney(99_999_999)
		assert.NoError(t, err)
		err = wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       nearMaxAmount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})
		assert.NoError(t, err)
		assert.NoError(t, controller.esHandler.Save(context.Background(), wallet))

		body := marshalBody(t, MockTransferDTO{
			AmountInCents: 2,
			TransferID:    valueobject.GenerateID().String(),
			FromWalletID:  valueobject.GenerateID().String(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID().String()+"/mock-transfer", body)
		rec := httptest.NewRecorder()
		newMockTransferMux(controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 409 when the same mock transfer is submitted twice", func(t *testing.T) {
		t.Parallel()

		controller := newTestController(t)
		wallet := domain.NewWallet(domain.CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		assert.NoError(t, controller.esHandler.Save(context.Background(), wallet))

		transferID := valueobject.GenerateID().String()
		dto := MockTransferDTO{
			AmountInCents: 100,
			TransferID:    transferID,
			FromWalletID:  valueobject.GenerateID().String(),
		}

		req1 := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID().String()+"/mock-transfer", marshalBody(t, dto))
		rec1 := httptest.NewRecorder()
		newMockTransferMux(controller).ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusNoContent, rec1.Code)

		req2 := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.ID().String()+"/mock-transfer", marshalBody(t, dto))
		rec2 := httptest.NewRecorder()
		newMockTransferMux(controller).ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusConflict, rec2.Code)
	})
}
