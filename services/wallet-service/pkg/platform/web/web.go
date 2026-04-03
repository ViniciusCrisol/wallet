package web

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"wallet/wallet-service/pkg"
)

func RespondWithError(response http.ResponseWriter, err error) {
	var status int
	switch {
	case errors.Is(err, pkg.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, pkg.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, pkg.ErrValidation):
		status = http.StatusBadRequest
	case errors.Is(err, pkg.ErrUnprocessableEntity):
		status = http.StatusUnprocessableEntity
	default:
		status = http.StatusInternalServerError
		err = pkg.ErrInternal
	}
	RespondWithJSON(response, status, map[string]string{"error": err.Error()})
}

func RespondWithJSON(response http.ResponseWriter, status int, data any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)

	if err := json.NewEncoder(response).Encode(data); err != nil {
		slog.Error("failed to encode response", slog.String("error", err.Error()))
	}
}
