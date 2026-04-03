package query

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	valueObject "wallet/wallet-service/pkg/domain/value_object"
	"wallet/wallet-service/pkg/platform/uuid"

	"github.com/stretchr/testify/assert"
)

func TestWalletQueryController_FindByID(t *testing.T) {
	t.Parallel()

	t.Run("It should return 200 with wallet data when wallet exists", func(t *testing.T) {
		t.Parallel()

		controller := NewWalletQueryController(db)
		walletID := uuid.NewUUID()
		holderID := uuid.NewUUID()
		createTestWallet(t, walletID, holderID)

		req := httptest.NewRequest(http.MethodGet, "/wallets/"+walletID, nil)
		rec := httptest.NewRecorder()
		newFindByIDMux(t, controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var wallet WalletResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&wallet))
		assert.Equal(t, walletID, wallet.WalletID)
		assert.Equal(t, holderID, wallet.HolderID)
		assert.Equal(t, 0, wallet.BalanceInCents)
	})

	t.Run("It should return 404 when wallet does not exist", func(t *testing.T) {
		t.Parallel()

		controller := NewWalletQueryController(db)
		nonExistentID := valueObject.GenerateID().String()

		req := httptest.NewRequest(http.MethodGet, "/wallets/"+nonExistentID, nil)
		rec := httptest.NewRecorder()
		newFindByIDMux(t, controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("It should return 400 when path ID is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		controller := NewWalletQueryController(db)

		req := httptest.NewRequest(http.MethodGet, "/wallets/not-a-uuid", nil)
		rec := httptest.NewRecorder()
		newFindByIDMux(t, controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestWalletQueryController_FindByHolderID(t *testing.T) {
	t.Parallel()

	t.Run("It should return 200 with wallets when holder has wallets", func(t *testing.T) {
		t.Parallel()

		controller := NewWalletQueryController(db)
		holderID := uuid.NewUUID()
		walletID1 := uuid.NewUUID()
		walletID2 := uuid.NewUUID()
		createTestWallet(t, walletID1, holderID)
		createTestWallet(t, walletID2, holderID)

		req := httptest.NewRequest(http.MethodGet, "/wallets?holder_id="+holderID, nil)
		rec := httptest.NewRecorder()
		newFindByHolderIDMux(t, controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var wallets []WalletResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&wallets))
		assert.Len(t, wallets, 2)

		walletIDs := []string{wallets[0].WalletID, wallets[1].WalletID}
		assert.Contains(t, walletIDs, walletID1)
		assert.Contains(t, walletIDs, walletID2)
		for _, w := range wallets {
			assert.Equal(t, holderID, w.HolderID)
			assert.Equal(t, 0, w.BalanceInCents)
		}
	})

	t.Run("It should return 200 with empty array when holder has no wallets", func(t *testing.T) {
		t.Parallel()

		controller := NewWalletQueryController(db)
		holderID := uuid.NewUUID()

		req := httptest.NewRequest(http.MethodGet, "/wallets?holder_id="+holderID, nil)
		rec := httptest.NewRecorder()
		newFindByHolderIDMux(t, controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var wallets []WalletResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&wallets))
		assert.Empty(t, wallets)
	})

	t.Run("It should return 400 when holder_id query param is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		controller := NewWalletQueryController(db)

		req := httptest.NewRequest(http.MethodGet, "/wallets?holder_id=not-a-uuid", nil)
		rec := httptest.NewRecorder()
		newFindByHolderIDMux(t, controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when holder_id query param is missing", func(t *testing.T) {
		t.Parallel()

		controller := NewWalletQueryController(db)

		req := httptest.NewRequest(http.MethodGet, "/wallets", nil)
		rec := httptest.NewRecorder()
		newFindByHolderIDMux(t, controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestWalletQueryController_FindTransfersByWalletID(t *testing.T) {
	t.Parallel()

	t.Run("It should return 200 with transfers when wallet has transfers", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		walletID := uuid.NewUUID()
		transferID1 := uuid.NewUUID()
		transferID2 := uuid.NewUUID()
		counterpartID := uuid.NewUUID()
		createTestWallet(t, walletID, uuid.NewUUID())
		createTestWallet(t, counterpartID, uuid.NewUUID())
		createTestTransfer(t, walletID, transferID1, counterpartID, "outgoing", 1000, now.Add(-time.Minute))
		createTestTransfer(t, walletID, transferID2, counterpartID, "incoming", 500, now)

		req := httptest.NewRequest(http.MethodGet, "/wallets/"+walletID+"/transfers?limit=100&offset=0", nil)
		rec := httptest.NewRecorder()
		newFindTransfersMux(t, NewWalletQueryController(db)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var transfers []TransferResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&transfers))
		assert.Len(t, transfers, 2)
		assert.Equal(t, transferID2, transfers[0].TransferID)
		assert.Equal(t, transferID1, transfers[1].TransferID)
	})

	t.Run("It should return 200 with empty array when wallet has no transfers", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/wallets/"+uuid.NewUUID()+"/transfers?limit=100&offset=0", nil)
		rec := httptest.NewRecorder()
		newFindTransfersMux(t, NewWalletQueryController(db)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var transfers []TransferResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&transfers))
		assert.Empty(t, transfers)
	})

	t.Run("It should respect limit query parameter", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		walletID := uuid.NewUUID()
		counterpartID := uuid.NewUUID()
		createTestWallet(t, walletID, uuid.NewUUID())
		createTestWallet(t, counterpartID, uuid.NewUUID())
		for i := range 5 {
			transferredAt := now.Add(time.Duration(i) * time.Second)
			createTestTransfer(t, walletID, uuid.NewUUID(), counterpartID, "outgoing", 100, transferredAt)
		}
		req := httptest.NewRequest(http.MethodGet, "/wallets/"+walletID+"/transfers?limit=2&offset=0", nil)
		rec := httptest.NewRecorder()
		newFindTransfersMux(t, NewWalletQueryController(db)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var transfers []TransferResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&transfers))
		assert.Len(t, transfers, 2)
	})

	t.Run("It should respect offset query parameter", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		walletID := uuid.NewUUID()
		counterpartID := uuid.NewUUID()
		createTestWallet(t, walletID, uuid.NewUUID())
		createTestWallet(t, counterpartID, uuid.NewUUID())
		transferIDs := make([]string, 3)
		for i := range 3 {
			transferIDs[i] = uuid.NewUUID()
			transferredAt := now.Add(time.Duration(i) * time.Second)
			createTestTransfer(t, walletID, transferIDs[i], counterpartID, "outgoing", 100, transferredAt)
		}
		req := httptest.NewRequest(http.MethodGet, "/wallets/"+walletID+"/transfers?limit=10&offset=1", nil)
		rec := httptest.NewRecorder()
		newFindTransfersMux(t, NewWalletQueryController(db)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var transfers []TransferResponse
		assert.NoError(t, json.NewDecoder(rec.Body).Decode(&transfers))
		assert.Len(t, transfers, 2)
		assert.Equal(t, transferIDs[1], transfers[0].TransferID)
		assert.Equal(t, transferIDs[0], transfers[1].TransferID)
	})

	t.Run("It should return 400 when wallet ID is not a valid UUID", func(t *testing.T) {
		t.Parallel()

		controller := NewWalletQueryController(db)

		req := httptest.NewRequest(http.MethodGet, "/wallets/not-a-uuid/transfers?limit=100&offset=0", nil)
		rec := httptest.NewRecorder()
		newFindTransfersMux(t, controller).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when limit is invalid", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/wallets/"+uuid.NewUUID()+"/transfers?limit=0&offset=0", nil)
		rec := httptest.NewRecorder()
		newFindTransfersMux(t, NewWalletQueryController(db)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when limit exceeds 100", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/wallets/"+uuid.NewUUID()+"/transfers?limit=101&offset=0", nil)
		rec := httptest.NewRecorder()
		newFindTransfersMux(t, NewWalletQueryController(db)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 400 when offset is negative", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/wallets/"+uuid.NewUUID()+"/transfers??limit=100offset=-1", nil)
		rec := httptest.NewRecorder()
		newFindTransfersMux(t, NewWalletQueryController(db)).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
