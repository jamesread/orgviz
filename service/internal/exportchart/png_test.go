package exportchart

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/jamesread/orgviz/internal/orgs"
	"github.com/stretchr/testify/require"
)

func TestFilename(t *testing.T) {
	require.Equal(t, "example-company.png", Filename("Example Company"))
	require.Equal(t, "org.png", Filename(""))
	require.Equal(t, "org.png", Filename("!!!"))
}

func TestPNGEncodesAChart(t *testing.T) {
	chart, err := orgs.Parse([]byte("title: Example Company\npeople:\n  - name: Alice\n    title: CEO\n  - name: Bob\n    reports: Alice\n"))
	require.NoError(t, err)

	data, err := PNG(chart, orgs.Style{}, "")
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(data, []byte("\x89PNG")))

	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)
	require.GreaterOrEqual(t, img.Bounds().Dx(), cardWidth)
	require.GreaterOrEqual(t, img.Bounds().Dy(), 40)
}

func TestPNGRejectsEmptyChart(t *testing.T) {
	_, err := PNG(&orgs.File{Title: "Empty"}, orgs.Style{}, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no people")
}
