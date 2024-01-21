package endpoints

import (
	"github.com/turkishjoe/xml-parser/internal/app/api"
	"github.com/turkishjoe/xml-parser/internal/app/api/domain"
)

type UpdateRequest struct {
}

type UpdateResponse struct {
	Result bool   `json:"result"`
	Info   string `json:"info"`
	Code   int    `json:"code"`
}

type StateRequest struct{}

type StateResponse struct {
	Result bool   `json:"result"`
	Info   string `json:"info"`
}

type GetNameRequest struct {
	Name                 string
	IndividualSearchType string
}

type GetNameResponse struct {
	Individuals []domain.Individual
}

var searchTypeStringMap = map[string]domain.SearchType{
	"strong": domain.Strong,
	"weak":   domain.Weak,
	"both":   domain.Both,
}

var stateMap = map[api.State]string{
	api.Empty:    "empty",
	api.Updating: "updating",
	api.Ok:       "ok",
}
