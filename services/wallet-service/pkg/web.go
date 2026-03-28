package pkg

import (
	"encoding/json"
	"errors"
	"net/http"
)

func RespondWithError(
	err error,
	response http.ResponseWriter,
) {
	response.Header().Set("Content-Type", "application/json")

	var status int
	switch {
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, ErrValidation):
		status = http.StatusBadRequest
	case errors.Is(err, ErrUnprocessableEntity):
		status = http.StatusUnprocessableEntity
	default:
		status = http.StatusInternalServerError
	}
	response.WriteHeader(status)

	json.NewEncoder(response).Encode(map[string]string{"error": err.Error()})
}
