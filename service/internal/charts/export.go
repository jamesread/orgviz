package charts

import (
	"fmt"

	"github.com/jamesread/orgviz/internal/exportchart"
)

// ExportPNG renders a chart as a PNG. The file name matches the browser export.
func (s *Store) ExportPNG(id string) ([]byte, string, error) {
	chart, ok := s.Get(id)
	if !ok {
		return nil, "", fmt.Errorf("%w: %q", ErrNotFound, id)
	}

	data, err := exportchart.PNG(chart, s.ResolveStyle(chart.Style), s.avatarsDir)
	if err != nil {
		return nil, "", err
	}

	return data, exportchart.Filename(chart.Title), nil
}
