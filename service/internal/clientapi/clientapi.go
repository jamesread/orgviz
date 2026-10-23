package clientapi

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	pb "github.com/jamesread/orgviz/gen/orgviz/clientapi/v1"
	"github.com/jamesread/orgviz/internal/buildinfo"
	"github.com/jamesread/orgviz/internal/charts"
	"github.com/jamesread/orgviz/internal/orgs"
)

// maxPreviewSourceBytes caps the YAML accepted by PreviewChart.
const maxPreviewSourceBytes = charts.MaxSourceBytes

type clientAPI struct {
	store *charts.Store
}

// NewServer serves the client API from the chart store.
func NewServer(store *charts.Store) *clientAPI {
	return &clientAPI{store: store}
}

func (c *clientAPI) GetClientInitialSettings(ctx context.Context, req *connect.Request[pb.GetClientInitialSettingsRequest]) (*connect.Response[pb.GetClientInitialSettingsResponse], error) {
	return connect.NewResponse(&pb.GetClientInitialSettingsResponse{
		Version: buildinfo.Version,
		Charts:  chartInfos(c.store.List()),
		Themes:  themeInfos(c.store.ListThemes()),
	}), nil
}

func (c *clientAPI) GetChart(ctx context.Context, req *connect.Request[pb.GetChartRequest]) (*connect.Response[pb.GetChartResponse], error) {
	chart, err := c.chartByID(req.Msg.ChartId)
	if err != nil {
		return nil, err
	}

	response := &pb.GetChartResponse{
		ChartId: chart.ID,
		Title:   chart.Title,
		People:  chartPeopleToPb(chart.People),
		Source:  chart.Source,
		Style:   c.resolvedStyle(chart.Style),
	}

	return connect.NewResponse(response), nil
}

// PreviewChart parses unsaved org YAML from the editor pane. Nothing is stored.
func (c *clientAPI) PreviewChart(ctx context.Context, req *connect.Request[pb.PreviewChartRequest]) (*connect.Response[pb.PreviewChartResponse], error) {
	if len(req.Msg.Source) > maxPreviewSourceBytes {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("org YAML is larger than %d bytes", maxPreviewSourceBytes))
	}

	chart, err := orgs.Parse([]byte(req.Msg.Source))
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	response := &pb.PreviewChartResponse{
		Title:  chart.Title,
		People: chartPeopleToPb(chart.People),
		Style:  c.resolvedStyle(chart.Style),
	}

	return connect.NewResponse(response), nil
}

func (c *clientAPI) ListThemes(ctx context.Context, req *connect.Request[pb.ListThemesRequest]) (*connect.Response[pb.ListThemesResponse], error) {
	return connect.NewResponse(&pb.ListThemesResponse{Themes: themeInfos(c.store.ListThemes())}), nil
}

func (c *clientAPI) SaveTheme(ctx context.Context, req *connect.Request[pb.SaveThemeRequest]) (*connect.Response[pb.SaveThemeResponse], error) {
	theme, err := c.store.SaveTheme(themeFromRequest(req.Msg))
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&pb.SaveThemeResponse{
		Theme: &pb.ThemeInfo{ThemeId: theme.ID, Title: theme.Title},
	}), nil
}

func (c *clientAPI) chartByID(chartID string) (*orgs.File, error) {
	chart, found := c.store.Get(chartID)
	if !found {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("unknown chart %q", chartID))
	}

	return chart, nil
}

func chartPeopleToPb(people []*orgs.Person) []*pb.Person {
	pbPeople := make([]*pb.Person, len(people))

	for i, person := range people {
		pbPeople[i] = &pb.Person{
			FullName:  person.Name,
			JobTitle:  person.Title,
			Id:        person.ID,
			ParentId:  person.ReportsToID,
			AvatarUrl: person.AvatarURL(),
			Team:      person.Team,
		}
	}

	return pbPeople
}

func (c *clientAPI) resolvedStyle(style orgs.Style) *pb.ChartStyle {
	return chartStyleToPb(c.store.ResolveStyle(style))
}

func chartInfos(files []*orgs.File) []*pb.ChartInfo {
	infos := make([]*pb.ChartInfo, 0, len(files))
	for _, chart := range files {
		infos = append(infos, &pb.ChartInfo{ChartId: chart.ID, Title: chart.Title})
	}

	return infos
}

func themeInfos(themes []*orgs.Theme) []*pb.ThemeInfo {
	infos := make([]*pb.ThemeInfo, 0, len(themes))
	for _, theme := range themes {
		infos = append(infos, &pb.ThemeInfo{ThemeId: theme.ID, Title: theme.Title})
	}

	return infos
}

func themeFromRequest(req *pb.SaveThemeRequest) *orgs.Theme {
	theme := &orgs.Theme{ID: req.GetThemeId(), Title: req.GetTitle()}
	if req.GetStyle() != nil {
		theme.Style = styleFromPb(req.GetStyle())
	}

	return theme
}

func styleFromPb(style *pb.ChartStyle) orgs.Style {
	return orgs.Style{
		Theme:           style.GetTheme(),
		CSS:             style.GetCss(),
		BackgroundColor: style.GetBackgroundColor(),
		Features: orgs.StyleFeatures{
			Avatars:       boolPtr(style.GetShowAvatars()),
			JobTitles:     boolPtr(style.GetShowJobTitles()),
			Team:          boolPtr(style.GetShowTeam()),
			ExpandButtons: boolPtr(style.GetShowExpandButtons()),
			Compact:       boolPtr(style.GetCompact()),
		},
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func chartStyleToPb(style orgs.Style) *pb.ChartStyle {
	features := style.Features.Resolve()

	return &pb.ChartStyle{
		Css:               style.CSS,
		BackgroundColor:   style.BackgroundColor,
		ShowAvatars:       features.Avatars,
		ShowJobTitles:     features.JobTitles,
		ShowTeam:          features.Team,
		ShowExpandButtons: features.ExpandButtons,
		Theme:             style.Theme,
		Compact:           features.Compact,
	}
}
