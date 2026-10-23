package orgs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const commentedOrg = `# Our org
title: Demo
people:
  - name: Alice # the boss
`

func strPtr(value string) *string {
	return &value
}

func TestSetStyleAddsFeaturesAndKeepsComments(t *testing.T) {
	updated, err := SetStyle([]byte(commentedOrg), StyleChange{Features: map[string]bool{"team": true, "avatars": false}})
	require.NoError(t, err)

	require.Contains(t, string(updated), "# Our org")
	require.Contains(t, string(updated), "# the boss")

	org, err := Parse(updated)
	require.NoError(t, err)
	features := org.Style.Features.Resolve()
	require.True(t, features.Team)
	require.False(t, features.Avatars)
	require.Equal(t, "Alice", org.People[0].Name)
}

func TestSetStyleWritesAndClearsCSS(t *testing.T) {
	css := "div.person {\n  border-top-color: teal;\n}\n"

	updated, err := SetStyle([]byte(commentedOrg), StyleChange{CSS: &css})
	require.NoError(t, err)
	require.Contains(t, string(updated), "css: |")

	org, err := Parse(updated)
	require.NoError(t, err)
	require.Equal(t, css, org.Style.CSS)

	cleared, err := SetStyle(updated, StyleChange{CSS: strPtr("")})
	require.NoError(t, err)
	require.NotContains(t, string(cleared), "style")
}

func TestSetStyleWritesAndClearsBackgroundColor(t *testing.T) {
	updated, err := SetStyle([]byte(commentedOrg), StyleChange{BackgroundColor: strPtr("#f8fafc")})
	require.NoError(t, err)
	require.Contains(t, string(updated), "backgroundColor:")

	org, err := Parse(updated)
	require.NoError(t, err)
	require.Equal(t, "#f8fafc", org.Style.BackgroundColor)

	cleared, err := SetStyle(updated, StyleChange{BackgroundColor: strPtr("")})
	require.NoError(t, err)
	require.NotContains(t, string(cleared), "style")
}

func TestSetStyleKeepsOtherSettings(t *testing.T) {
	updated, err := SetStyle([]byte(commentedOrg), StyleChange{
		BackgroundColor: strPtr("#f8fafc"),
		Features:        map[string]bool{"team": true},
	})
	require.NoError(t, err)

	updated, err = SetStyle(updated, StyleChange{CSS: strPtr("h3 { color: red; }")})
	require.NoError(t, err)

	org, err := Parse(updated)
	require.NoError(t, err)
	require.Equal(t, "#f8fafc", org.Style.BackgroundColor)
	require.True(t, org.Style.Features.Resolve().Team)
	require.Equal(t, "h3 { color: red; }", org.Style.CSS)
}

func TestSetStyleWritesThemeAndClearsOverrides(t *testing.T) {
	updated, err := SetStyle([]byte(commentedOrg), StyleChange{
		CSS:      strPtr("h3 { color: red; }"),
		Features: map[string]bool{"team": true},
	})
	require.NoError(t, err)

	updated, err = SetStyle(updated, StyleChange{Theme: strPtr("ocean"), Reset: true})
	require.NoError(t, err)

	org, err := Parse(updated)
	require.NoError(t, err)
	require.Equal(t, "ocean", org.Style.Theme)
	require.Equal(t, "", org.Style.CSS)
	require.False(t, org.Style.Features.Resolve().Team)

	cleared, err := SetStyle(updated, StyleChange{Theme: strPtr("")})
	require.NoError(t, err)
	require.NotContains(t, string(cleared), "style")
}

func TestSetStyleRejectsUnknownFeature(t *testing.T) {
	_, err := SetStyle([]byte(commentedOrg), StyleChange{Features: map[string]bool{"sparkles": true}})
	require.ErrorContains(t, err, "sparkles")
}

func TestSetStyleRejectsInvalidYAML(t *testing.T) {
	_, err := SetStyle([]byte("people: ["), StyleChange{CSS: strPtr("x")})
	require.Error(t, err)
}
