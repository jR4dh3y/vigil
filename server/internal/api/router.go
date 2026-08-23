package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterOpenAPIRoutes registers generated routes with the API's JSON error contract.
func RegisterOpenAPIRoutes(server ServerInterface, router chi.Router, baseURL string) {
	HandlerWithOptions(server, ChiServerOptions{
		BaseURL:          baseURL,
		BaseRouter:       router,
		ErrorHandlerFunc: writeParameterError,
	})
}

func writeParameterError(w http.ResponseWriter, _ *http.Request, err error) {
	message := "invalid request parameters"

	var required *RequiredParamError
	var invalid *InvalidParamFormatError
	var tooMany *TooManyValuesForParamError
	switch {
	case errors.As(err, &required):
		message = fmt.Sprintf("%s is required", required.ParamName)
	case errors.As(err, &invalid):
		message = fmt.Sprintf("%s has invalid format", invalid.ParamName)
	case errors.As(err, &tooMany):
		message = fmt.Sprintf("%s must be provided once", tooMany.ParamName)
	}

	writeError(w, http.StatusBadRequest, message, "validation")
}
