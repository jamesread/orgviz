package orgs

// Style is the optional `style:` block of an org file. Theme names a shared
// theme file; the other fields override that theme when set.
type Style struct {
	Theme           string        `yaml:"theme,omitempty"`
	CSS             string        `yaml:"css,omitempty"`
	BackgroundColor string        `yaml:"backgroundColor,omitempty"`
	Features        StyleFeatures `yaml:"features,omitempty"`
}

// StyleFeatures holds the feature switches as written in the file. A nil
// value means the switch was left out and its default applies.
type StyleFeatures struct {
	Avatars       *bool `yaml:"avatars,omitempty"`
	JobTitles     *bool `yaml:"jobTitles,omitempty"`
	Team          *bool `yaml:"team,omitempty"`
	ExpandButtons *bool `yaml:"expandButtons,omitempty"`
	Compact       *bool `yaml:"compact,omitempty"`
}

// ResolvedFeatures is StyleFeatures with defaults applied.
type ResolvedFeatures struct {
	Avatars       bool
	JobTitles     bool
	Team          bool
	ExpandButtons bool
	Compact       bool
}

// Resolve applies the defaults: everything is shown except the team.
func (f StyleFeatures) Resolve() ResolvedFeatures {
	return ResolvedFeatures{
		Avatars:       boolOr(f.Avatars, true),
		JobTitles:     boolOr(f.JobTitles, true),
		Team:          boolOr(f.Team, false),
		ExpandButtons: boolOr(f.ExpandButtons, true),
		Compact:       boolOr(f.Compact, true),
	}
}

func boolOr(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}

	return *value
}

func stringOr(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func ptrOr[T any](value, fallback *T) *T {
	if value == nil {
		return fallback
	}

	return value
}

// Overlay fills empty fields from base, so a chart can name a theme and
// override only some of it.
func (s Style) Overlay(base Style) Style {
	return Style{
		Theme:           s.Theme,
		CSS:             stringOr(s.CSS, base.CSS),
		BackgroundColor: stringOr(s.BackgroundColor, base.BackgroundColor),
		Features: StyleFeatures{
			Avatars:       ptrOr(s.Features.Avatars, base.Features.Avatars),
			JobTitles:     ptrOr(s.Features.JobTitles, base.Features.JobTitles),
			Team:          ptrOr(s.Features.Team, base.Features.Team),
			ExpandButtons: ptrOr(s.Features.ExpandButtons, base.Features.ExpandButtons),
			Compact:       ptrOr(s.Features.Compact, base.Features.Compact),
		},
	}
}
