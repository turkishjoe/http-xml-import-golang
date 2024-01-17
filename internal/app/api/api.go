package api

import "context"

type ApiService struct {
}

func NewService() Service {
	return &ApiService{}
}

func (w *ApiService) Update(ctx context.Context) {

}
