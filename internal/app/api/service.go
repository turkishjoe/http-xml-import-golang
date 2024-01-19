package api

import (
	"context"
)

type SearchType uint
type State uint

const (
	Strong SearchType = iota
	Weak
	Both
)

const (
	Empty State = iota
	Updating
	Ok
)

type Individual struct {
	Uid       int    `json:"uid"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Service interface {
	// Get the list of all documents
	Update(ctx context.Context)
	State(ctx context.Context) State
	GetNames(ctx context.Context, name string, searchType SearchType) []Individual
}
