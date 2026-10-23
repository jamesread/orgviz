package mcpserver

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/jamesread/orgviz/internal/charts"
	"github.com/jamesread/orgviz/internal/orgs"
	"github.com/mark3labs/mcp-go/mcp"
)

type chartSummary struct {
	ChartID string `json:"chart_id"`
	Title   string `json:"title"`
	People  int    `json:"people"`
	URL     string `json:"url"`
}

type chartDetail struct {
	ChartID  string    `json:"chart_id"`
	Title    string    `json:"title"`
	URL      string    `json:"url"`
	YAML     string    `json:"yaml"`
	Style    styleView `json:"style"`
	Warnings []string  `json:"warnings"`
}

type styleView struct {
	Theme           string `json:"theme"`
	CSS             string `json:"css"`
	BackgroundColor string `json:"background_color"`
	Avatars         bool   `json:"avatars"`
	JobTitles       bool   `json:"job_titles"`
	Team            bool   `json:"team"`
	ExpandButtons   bool   `json:"expand_buttons"`
	Compact         bool   `json:"compact"`
}

func newStyleView(style orgs.Style) styleView {
	features := style.Features.Resolve()

	return styleView{
		Theme:           style.Theme,
		CSS:             style.CSS,
		BackgroundColor: style.BackgroundColor,
		Avatars:         features.Avatars,
		JobTitles:       features.JobTitles,
		Team:            features.Team,
		ExpandButtons:   features.ExpandButtons,
		Compact:         features.Compact,
	}
}

type themeSummary struct {
	ThemeID string `json:"theme_id"`
	Title   string `json:"title"`
}

type themeDetail struct {
	ThemeID string    `json:"theme_id"`
	Title   string    `json:"title"`
	Style   styleView `json:"style"`
}

// styleArgs maps set_chart_style arguments to the YAML feature keys.
var styleArgs = map[string]string{
	"avatars":        "avatars",
	"job_titles":     "jobTitles",
	"team":           "team",
	"expand_buttons": "expandButtons",
	"compact":        "compact",
}

func (t *tools) listCharts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	summaries := []chartSummary{}

	for _, chart := range t.store.List() {
		summaries = append(summaries, chartSummary{
			ChartID: chart.ID,
			Title:   chart.Title,
			People:  len(chart.People),
			URL:     chartURL(ctx, chart.ID),
		})
	}

	return jsonResult(summaries)
}

func (t *tools) exportChart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("chart_id")
	if err != nil {
		return errorResult(err)
	}

	data, name, err := t.store.ExportPNG(id)
	if err != nil {
		return errorResult(err)
	}

	return mcp.NewToolResultImage(
		fmt.Sprintf("Exported %s. Open %s to review the live chart.", name, chartURL(ctx, id)),
		base64.StdEncoding.EncodeToString(data),
		"image/png",
	), nil
}

func (t *tools) getChart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("chart_id")
	if err != nil {
		return errorResult(err)
	}

	chart, found := t.store.Get(id)
	if !found {
		return errorResult(fmt.Errorf("%w: %q", charts.ErrNotFound, id))
	}

	return jsonResult(t.detail(ctx, chart))
}

func (t *tools) createChart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return t.writeChart(ctx, req, t.store.Create)
}

func (t *tools) updateChart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return t.writeChart(ctx, req, t.store.Update)
}

func (t *tools) writeChart(ctx context.Context, req mcp.CallToolRequest, write func(id, source string) (*orgs.File, error)) (*mcp.CallToolResult, error) {
	id, source, err := chartArgs(req)
	if err != nil {
		return errorResult(err)
	}

	chart, err := write(id, source)

	return t.chartResult(ctx, chart, err)
}

func chartArgs(req mcp.CallToolRequest) (string, string, error) {
	id, err := req.RequireString("chart_id")
	if err != nil {
		return "", "", err
	}

	source, err := req.RequireString("yaml")

	return id, source, err
}

func (t *tools) chartResult(ctx context.Context, chart *orgs.File, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return errorResult(err)
	}

	return jsonResult(t.detail(ctx, chart))
}

func (t *tools) setChartStyle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("chart_id")
	if err != nil {
		return errorResult(err)
	}

	change, err := styleChange(req.GetArguments())
	if err != nil {
		return errorResult(err)
	}

	chart, err := t.store.SetStyle(id, change)

	return t.chartResult(ctx, chart, err)
}

