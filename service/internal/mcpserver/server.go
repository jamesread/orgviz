// Package mcpserver serves orgviz's Model Context Protocol tools at /mcp, so
// an AI assistant can create and edit charts for a person to review in the
// browser.
package mcpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/jamesread/orgviz/internal/buildinfo"
	"github.com/jamesread/orgviz/internal/charts"
	"github.com/jamesread/orgviz/internal/orgs"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const instructions = `orgviz draws organisation charts from YAML files. Use these tools to create
or change a chart, then give the person the chart's "url" so they can review
it in their browser. Call orgviz_yaml_reference first for the file format.
Shared themes live in the themes directory; save one with orgviz_save_theme
and apply it with orgviz_set_chart_style theme. Call orgviz_export_chart to
get a PNG of the whole tree. Check the "warnings" in each result and fix them.`

type baseURLKey struct{}

type tools struct {
	store *charts.Store
}

// NewHandler returns the Streamable HTTP MCP endpoint, mounted at /mcp.
func NewHandler(store *charts.Store) http.Handler {
	mcpServer := server.NewMCPServer(
		"orgviz",
		buildinfo.Version,
		server.WithToolCapabilities(false),
		server.WithInstructions(instructions),
		server.WithRecovery(),
	)

	(&tools{store: store}).register(mcpServer)

	return server.NewStreamableHTTPServer(mcpServer,
		server.WithEndpointPath("/mcp"),
		server.WithStateLess(true),
		server.WithHTTPContextFunc(withBaseURL),
	)
}

