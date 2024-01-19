package endpoints

import (
	"context"
	"github.com/go-kit/kit/endpoint"
	"github.com/turkishjoe/xml-parser/internal/app/api"
)

type Set struct {
	UpdateEndpoint   endpoint.Endpoint
	StateEndpoint    endpoint.Endpoint
	GetNamesEndpoint endpoint.Endpoint
}

func NewEndpoints(svc api.Service) Set {
	return Set{
		UpdateEndpoint: MakeUpdateEndpoint(svc),
		//StateEndpoint:   MakeStateEndpoint(svc),
		GetNamesEndpoint: MakeGetNamesEndpoint(svc),
	}
}

func MakeUpdateEndpoint(svc api.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		//	req := request.(UpdateRequest)
		svc.Update(ctx)
		/*		if err != nil {
				return UpdateResponse{}, nil
			}*/
		return UpdateResponse{""}, nil
	}
}

func MakeGetNamesEndpoint(svc api.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetNameRequest)
		names := svc.GetNames(ctx, req.Name, api.CreateSearchTypeFromString(req.IndividualSearchType))

		//Для декода в пустом массиве
		if names == nil {
			names = make([]api.Individual, 0)
		}

		return GetNameResponse{Individuals: names}, nil
	}
}
