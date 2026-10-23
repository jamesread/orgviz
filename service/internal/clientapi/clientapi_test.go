package clientapi

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"
	pb "github.com/jamesread/orgviz/gen/orgviz/clientapi/v1"
	"github.com/jamesread/orgviz/internal/charts"
	"github.com/jamesread/orgviz/internal/orgs"
	"github.com/stretchr/testify/require"
)

func testServer(orgsToServe []*orgs.File) *clientAPI {
	return NewServer(charts.NewFromFiles(orgsToServe...))
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

func TestGetChartReturnsSource(t *testing.T) {
	source := "title: Demo\npeople:\n  - name: Alice\n"
	org, err := orgs.Parse([]byte(source))
	require.NoError(t, err)

	api := testServer([]*orgs.File{org})
	resp, err := api.GetChart(context.Background(), connect.NewRequest(&pb.GetChartRequest{ChartId: "0"}))
	require.NoError(t, err)
	require.Equal(t, source, resp.Msg.Source)
}

func TestPreviewChartReturnsPeople(t *testing.T) {
	api := testServer(nil)
	resp, err := api.PreviewChart(context.Background(), connect.NewRequest(&pb.PreviewChartRequest{Source: `
title: Draft
people:
  - name: Alice
    alias: ali
  - name: Bob
    reports: ali
`}))
	require.NoError(t, err)
	require.Equal(t, "Draft", resp.Msg.Title)
	require.Len(t, resp.Msg.People, 2)
	require.Equal(t, int32(1), resp.Msg.People[1].ParentId)
}

func TestPreviewChartRejectsInvalidYAML(t *testing.T) {
	api := testServer(nil)
	_, err := api.PreviewChart(context.Background(), connect.NewRequest(&pb.PreviewChartRequest{Source: "people: ["}))
	require.Error(t, err)
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestPreviewChartRejectsOversizedSource(t *testing.T) {
	api := testServer(nil)
	_, err := api.PreviewChart(context.Background(), connect.NewRequest(&pb.PreviewChartRequest{Source: strings.Repeat("#", maxPreviewSourceBytes+1)}))
	require.Error(t, err)
	require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestGetChartReturnsStyleAndTeam(t *testing.T) {
	org, err := orgs.Parse([]byte(`
title: Demo
style:
  backgroundColor: "#f8fafc"
  css: "h3 { color: red; }"
  features:
    jobTitles: false
    compact: false
people:
  - name: Alice
    team: Leadership
`))
	require.NoError(t, err)

	api := testServer([]*orgs.File{org})
	resp, err := api.GetChart(context.Background(), connect.NewRequest(&pb.GetChartRequest{ChartId: "0"}))
	require.NoError(t, err)
	require.Equal(t, "Leadership", resp.Msg.People[0].Team)
	require.Equal(t, "#f8fafc", resp.Msg.Style.BackgroundColor)
	require.Equal(t, "h3 { color: red; }", resp.Msg.Style.Css)
	require.True(t, resp.Msg.Style.ShowAvatars)
	require.False(t, resp.Msg.Style.ShowJobTitles)
	require.False(t, resp.Msg.Style.ShowTeam)
	require.True(t, resp.Msg.Style.ShowExpandButtons)
	require.False(t, resp.Msg.Style.Compact)
}

func TestSaveAndListThemes(t *testing.T) {
	dir := t.TempDir()
	store := charts.New("", "", dir)
	api := NewServer(store)

	_, err := api.SaveTheme(context.Background(), connect.NewRequest(&pb.SaveThemeRequest{
		ThemeId: "ocean",
		Title:   "Ocean",
		Style:   &pb.ChartStyle{Css: "h3 { color: red; }", ShowTeam: true, BackgroundColor: "#eee"},
	}))
	require.NoError(t, err)

	listed, err := api.ListThemes(context.Background(), connect.NewRequest(&pb.ListThemesRequest{}))
	require.NoError(t, err)
	require.Equal(t, []*pb.ThemeInfo{{ThemeId: "ocean", Title: "Ocean"}}, listed.Msg.Themes)

	resp, err := api.GetClientInitialSettings(context.Background(), connect.NewRequest(&pb.GetClientInitialSettingsRequest{}))
	require.NoError(t, err)
	require.Equal(t, "ocean", resp.Msg.Themes[0].ThemeId)
}

func TestPreviewChartResolvesTheme(t *testing.T) {
	dir := t.TempDir()
	store := charts.New("", "", dir)
	_, err := store.SaveTheme(&orgs.Theme{ID: "ocean", Title: "Ocean", Style: orgs.Style{CSS: "h3 { color: red; }"}})
	require.NoError(t, err)

	api := NewServer(store)
	resp, err := api.PreviewChart(context.Background(), connect.NewRequest(&pb.PreviewChartRequest{Source: `
style:
  theme: ocean
people:
  - name: Alice
`}))
	require.NoError(t, err)
	require.Equal(t, "ocean", resp.Msg.Style.Theme)
	require.Equal(t, "h3 { color: red; }", resp.Msg.Style.Css)
}

func TestPreviewChartReturnsStyle(t *testing.T) {
	api := testServer(nil)
	resp, err := api.PreviewChart(context.Background(), connect.NewRequest(&pb.PreviewChartRequest{Source: `
style:
  backgroundColor: teal
  features:
    team: true
people:
  - name: Alice
`}))
	require.NoError(t, err)
	require.Equal(t, "teal", resp.Msg.Style.BackgroundColor)
	require.True(t, resp.Msg.Style.ShowTeam)
}
