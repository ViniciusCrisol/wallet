package query

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"wallet/wallet-service/pkg/uuid"
	"wallet/wallet-service/pkg/valueobject"

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
		nonExistentID := valueobject.GenerateID().String()

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
