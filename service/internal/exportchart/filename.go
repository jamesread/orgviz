package exportchart

import (
	"regexp"
	"strings"
)

var notSlug = regexp.MustCompile(`[^a-z0-9]+`)

// Filename is the PNG name for a chart title, matching the browser export.
func Filename(title string) string {
	slug := strings.Trim(notSlug.ReplaceAllString(strings.ToLower(title), "-"), "-")
	if slug == "" {
		slug = "org"
	}

	return slug + ".png"
}
