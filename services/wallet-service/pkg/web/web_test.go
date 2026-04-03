package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"wallet/wallet-service/pkg/apperr"

	"github.com/stretchr/testify/assert"
)

func TestRespondWithError(t *testing.T) {
	t.Parallel()

	t.Run("It should set Content-Type to application/json", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		RespondWithError(rec, errors.New("any error"))
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	})

	t.Run("It should include the error message in the response body", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		var body map[string]string
		RespondWithError(rec, apperr.ErrNegativeAmount)
		json.Unmarshal(rec.Body.Bytes(), &body)
		assert.Equal(t, apperr.ErrNegativeAmount.Error(), body["error"])
	})

	t.Run("It should mask the original error message for 500 responses", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		RespondWithError(rec, errors.New("database connection refused"))
		var body map[string]string
		json.Unmarshal(rec.Body.Bytes(), &body)
		assert.Equal(t, apperr.ErrInternal.Error(), body["error"])
		assert.NotContains(t, body["error"], "database connection refused")
	})

	t.Run("It should return 404 when error wraps ErrNotFound", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		RespondWithError(rec, apperr.ErrWalletNotFound)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("It should return 409 when error is ErrConflict", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		RespondWithError(rec, apperr.ErrConflict)
		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("It should return 400 when error wraps ErrValidation", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		RespondWithError(rec, apperr.ErrNegativeAmount)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("It should return 422 when error is ErrUnprocessableEntity", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		RespondWithError(rec, apperr.ErrUnprocessableEntity)
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	})

	t.Run("It should return 500 when error does not match any known type", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		RespondWithError(rec, errors.New("unexpected failure"))
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestRespondWithJSON(t *testing.T) {
	t.Parallel()

	t.Run("It should set Content-Type to application/json", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		RespondWithJSON(rec, http.StatusOK, map[string]string{"key": "value"})
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	})

	t.Run("It should set the provided status code", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		RespondWithJSON(rec, http.StatusCreated, map[string]string{"key": "value"})
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("It should encode the data as JSON in the response body", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		data := map[string]string{"name": "test"}
		RespondWithJSON(rec, http.StatusOK, data)

		var body map[string]string
		json.Unmarshal(rec.Body.Bytes(), &body)
		assert.Equal(t, "test", body["name"])
	})
}
