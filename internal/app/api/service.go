package api

import (
	"context"
)

type SearchType uint

const (
	Strong SearchType = iota
	Weak
	Both
)

var searchTypeStringMap = map[string]SearchType{
	"strong": Strong,
	"weak":   Weak,
	"both":   Both,
}

type Individual struct {
	Uid       int    `json:"uid"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Service interface {
	// Get the list of all documents
	Update(ctx context.Context)
	//	State(ctx context.Context, ticketID string)
	GetNames(ctx context.Context, name string, searchType SearchType) []Individual
}

func CreateSearchTypeFromString(searchTypeRaw string) SearchType {
	v, ok := searchTypeStringMap[searchTypeRaw]

	if !ok {
		return Both
	}

	return v
}