func (t *tools) register(s *server.MCPServer) {
	chartID := mcp.WithString("chart_id", mcp.Required(), mcp.Description("Chart ID, the org file name without .yml, such as \"Acme\". Letters, digits, '-' and '_'."))
	yamlSource := mcp.WithString("yaml", mcp.Required(), mcp.Description("The complete org YAML. See orgviz_yaml_reference."))

	s.AddTool(mcp.NewTool("orgviz_ping",
		mcp.WithDescription("Check that orgviz is reachable. Returns its version."),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.ping)

	s.AddTool(mcp.NewTool("orgviz_yaml_reference",
		mcp.WithDescription("Describe the org YAML format: people, reporting lines, avatars and the style block."),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.yamlReference)

	s.AddTool(mcp.NewTool("orgviz_list_charts",
		mcp.WithDescription("List all charts with their ID, title, number of people and browser URL."),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.listCharts)

	s.AddTool(mcp.NewTool("orgviz_get_chart",
		mcp.WithDescription("Get one chart's YAML, resolved style, browser URL and warnings."),
		mcp.WithReadOnlyHintAnnotation(true),
		chartID,
	), t.getChart)

	s.AddTool(mcp.NewTool("orgviz_export_chart",
		mcp.WithDescription("Export a chart as a PNG image of the whole tree. Uses resolved card features and avatars. Use this to review how the chart looks."),
		mcp.WithReadOnlyHintAnnotation(true),
		chartID,
	), t.exportChart)

	s.AddTool(mcp.NewTool("orgviz_create_chart",
		mcp.WithDescription("Create a new chart from org YAML and save it on the server. Fails if the chart ID is taken. Returns the browser URL for review."),
		mcp.WithDestructiveHintAnnotation(false),
		chartID, yamlSource,
	), t.createChart)

	s.AddTool(mcp.NewTool("orgviz_update_chart",
		mcp.WithDescription("Replace an existing chart's YAML. The YAML is checked before saving; if it does not parse, the chart is left unchanged."),
		mcp.WithDestructiveHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		chartID, yamlSource,
	), t.updateChart)

	s.AddTool(mcp.NewTool("orgviz_set_chart_style",
		mcp.WithDescription("Change a chart's theme: shared theme, background color, custom CSS and card features. Only the arguments given are changed; the rest of the YAML is kept."),
		mcp.WithIdempotentHintAnnotation(true),
		chartID,
		mcp.WithString("theme", mcp.Description("Shared theme ID from orgviz_list_themes, such as \"ocean\". An empty string removes it. Passed alone, local css and features are cleared so the theme applies as written.")),
		mcp.WithString("background_color", mcp.Description("Background color for the chart area (<main />), such as \"#f8fafc\" or \"teal\"; an empty string removes it.")),
		mcp.WithString("css", mcp.Description("Custom CSS for the chart page; an empty string removes it. Cards use div.person, div.person.level-0 (root layer), img.avatar, .person h3, .person .job-title and .person .team.")),
		mcp.WithBoolean("avatars", mcp.Description("Show profile photos. Default true.")),
		mcp.WithBoolean("job_titles", mcp.Description("Show job titles. Default true.")),
		mcp.WithBoolean("team", mcp.Description("Show each person's team. Default false.")),
		mcp.WithBoolean("expand_buttons", mcp.Description("Show expand/collapse buttons. When false the whole tree is shown. Default true.")),
		mcp.WithBoolean("compact", mcp.Description("Use compact layout (reports side by side). When false, use an expanded tree. Default true.")),
	), t.setChartStyle)

	themeID := mcp.WithString("theme_id", mcp.Required(), mcp.Description("Theme ID, the file name without .yml, such as \"ocean\". Letters, digits, '-' and '_'."))

	s.AddTool(mcp.NewTool("orgviz_list_themes",
		mcp.WithDescription("List shared themes with their ID and title."),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.listThemes)

	s.AddTool(mcp.NewTool("orgviz_get_theme",
		mcp.WithDescription("Get one shared theme's title and style."),
		mcp.WithReadOnlyHintAnnotation(true),
		themeID,
	), t.getTheme)

	s.AddTool(mcp.NewTool("orgviz_save_theme",
		mcp.WithDescription("Create or replace a shared theme. Charts can select it with style.theme or orgviz_set_chart_style theme."),
		mcp.WithIdempotentHintAnnotation(true),
		themeID,
		mcp.WithString("title", mcp.Description("Display name, such as \"Ocean\". Defaults to the theme ID.")),
		mcp.WithString("background_color", mcp.Description("Background color for the chart area (<main />), such as \"#f8fafc\" or \"teal\".")),
		mcp.WithString("css", mcp.Description("Custom CSS for charts that use this theme. Cards use div.person, div.person.level-0 (root layer), img.avatar, .person h3, .person .job-title and .person .team.")),
		mcp.WithBoolean("avatars", mcp.Description("Show profile photos. Default true.")),
		mcp.WithBoolean("job_titles", mcp.Description("Show job titles. Default true.")),
		mcp.WithBoolean("team", mcp.Description("Show each person's team. Default false.")),
		mcp.WithBoolean("expand_buttons", mcp.Description("Show expand/collapse buttons. When false the whole tree is shown. Default true.")),
		mcp.WithBoolean("compact", mcp.Description("Use compact layout (reports side by side). When false, use an expanded tree. Default true.")),
	), t.saveTheme)

	s.AddTool(mcp.NewTool("orgviz_upload_avatar",
		mcp.WithDescription("Upload a profile photo to the avatars directory. Then set `avatar: <filename>` on the person in the chart YAML."),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Lowercase file name ending .png, .jpg, .jpeg, .webp or .svg, such as \"alice-smith.jpg\".")),
		mcp.WithString("image_base64", mcp.Required(), mcp.Description("The image, base64 encoded. A data: URL is also accepted. Up to 2 MiB.")),
		mcp.WithBoolean("overwrite", mcp.Description("Replace an existing file with the same name. Default false.")),
	), t.uploadAvatar)

	s.AddTool(mcp.NewTool("orgviz_list_avatars",
		mcp.WithDescription("List the image files in the avatars directory."),
		mcp.WithReadOnlyHintAnnotation(true),
	), t.listAvatars)
}

// withBaseURL records the URL the client used, so results can link to charts.
func withBaseURL(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, baseURLKey{}, requestBaseURL(r))
}

func requestBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	return scheme + "://" + firstNonEmpty(r.Header.Get("X-Forwarded-Host"), r.Host)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func absoluteURL(ctx context.Context, path string) string {
	base, _ := ctx.Value(baseURLKey{}).(string)
	return base + path
}

func chartURL(ctx context.Context, id string) string {
	return absoluteURL(ctx, "/chart/"+id)
}

func jsonResult(value any) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(string(data)), nil
}

func errorResult(err error) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultError(err.Error()), nil
}

func (t *tools) ping(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return jsonResult(map[string]string{"status": "ok", "version": buildinfo.Version})
}

func (t *tools) yamlReference(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(yamlReference), nil
}

// decodeImage accepts plain base64 or a data: URL.
func decodeImage(encoded string) ([]byte, error) {
	if _, data, found := strings.Cut(encoded, ";base64,"); found {
		encoded = data
	}

	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return nil, fmt.Errorf("image_base64 is not valid base64: %w", err)
	}

	return data, nil
}

// detail describes a chart for tool results, with the warnings an assistant
// should fix.
func (t *tools) detail(ctx context.Context, chart *orgs.File) chartDetail {
	warnings := chart.Warnings()
	for _, name := range t.store.MissingAvatars(chart) {
		warnings = append(warnings, fmt.Sprintf("avatar file %s is not uploaded yet", name))
	}

	if warning := t.store.ThemeWarning(chart.Style); warning != "" {
		warnings = append(warnings, warning)
	}

	return chartDetail{
		ChartID:  chart.ID,
		Title:    chart.Title,
		URL:      chartURL(ctx, chart.ID),
		YAML:     chart.Source,
		Style:    newStyleView(t.store.ResolveStyle(chart.Style)),
		Warnings: warnings,
	}
}
