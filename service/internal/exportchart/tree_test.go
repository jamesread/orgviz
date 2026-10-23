package exportchart

import (
	"testing"

	"github.com/jamesread/orgviz/internal/orgs"
	"github.com/stretchr/testify/require"
)

func parsePeople(t *testing.T, source string) []*orgs.Person {
	t.Helper()
	chart, err := orgs.Parse([]byte(source))
	require.NoError(t, err)
	return chart.People
}

func TestTreeEmpty(t *testing.T) {
	require.Nil(t, tree(nil))
}

func TestTreeRootAndReports(t *testing.T) {
	root := tree(parsePeople(t, "people:\n  - name: Alice\n    alias: ali\n  - name: Bob\n    reports: ali\n  - name: Carol\n    reports: Bob\n"))
	require.Equal(t, "Alice", root.person.Name)
	require.Len(t, root.children, 1)
	require.Equal(t, "Bob", root.children[0].person.Name)
	require.Equal(t, "Carol", root.children[0].children[0].person.Name)
}

func TestTreeUnknownReportsAttachToFirstPerson(t *testing.T) {
	root := tree(parsePeople(t, "people:\n  - name: Alice\n  - name: Bob\n    reports: Nobody\n"))
	require.Equal(t, "Alice", root.person.Name)
	require.Equal(t, "Bob", root.children[0].person.Name)
}
