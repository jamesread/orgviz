package exportchart

import (
	"testing"

	"github.com/jamesread/orgviz/internal/orgs"
	"github.com/stretchr/testify/require"
)

func layoutFrom(t *testing.T, source string, compact bool) chartLayout {
	t.Helper()
	chart, err := orgs.Parse([]byte(source))
	require.NoError(t, err)
	features := orgs.StyleFeatures{}
	if !compact {
		off := false
		features.Compact = &off
	}
	chart.Style.Features = features
	return layoutChart(chart, chart.Style)
}

func TestLayoutSinglePersonFitsACard(t *testing.T) {
	laid := layoutFrom(t, "people:\n  - name: Alice\n", true)
	require.Len(t, laid.cards, 1)
	require.GreaterOrEqual(t, laid.cards[0].w, 200)
	require.GreaterOrEqual(t, laid.cards[0].h, 20)
	require.GreaterOrEqual(t, laid.w, laid.cards[0].w)
	require.GreaterOrEqual(t, laid.h, laid.cards[0].h)
}

func TestLayoutPlacesReportsBelowManager(t *testing.T) {
	laid := layoutFrom(t, "people:\n  - name: Alice\n  - name: Bob\n    reports: Alice\n  - name: Carol\n    reports: Alice\n", false)
	require.Len(t, laid.cards, 3)
	alice := cardNamed(laid, "Alice")
	bob := cardNamed(laid, "Bob")
	carol := cardNamed(laid, "Carol")
	require.Greater(t, bob.y, alice.y+alice.h)
	require.Equal(t, bob.y, carol.y)
	require.NotEqual(t, bob.x, carol.x)
}

func TestLayoutCompactStacksLeafPairs(t *testing.T) {
	source := "people:\n  - name: Alice\n  - name: Bob\n    reports: Alice\n  - name: Carol\n    reports: Alice\n  - name: Dan\n    reports: Alice\n  - name: Eve\n    reports: Alice\n"
	compact := layoutFrom(t, source, true)
	expanded := layoutFrom(t, source, false)
	require.Less(t, compact.w, expanded.w)
	require.Greater(t, cardNamed(compact, "Dan").y, cardNamed(compact, "Bob").y)
	require.Equal(t, cardNamed(compact, "Bob").y, cardNamed(compact, "Carol").y)
}

func cardNamed(laid chartLayout, name string) placed {
	for _, card := range laid.cards {
		if card.person.Name == name {
			return card
		}
	}
	return placed{}
}
