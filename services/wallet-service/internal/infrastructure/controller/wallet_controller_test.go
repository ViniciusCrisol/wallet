package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wallet/wallet-service/config"
	"wallet/wallet-service/internal/domain"
	"wallet/wallet-service/internal/infrastructure/persistence"
	"wallet/wallet-service/pkg/valueobject"

	"github.com/kurrent-io/KurrentDB-Client-Go/kurrentdb"
	"github.com/stretchr/testify/assert"
)

func newTestController(t *testing.T) *WalletController {
	t.Helper()

	settings, err := kurrentdb.ParseConnectionString(config.Load().KurrentDBConnectionString)
	if err != nil {
		t.Fatalf("failed to parse connection string: %v", err)
	}
	db, err := kurrentdb.NewClient(settings)
	if err != nil {
		t.Fatalf("failed to create kurrentdb client: %v", err)
	}
	return &WalletController{
		walletKurrentDBDAO: persistence.NewWalletKurrentDBDAO(db),
	}
}

func newCreateMux(ctrl *WalletController) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets", ctrl.Create)
	return mux
}

func newTransferMux(ctrl *WalletController) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /wallets/{id}/transfer", ctrl.TransferFunds)
	return mux
}

func marshalBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()

	j, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}
	return bytes.NewBuffer(j)
}

func TestWalletController_Create(t *testing.T) {
	t.Parallel()

	t.Run("It should return 201 when valid wallet data is provided", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		body := marshalBody(t, CreateWalletDTO{
			WalletID: valueobject.GenerateID().ToString(),
			HolderID: valueobject.GenerateID().ToString(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets", body)
		rec := httptest.NewRecorder()
		newCreateMux(ctrl).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("It should return 422 when request body is invalid JSON", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		req := httptest.NewRequest(http.MethodPost, "/wallets", bytes.NewBufferString("invalid"))
		rec := httptest.NewRecorder()
		newCreateMux(ctrl).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("It should return 400 when wallet_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		body := marshalBody(t, CreateWalletDTO{
			WalletID: "not-a-uuid",
			HolderID: valueobject.GenerateID().ToString(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets", body)
		rec := httptest.NewRecorder()
		newCreateMux(ctrl).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when holder_id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		body := marshalBody(t, CreateWalletDTO{
			WalletID: valueobject.GenerateID().ToString(),
			HolderID: "not-a-uuid",
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets", body)
		rec := httptest.NewRecorder()
		newCreateMux(ctrl).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 409 when a wallet with the same ID already exists", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		dto := CreateWalletDTO{
			WalletID: valueobject.GenerateID().ToString(),
			HolderID: valueobject.GenerateID().ToString(),
		}

		req1 := httptest.NewRequest(http.MethodPost, "/wallets", marshalBody(t, dto))
		rec1 := httptest.NewRecorder()
		newCreateMux(ctrl).ServeHTTP(rec1, req1)
		assert.Equal(t, http.StatusCreated, rec1.Code)

		req2 := httptest.NewRequest(http.MethodPost, "/wallets", marshalBody(t, dto))
		rec2 := httptest.NewRecorder()
		newCreateMux(ctrl).ServeHTTP(rec2, req2)
		assert.Equal(t, http.StatusConflict, rec2.Code)
	})
}

func TestWalletController_TransferFunds(t *testing.T) {
	t.Parallel()

	t.Run("It should return 204 when wallet has sufficient balance for the transfer", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		wallet, err := domain.NewWallet(domain.CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		assert.NoError(t, err)

		amount, err := valueobject.NewMoney(1000)
		assert.NoError(t, err)

		err = wallet.ReceiveFundsTransfer(domain.ReceiveFundsTransferCommand{
			Amount:       amount,
			TransferID:   valueobject.GenerateID(),
			FromWalletID: valueobject.GenerateID(),
			Timestamp:    time.Now(),
		})
		assert.NoError(t, err)
		assert.NoError(t, ctrl.walletKurrentDBDAO.Save(wallet))

		body := marshalBody(t, TransferFundsDTO{
			AmountInCents: 500,
			TransferID:    valueobject.GenerateID().ToString(),
			ToWalletID:    valueobject.GenerateID().ToString(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.GetID().ToString()+"/transfer", body)
		rec := httptest.NewRecorder()
		newTransferMux(ctrl).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("It should return 422 when request body is invalid JSON", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		nonExistentID := valueobject.GenerateID().ToString()

		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/transfer", bytes.NewBufferString("invalid"))
		rec := httptest.NewRecorder()
		newTransferMux(ctrl).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("It should return 400 when path wallet id is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		body := marshalBody(t, TransferFundsDTO{
			AmountInCents: 100,
			TransferID:    valueobject.GenerateID().ToString(),
			ToWalletID:    valueobject.GenerateID().ToString(),
		})

		req := httptest.NewRequest(http.MethodPost, "/wallets/not-a-uuid/transfer", body)
		rec := httptest.NewRecorder()
		newTransferMux(ctrl).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when transfer amount is invalid", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		nonExistentID := valueobject.GenerateID().ToString()
		body := marshalBody(t, TransferFundsDTO{
			AmountInCents: 0,
			TransferID:    valueobject.GenerateID().ToString(),
			ToWalletID:    valueobject.GenerateID().ToString(),
		})

		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/transfer", body)
		rec := httptest.NewRecorder()
		newTransferMux(ctrl).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 404 when wallet does not exist", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		nonExistentID := valueobject.GenerateID().ToString()
		body := marshalBody(t, TransferFundsDTO{
			AmountInCents: 100,
			TransferID:    valueobject.GenerateID().ToString(),
			ToWalletID:    valueobject.GenerateID().ToString(),
		})

		req := httptest.NewRequest(http.MethodPost, "/wallets/"+nonExistentID+"/transfer", body)
		rec := httptest.NewRecorder()
		newTransferMux(ctrl).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("It should return 400 when wallet has insufficient balance", func(t *testing.T) {
		t.Parallel()

		ctrl := newTestController(t)
		wallet, err := domain.NewWallet(domain.CreateWalletCommand{
			WalletID:  valueobject.GenerateID(),
			HolderID:  valueobject.GenerateID(),
			Timestamp: time.Now(),
		})
		assert.NoError(t, err)
		assert.NoError(t, ctrl.walletKurrentDBDAO.Save(wallet))

		body := marshalBody(t, TransferFundsDTO{
			AmountInCents: 500,
			TransferID:    valueobject.GenerateID().ToString(),
			ToWalletID:    valueobject.GenerateID().ToString(),
		})
		req := httptest.NewRequest(http.MethodPost, "/wallets/"+wallet.GetID().ToString()+"/transfer", body)
		rec := httptest.NewRecorder()
		newTransferMux(ctrl).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
