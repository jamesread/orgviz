package clientapi

import (
	"connectrpc.com/connect"
	"context"
	"github.com/jamesread/orgviz/internal/buildinfo"
	pb "github.com/jamesread/orgviz/gen/orgviz/clientapi/v1"
)

type clientApi struct {
	People []*pb.Person
}

func NewServer() *clientApi {
	api := &clientApi{}
	api.People = make([]*pb.Person, 0)


	api.AddPerson(1, -1, "Alice")
	api.AddPerson(2, 1, "Bob")
	api.AddPerson(3, 1, "Charlie")
	api.AddPerson(4, 2, "David")
	api.AddPerson(5, 2, "Eve")
	api.AddPerson(6, 3, "Frank")

	return api
}

func (c *clientApi) AddPerson(id int32, parent int32, fullName string) {
	person := &pb.Person{
		Id:       id,
		ParentId:   parent,
		Name:     fullName,
		FullName: fullName,
	}

	c.People = append(c.People, person)
}

func (c *clientApi) GetClientInitialSettings(ctx context.Context, req *connect.Request[pb.GetClientInitialSettingsRequest]) (*connect.Response[pb.GetClientInitialSettingsResponse], error) {
	// This is where you would implement the logic to retrieve the initial settings for the client.
	// For now, we return an empty response.
	response := &pb.GetClientInitialSettingsResponse{
		Version: buildinfo.Version,
	}

	return connect.NewResponse(response), nil

}

func (c *clientApi) GetChart(ctx context.Context, req *connect.Request[pb.GetChartRequest]) (*connect.Response[pb.GetChartResponse], error) {
	response := &pb.GetChartResponse{
		People: c.People,
	}

	return connect.NewResponse(response), nil
}

//func (c *ClientApi) Init(ctx context.Context, req *connect.Request[pb.InitRequest]) (*connect.Response[pb.InitResponse], error) {
