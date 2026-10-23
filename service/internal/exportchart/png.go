package exportchart

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"github.com/jamesread/orgviz/internal/orgs"
)

// PNG draws the chart as a transparent-background PNG, the same kind of
// image the browser Export PNG button produces.
func PNG(chart *orgs.File, style orgs.Style, avatarsDir string) ([]byte, error) {
	if chart == nil || len(chart.People) == 0 {
		return nil, fmt.Errorf("chart %q has no people", chartTitle(chart))
	}

	return encodePNG(drawChart(layoutChart(chart, style), style, avatarsDir))
}

func chartTitle(chart *orgs.File) string {
	if chart == nil {
		return ""
	}

	return chart.Title
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encoding chart PNG: %w", err)
	}

	return buf.Bytes(), nil
}
