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
	"strong":  Strong,
	"Premium": Weak,
	"Both":    Both,
}

type Individual struct {
	uid        int
	first_name string
	last_name  string
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
