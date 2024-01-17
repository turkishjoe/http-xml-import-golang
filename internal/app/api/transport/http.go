package transport

import (
	"context"
	"encoding/json"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/turkishjoe/xml-parser/internal/app/api/endpoints"
	"net/http"
)

func NewHTTPHandler(ep endpoints.Set) http.Handler {
	m := http.NewServeMux()

	m.Handle("/update", httptransport.NewServer(
		ep.UpdateEndpoint,
		decodeHTTPServiceUpdateRequest,
		encodeResponse,
	))
	/*m.Handle("/state", httptransport.NewServer(
		ep.UpdateEndpoint,
		decodeHTTPUpdateRequest,
		encodeResponse,
	))
	m.Handle("/get_names", httptransport.NewServer(
		ep.AddEndpoint,
		decodeHTTPAddRequest,
		encodeResponse,
	))*/

	return m
}

func decodeHTTPServiceUpdateRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	var req endpoints.UpdateRequest
	return req, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(error); ok && e != nil {
		encodeError(ctx, e, w)
		return nil
	}
	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
