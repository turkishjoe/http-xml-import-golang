package endpoints

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/turkishjoe/xml-parser/internal/app/api"
	"github.com/turkishjoe/xml-parser/internal/app/api/domain"
	"net/http"
)

type Set struct {
	UpdateEndpoint   endpoint.Endpoint
	StateEndpoint    endpoint.Endpoint
	GetNamesEndpoint endpoint.Endpoint
}

func NewEndpoints(svc api.Service) Set {
	return Set{
		UpdateEndpoint:   MakeUpdateEndpoint(svc),
		StateEndpoint:    MakeStateEndpoint(svc),
		GetNamesEndpoint: MakeGetNamesEndpoint(svc),
	}
}

func MakeUpdateEndpoint(svc api.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		err := svc.Update(ctx)

		if err != nil {
			return UpdateResponse{
				false,
				"service unavailable",
				http.StatusServiceUnavailable,
			}, nil
		}

		return UpdateResponse{
			true,
			"ok",
			http.StatusOK,
		}, nil
	}
}

func MakeStateEndpoint(svc api.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		state := svc.State(ctx)

		return StateResponse{Result: state == api.Ok, Info: stateMap[state]}, nil
	}
}

func MakeGetNamesEndpoint(svc api.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetNameRequest)

		searchType, ok := searchTypeStringMap[req.IndividualSearchType]

		if !ok {
			searchType = domain.Both
		}

		names := svc.GetNames(ctx, req.Name, searchType)

		//Для декода в пустом массиве
		if names == nil {
			names = make([]domain.Individual, 0)
		}

		return GetNameResponse{Individuals: names}, nil
	}
}
