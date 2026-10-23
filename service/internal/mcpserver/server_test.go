package mcpserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesread/orgviz/internal/charts"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
)

const acme = "title: Acme\npeople:\n  - name: Alice\n    alias: ali\n  - name: Bob\n    reports: ali\n"

// A 1x1 PNG.
var pngBytes = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\rIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x18\xd8N\x00\x00\x00\x00IEND\xaeB`\x82")

func newTestTools(t *testing.T) *tools {
	t.Helper()
	orgsDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(orgsDir, "Acme.yml"), []byte(acme), 0o644))

	return &tools{store: charts.New(orgsDir, t.TempDir(), t.TempDir())}
}

func call(t *testing.T, handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error), args map[string]any) (*mcp.CallToolResult, string) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Arguments = args

	ctx := context.WithValue(context.Background(), baseURLKey{}, "http://orgviz.test")
	result, err := handler(ctx, req)
	require.NoError(t, err)
	require.NotEmpty(t, result.Content)

	return result, result.Content[0].(mcp.TextContent).Text
}

func decode(t *testing.T, text string, into any) {
	t.Helper()
	require.NoError(t, json.Unmarshal([]byte(text), into))
}

func TestListCharts(t *testing.T) {
	tools := newTestTools(t)

	result, text := call(t, tools.listCharts, nil)
	require.False(t, result.IsError)

	var charts []chartSummary
	decode(t, text, &charts)
	require.Equal(t, []chartSummary{{ChartID: "Acme", Title: "Acme", People: 2, URL: "http://orgviz.test/chart/Acme"}}, charts)
}

func TestExportChart(t *testing.T) {
	tools := newTestTools(t)

	result, text := call(t, tools.exportChart, map[string]any{"chart_id": "Acme"})
	require.False(t, result.IsError, text)
	require.Contains(t, text, "acme.png")
	require.Contains(t, text, "http://orgviz.test/chart/Acme")
	require.Len(t, result.Content, 2)
	image := result.Content[1].(mcp.ImageContent)
	require.Equal(t, "image/png", image.MIMEType)
	raw, err := base64.StdEncoding.DecodeString(image.Data)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(raw), "\x89PNG"))

	result, text = call(t, tools.exportChart, map[string]any{"chart_id": "Nope"})
	require.True(t, result.IsError)
	require.Contains(t, text, "not found")
}

func TestGetChart(t *testing.T) {
	tools := newTestTools(t)

	_, text := call(t, tools.getChart, map[string]any{"chart_id": "Acme"})

	var chart chartDetail
	decode(t, text, &chart)
	require.Equal(t, acme, chart.YAML)
	require.True(t, chart.Style.Avatars)
	require.Empty(t, chart.Warnings)

	result, text := call(t, tools.getChart, map[string]any{"chart_id": "Nope"})
	require.True(t, result.IsError)
	require.Contains(t, text, "not found")
}

func TestCreateAndUpdateChart(t *testing.T) {
	tools := newTestTools(t)

	result, text := call(t, tools.createChart, map[string]any{"chart_id": "NewCo", "yaml": "title: NewCo\npeople:\n  - name: Ann\n  - name: Ben\n    reports: Zed\n"})
	require.False(t, result.IsError, text)

	var chart chartDetail
	decode(t, text, &chart)
	require.Equal(t, "http://orgviz.test/chart/NewCo", chart.URL)
	require.Equal(t, []string{`Ben reports to "Zed", which matches no name or alias`}, chart.Warnings)

	result, text = call(t, tools.createChart, map[string]any{"chart_id": "NewCo", "yaml": acme})
	require.True(t, result.IsError)
	require.Contains(t, text, "already exists")

	result, text = call(t, tools.updateChart, map[string]any{"chart_id": "NewCo", "yaml": "title: Renamed\npeople:\n  - name: Ann\n"})
	require.False(t, result.IsError, text)
	decode(t, text, &chart)
	require.Equal(t, "Renamed", chart.Title)
}

func TestCreateChartRejectsBadYAML(t *testing.T) {
	tools := newTestTools(t)

	result, text := call(t, tools.createChart, map[string]any{"chart_id": "Bad", "yaml": "people: ["})
	require.True(t, result.IsError)
	require.Contains(t, text, "invalid")
}

func TestSetChartStyle(t *testing.T) {
	tools := newTestTools(t)

	result, text := call(t, tools.setChartStyle, map[string]any{
		"chart_id":         "Acme",
		"background_color": "#f8fafc",
		"css":              "h3 { color: teal; }",
		"avatars":          false,
		"expand_buttons":   false,
		"compact":          false,
	})
	require.False(t, result.IsError, text)

	var chart chartDetail
	decode(t, text, &chart)
	require.Equal(t, "#f8fafc", chart.Style.BackgroundColor)
	require.Equal(t, "h3 { color: teal; }", chart.Style.CSS)
	require.False(t, chart.Style.Avatars)
	require.True(t, chart.Style.JobTitles)
	require.False(t, chart.Style.ExpandButtons)
	require.False(t, chart.Style.Compact)
	require.Contains(t, chart.YAML, "backgroundColor: '#f8fafc'")
	require.Contains(t, chart.YAML, "expandButtons: false")
	require.Contains(t, chart.YAML, "compact: false")

	result, text = call(t, tools.setChartStyle, map[string]any{
		"chart_id":         "Acme",
		"background_color": "",
	})
	require.False(t, result.IsError, text)
	decode(t, text, &chart)
	require.Equal(t, "", chart.Style.BackgroundColor)
	require.NotContains(t, chart.YAML, "backgroundColor")

	result, _ = call(t, tools.setChartStyle, map[string]any{"chart_id": "Acme"})
	require.True(t, result.IsError, "a call with nothing to change")
}

func TestSaveAndApplyTheme(t *testing.T) {
	tools := newTestTools(t)

	result, text := call(t, tools.saveTheme, map[string]any{
		"theme_id": "ocean",
		"title":    "Ocean",
		"css":      "h3 { color: red; }",
		"team":     true,
	})
	require.False(t, result.IsError, text)

	var theme themeDetail
	decode(t, text, &theme)
	require.Equal(t, "ocean", theme.ThemeID)
	require.Equal(t, "h3 { color: red; }", theme.Style.CSS)

	_, text = call(t, tools.listThemes, nil)
	require.Contains(t, text, "ocean")

	_, text = call(t, tools.getTheme, map[string]any{"theme_id": "ocean"})
	require.Contains(t, text, "Ocean")

	result, text = call(t, tools.setChartStyle, map[string]any{"chart_id": "Acme", "theme": "ocean"})
	require.False(t, result.IsError, text)

	var chart chartDetail
	decode(t, text, &chart)
	require.Equal(t, "ocean", chart.Style.Theme)
	require.Equal(t, "h3 { color: red; }", chart.Style.CSS)
	require.True(t, chart.Style.Team)
	require.Contains(t, chart.YAML, "theme: ocean")
	require.NotContains(t, chart.YAML, "css:")
}

func TestUploadAvatar(t *testing.T) {
	tools := newTestTools(t)

	result, text := call(t, tools.uploadAvatar, map[string]any{
		"filename":     "alice.png",
		"image_base64": "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes),
	})
	require.False(t, result.IsError, text)
	require.Contains(t, text, "/avatars/alice.png")
	require.Contains(t, text, "avatar: alice.png")

	_, text = call(t, tools.listAvatars, nil)
	require.Contains(t, text, "alice.png")

	result, text = call(t, tools.uploadAvatar, map[string]any{"filename": "bob.png", "image_base64": "%%%"})
	require.True(t, result.IsError)
	require.Contains(t, text, "base64")

	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="1" height="1"/>`)
	result, text = call(t, tools.uploadAvatar, map[string]any{
		"filename":     "carol.svg",
		"image_base64": base64.StdEncoding.EncodeToString(svg),
	})
	require.False(t, result.IsError, text)
	require.Contains(t, text, "/avatars/carol.svg")
}

