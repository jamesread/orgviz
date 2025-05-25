package clientapi

import (
	"connectrpc.com/connect"
	"context"
	"github.com/jamesread/orgviz/internal/buildinfo"
	pb "github.com/jamesread/orgviz/gen/orgviz/clientapi/v1"
)

type clientApi struct {

}

func NewServer() *clientApi {
	return &clientApi{}
}

func (c *clientApi) GetClientInitialSettings(ctx context.Context, req *connect.Request[pb.GetClientInitialSettingsRequest]) (*connect.Response[pb.GetClientInitialSettingsResponse], error) {
	// This is where you would implement the logic to retrieve the initial settings for the client.
	// For now, we return an empty response.
	response := &pb.GetClientInitialSettingsResponse{
		Version: buildinfo.Version,
	}

	return connect.NewResponse(response), nil

}
//func (c *ClientApi) Init(ctx context.Context, req *connect.Request[pb.InitRequest]) (*connect.Response[pb.InitResponse], error) {
