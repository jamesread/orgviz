package clientapi

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	pb "github.com/jamesread/orgviz/gen/orgviz/clientapi/v1"
	"github.com/jamesread/orgviz/internal/orgs"
	"github.com/stretchr/testify/require"
)

func testServer(orgsToServe []*orgs.File) *clientAPI {
	return &clientAPI{orgfiles: orgsToServe}
}

func TestGetChartUnknownID(t *testing.T) {
	api := testServer(nil)
	_, err := api.GetChart(context.Background(), connect.NewRequest(&pb.GetChartRequest{ChartId: "9"}))
	require.Error(t, err)
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestGetChartReturnsPeople(t *testing.T) {
	org, err := orgs.Parse([]byte(`
title: Demo
people:
  - name: Alice
    title: CEO
`))
	require.NoError(t, err)

	api := testServer([]*orgs.File{org})
	resp, err := api.GetChart(context.Background(), connect.NewRequest(&pb.GetChartRequest{ChartId: "0"}))
	require.NoError(t, err)
	require.Equal(t, "Demo", resp.Msg.Title)
	require.Len(t, resp.Msg.People, 1)
	require.Equal(t, "Alice", resp.Msg.People[0].FullName)
	require.Equal(t, "/avatars/alice.jpg", resp.Msg.People[0].AvatarUrl)
}
