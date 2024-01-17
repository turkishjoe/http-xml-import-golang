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
		//GetNamesEndpoint: MakeGetNamesEndpoint(svc),
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

/*func MakeGetNamesEndpoint(svc api.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(StatusRequest)
		status, err := svc.Status(ctx, req.TicketID)
		if err != nil {
			return StatusResponse{Status: status, Err: err.Error()}, nil
		}
		return StatusResponse{Status: status, Err: ""}, nil
	}
}*/
