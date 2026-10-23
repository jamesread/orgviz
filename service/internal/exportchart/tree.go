package exportchart

import "github.com/jamesread/orgviz/internal/orgs"

type personNode struct {
	person   *orgs.Person
	children []*personNode
}

func tree(people []*orgs.Person) *personNode {
	if len(people) == 0 {
		return nil
	}

	return attach(people[0], groupByManager(people))
}

func groupByManager(people []*orgs.Person) map[int32][]*orgs.Person {
	rootID := people[0].ID
	groups := map[int32][]*orgs.Person{}

	for _, person := range people[1:] {
		parent := person.ReportsToID
		if parent == 0 {
			parent = rootID
		}
		groups[parent] = append(groups[parent], person)
	}

	return groups
}

func attach(person *orgs.Person, byParent map[int32][]*orgs.Person) *personNode {
	node := &personNode{person: person}
	for _, child := range byParent[person.ID] {
		node.children = append(node.children, attach(child, byParent))
	}

	return node
}

func allLeaves(node *personNode) bool {
	for _, child := range node.children {
		if len(child.children) > 0 {
			return false
		}
	}

	return true
}
