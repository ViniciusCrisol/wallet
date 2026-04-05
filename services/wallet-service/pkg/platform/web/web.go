package web

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	appErr "wallet/wallet-service/pkg/app_err"
)

func GetIntQueryParam(request *http.Request, key string) (int, bool) {
	param := request.URL.Query().Get(key)
	if param == "" {
		return 0, false
	}
	parsed, err := strconv.Atoi(param)
	if err != nil {
		return 0, false
	}
	return parsed, true
}

func RespondWithError(response http.ResponseWriter, err error) {
	var status int
	switch {
	case errors.Is(err, appErr.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, appErr.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, appErr.ErrValidation):
		status = http.StatusBadRequest
	case errors.Is(err, appErr.ErrUnprocessableEntity):
		status = http.StatusUnprocessableEntity
	default:
		status = http.StatusInternalServerError
		err = appErr.ErrInternal
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
