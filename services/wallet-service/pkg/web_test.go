package pkg

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRespondWithError(t *testing.T) {
	t.Run("It should set Content-Type to application/json", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(rec, errors.New("any error"))
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	})

	t.Run("It should include the error message in the response body", func(t *testing.T) {
		rec := httptest.NewRecorder()
		var body map[string]string
		RespondWithError(rec, ErrNegativeOrZeroAmount)
		json.Unmarshal(rec.Body.Bytes(), &body)
		assert.Equal(t, ErrNegativeOrZeroAmount.Error(), body["error"])
	})

	t.Run("It should return 404 when error wraps ErrNotFound", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(rec, ErrWalletNotFound)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("It should return 409 when error is ErrConflict", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(rec, ErrConflict)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("It should return 400 when error wraps ErrValidation", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(rec, ErrNegativeOrZeroAmount)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 422 when error is ErrUnprocessableEntity", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(rec, ErrUnprocessableEntity)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("It should return 500 when error does not match any known type", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(rec, errors.New("unexpected failure"))
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
