package api

import (
	"context"
)

type Service interface {
	// Get the list of all documents
	Update(ctx context.Context)
	//	State(ctx context.Context, ticketID string)
	//	GetNames(ctx context.Context, ticketID, mark string)
}