func TestGetChartWarnsAboutMissingAvatar(t *testing.T) {
	tools := newTestTools(t)

	_, text := call(t, tools.updateChart, map[string]any{"chart_id": "Acme", "yaml": "people:\n  - name: Alice\n    avatar: alice.png\n"})

	var chart chartDetail
	decode(t, text, &chart)
	require.Equal(t, []string{"avatar file alice.png is not uploaded yet"}, chart.Warnings)
}

func TestYAMLReferenceDescribesFormat(t *testing.T) {
	_, text := call(t, (&tools{}).yamlReference, nil)
	for _, want := range []string{"people:", "reports:", "alias:", "avatar:", "style:", "expandButtons", "compact:"} {
		require.Contains(t, text, want)
	}
}

func TestHandlerServesToolsOverHTTP(t *testing.T) {
	server := httptest.NewServer(NewHandler(newTestTools(t).store))
	defer server.Close()

	post := func(body string) string {
		req, err := http.NewRequest(http.MethodPost, server.URL+"/mcp", strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var out strings.Builder
		_, err = io.Copy(&out, resp.Body)
		require.NoError(t, err)
		return out.String()
	}

	initialize := post(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}`)
	require.Contains(t, initialize, `"name":"orgviz"`)

	list := post(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	for _, name := range []string{"orgviz_list_charts", "orgviz_create_chart", "orgviz_update_chart", "orgviz_set_chart_style", "orgviz_export_chart", "orgviz_list_themes", "orgviz_save_theme", "orgviz_upload_avatar"} {
		require.Contains(t, list, name)
	}
}
