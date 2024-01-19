package transport

import (
	"context"
	"encoding/json"
	"errors"
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
	))*/
	m.Handle("/get_names", httptransport.NewServer(
		ep.GetNamesEndpoint,
		decodeHTTPGetNameRequest,
		encodeResponse,
	))

	return m
}

func decodeHTTPServiceUpdateRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	var req endpoints.UpdateRequest
	return req, nil
}

func decodeHTTPGetNameRequest(_ context.Context, httpReq *http.Request) (interface{}, error) {
	var req endpoints.GetNameRequest

	queryParams := httpReq.URL.Query()

	req.Name = queryParams.Get("name")

	if len(req.Name) == 0 {
		return nil, errors.New("Name parameter does not pass")
	}

	req.IndividualSearchType = queryParams.Get("type")

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