func styleChange(args map[string]any) (orgs.StyleChange, error) {
	change := orgs.StyleChange{
		CSS:             stringArg(args, "css"),
		BackgroundColor: stringArg(args, "background_color"),
		Theme:           stringArg(args, "theme"),
		Features:        featureArgs(args),
	}
	change.Reset = shouldResetOverrides(change)

	if !hasStyleChange(change) {
		return change, errors.New("nothing to change: pass theme, css, background_color, or at least one of avatars, job_titles, team, expand_buttons, compact")
	}

	return change, nil
}

func shouldResetOverrides(change orgs.StyleChange) bool {
	return applyingTheme(change) && !hasLocalStyle(change)
}

func applyingTheme(change orgs.StyleChange) bool {
	return change.Theme != nil && *change.Theme != ""
}

func hasLocalStyle(change orgs.StyleChange) bool {
	return change.CSS != nil || change.BackgroundColor != nil || len(change.Features) > 0
}

func stringArg(args map[string]any, key string) *string {
	value, ok := args[key].(string)
	if !ok {
		return nil
	}

	return &value
}

func hasStyleChange(change orgs.StyleChange) bool {
	return change.CSS != nil || change.BackgroundColor != nil || change.Theme != nil || len(change.Features) > 0
}

func (t *tools) listThemes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	summaries := []themeSummary{}
	for _, theme := range t.store.ListThemes() {
		summaries = append(summaries, themeSummary{ThemeID: theme.ID, Title: theme.Title})
	}

	return jsonResult(summaries)
}

func (t *tools) getTheme(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("theme_id")
	if err != nil {
		return errorResult(err)
	}

	theme, found := t.store.GetTheme(id)
	if !found {
		return errorResult(fmt.Errorf("%w: %q", charts.ErrNotFound, id))
	}

	return jsonResult(themeDetail{ThemeID: theme.ID, Title: theme.Title, Style: newStyleView(theme.Style)})
}

func (t *tools) saveTheme(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("theme_id")
	if err != nil {
		return errorResult(err)
	}

	theme, err := t.store.SaveTheme(themeFromArgs(id, req.GetArguments()))
	if err != nil {
		return errorResult(err)
	}

	return jsonResult(themeDetail{ThemeID: theme.ID, Title: theme.Title, Style: newStyleView(theme.Style)})
}

func themeFromArgs(id string, args map[string]any) *orgs.Theme {
	title, _ := args["title"].(string)
	theme := &orgs.Theme{ID: id, Title: title}
	if css, ok := args["css"].(string); ok {
		theme.Style.CSS = css
	}
	if color, ok := args["background_color"].(string); ok {
		theme.Style.BackgroundColor = color
	}
	theme.Style.Features = featurePointers(featureArgs(args))

	return theme
}

func featurePointers(values map[string]bool) orgs.StyleFeatures {
	features := orgs.StyleFeatures{}
	setFeature := map[string]**bool{
		"avatars":       &features.Avatars,
		"jobTitles":     &features.JobTitles,
		"team":          &features.Team,
		"expandButtons": &features.ExpandButtons,
		"compact":       &features.Compact,
	}

	for key, dest := range setFeature {
		if value, ok := values[key]; ok {
			copied := value
			*dest = &copied
		}
	}

	return features
}

// featureArgs collects the feature switches that were passed.
func featureArgs(args map[string]any) map[string]bool {
	features := map[string]bool{}

	for arg, key := range styleArgs {
		if value, ok := args[arg].(bool); ok {
			features[key] = value
		}
	}

	return features
}

func (t *tools) uploadAvatar(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, data, err := avatarArgs(req)
	if err != nil {
		return errorResult(err)
	}

	url, err := t.store.SaveAvatar(name, data, req.GetBool("overwrite", false))
	if err != nil {
		return errorResult(err)
	}

	return jsonResult(map[string]string{
		"filename": name,
		"url":      absoluteURL(ctx, url),
		"next":     fmt.Sprintf("Add `avatar: %s` to the person in the chart YAML.", name),
	})
}

func avatarArgs(req mcp.CallToolRequest) (string, []byte, error) {
	name, err := req.RequireString("filename")
	if err != nil {
		return "", nil, err
	}

	encoded, err := req.RequireString("image_base64")
	if err != nil {
		return "", nil, err
	}

	data, err := decodeImage(encoded)

	return name, data, err
}

func (t *tools) listAvatars(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	names, err := t.store.ListAvatars()
	if err != nil {
		return errorResult(err)
	}

	return jsonResult(names)
}
