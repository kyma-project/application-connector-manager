package externalapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/kyma-project/application-connector-manager/components/central-application-gateway/internal/httperrors"
	"github.com/kyma-project/application-connector-manager/components/central-application-gateway/pkg/httpconsts"
)

type ErrorHandler struct {
	Message string
	Code    int
}

func NewErrorHandler(code int, message string) *ErrorHandler {
	return &ErrorHandler{Message: message, Code: code}
}

func (eh *ErrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	responseBody := httperrors.ErrorResponse{Code: eh.Code, Error: eh.Message}

	w.Header().Set(httpconsts.HeaderContentType, httpconsts.ContentTypeApplicationJson)
	w.WriteHeader(eh.Code)
	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		slog.Warn("encode failed", "body", responseBody, "err", err.Error())
	}
}
