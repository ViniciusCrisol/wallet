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
		RespondWithError(errors.New("any error"), rec)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	})

	t.Run("It should include the error message in the response body", func(t *testing.T) {
		rec := httptest.NewRecorder()
		var body map[string]string
		RespondWithError(ErrInvalidAmount, rec)
		json.Unmarshal(rec.Body.Bytes(), &body)
		assert.Equal(t, ErrInvalidAmount.Error(), body["error"])
	})

	t.Run("It should return 404 when error wraps ErrNotFound", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(ErrWalletNotFound, rec)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("It should return 409 when error is ErrConflict", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(ErrConflict, rec)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("It should return 400 when error wraps ErrValidation", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(ErrInvalidAmount, rec)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 422 when error is ErrUnprocessableEntity", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(ErrUnprocessableEntity, rec)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("It should return 500 when error does not match any known type", func(t *testing.T) {
		rec := httptest.NewRecorder()
		RespondWithError(errors.New("unexpected failure"), rec)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
