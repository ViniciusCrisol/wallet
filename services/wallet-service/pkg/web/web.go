package web

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"wallet/wallet-service/pkg/apperr"
)

func RespondWithError(response http.ResponseWriter, err error) {
	response.Header().Set("Content-Type", "application/json")

	var status int
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, apperr.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, apperr.ErrValidation):
		status = http.StatusBadRequest
	case errors.Is(err, apperr.ErrUnprocessableEntity):
		status = http.StatusUnprocessableEntity
	default:
		slog.Error("unhandled error in HTTP response", slog.String("error", err.Error()))
		status = http.StatusInternalServerError
		err = apperr.ErrInternal
	}
	response.WriteHeader(status)

	if err := json.NewEncoder(response).Encode(map[string]string{"error": err.Error()}); err != nil {
		slog.Error("failed to encode error response", slog.String("error", err.Error()))
	}
}
