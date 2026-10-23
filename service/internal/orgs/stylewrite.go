package orgs

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// StyleChange is a change to an org file's style block. Nil strings leave that
// setting alone and an empty string removes it. Features sets switches by YAML
// key. Reset drops local css, backgroundColor and features so a theme applies
// as written.
type StyleChange struct {
	CSS             *string
	BackgroundColor *string
	Theme           *string
	Features        map[string]bool
	Reset           bool
}

var featureKeys = map[string]bool{
	"avatars":       true,
	"jobTitles":     true,
	"team":          true,
	"expandButtons": true,
	"compact":       true,
}

// SetStyle returns source with change applied to its style block. Comments and
// the rest of the document are kept, though indentation is normalised.
func SetStyle(source []byte, change StyleChange) ([]byte, error) {
	if err := checkFeatureKeys(change.Features); err != nil {
		return nil, err
	}

	root, doc, err := parseMappingDocument(source)
	if err != nil {
		return nil, err
	}

	applyStyleChange(root, change)

	return encodeAndCheck(doc)
}

func checkFeatureKeys(features map[string]bool) error {
	for key := range features {
		if !featureKeys[key] {
			return fmt.Errorf("unknown style feature %q (want avatars, jobTitles, team, expandButtons or compact)", key)
		}
	}

	return nil
}

func parseMappingDocument(source []byte) (*yaml.Node, *yaml.Node, error) {
	doc := &yaml.Node{}
	if err := yaml.Unmarshal(source, doc); err != nil {
		return nil, nil, fmt.Errorf("parsing org file: %w", err)
	}

	if len(doc.Content) == 0 {
		doc.Kind = yaml.DocumentNode
		doc.Content = []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}
	}

	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("org file must be a YAML mapping")
	}

	return root, doc, nil
}

func applyStyleChange(root *yaml.Node, change StyleChange) {
	style := ensureMapping(root, "style")

	applyThemeChange(style, change)
	setOptionalString(style, "backgroundColor", change.BackgroundColor)
	setOptionalString(style, "css", change.CSS)
	applyFeatureChange(style, change.Features)

	removeIfEmpty(style, "features")
	removeIfEmpty(root, "style")
}

func applyThemeChange(style *yaml.Node, change StyleChange) {
	if change.Reset {
		removeKey(style, "css")
		removeKey(style, "backgroundColor")
		removeKey(style, "features")
	}

	setOptionalString(style, "theme", change.Theme)
}

func applyFeatureChange(style *yaml.Node, features map[string]bool) {
	if len(features) == 0 {
		return
	}

	setFeatures(ensureMapping(style, "features"), features)
}

func setOptionalString(mapping *yaml.Node, key string, value *string) {
	if value == nil {
		return
	}

	setString(mapping, key, *value)
}

func setString(mapping *yaml.Node, key, value string) {
	if value == "" {
		removeKey(mapping, key)
		return
	}

	node := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}
	if strings.Contains(value, "\n") {
		node.Style = yaml.LiteralStyle
	}

	setValue(mapping, key, node)
}

func setFeatures(features *yaml.Node, values map[string]bool) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		setValue(features, key, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: fmt.Sprint(values[key])})
	}
}

func valueIndex(mapping *yaml.Node, key string) int {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return i + 1
		}
	}

	return -1
}

func setValue(mapping *yaml.Node, key string, value *yaml.Node) {
	if i := valueIndex(mapping, key); i >= 0 {
		mapping.Content[i] = value
		return
	}

	keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
	mapping.Content = append(mapping.Content, keyNode, value)
}

// ensureMapping returns the mapping under key, replacing a missing, empty or
// scalar value with an empty mapping.
func ensureMapping(mapping *yaml.Node, key string) *yaml.Node {
	if i := valueIndex(mapping, key); i >= 0 && mapping.Content[i].Kind == yaml.MappingNode {
		return mapping.Content[i]
	}

	node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	setValue(mapping, key, node)

	return node
}

func removeKey(mapping *yaml.Node, key string) {
	if i := valueIndex(mapping, key); i >= 0 {
		mapping.Content = append(mapping.Content[:i-1], mapping.Content[i+1:]...)
	}
}

func removeIfEmpty(mapping *yaml.Node, key string) {
	if i := valueIndex(mapping, key); i >= 0 && len(mapping.Content[i].Content) == 0 {
		removeKey(mapping, key)
	}
}

func encodeAndCheck(doc *yaml.Node) ([]byte, error) {
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(2)

	if err := encoder.Encode(doc); err != nil {
		return nil, fmt.Errorf("writing org file: %w", err)
	}

	if _, err := Parse(out.Bytes()); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
