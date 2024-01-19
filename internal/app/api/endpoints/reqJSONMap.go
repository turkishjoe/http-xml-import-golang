package endpoints

import "github.com/turkishjoe/xml-parser/internal/app/api"

type UpdateRequest struct {
}

type UpdateResponse struct {
	//	Documents []internal.Document `json:"documents"`
	//Err string `json:"err,omitempty"`
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
	Individuals []api.Individual
}

var searchTypeStringMap = map[string]api.SearchType{
	"strong": api.Strong,
	"weak":   api.Weak,
	"both":   api.Both,
}

var stateMap = map[api.State]string{
	api.Empty:    "empty",
	api.Updating: "updating",
	api.Ok:       "ok",
}
