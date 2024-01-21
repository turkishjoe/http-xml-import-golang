package api

import (
	"context"
	"github.com/turkishjoe/xml-parser/internal/app/api/domain"
)

type State uint

const (
	Empty State = iota
	Updating
	Ok
)

type Service interface {
	// Get the list of all documents
	Update(ctx context.Context)
	State(ctx context.Context) State
	GetNames(ctx context.Context, name string, searchType domain.SearchType) []domain.Individual
}
