package yaml2

import (
	"strconv"

	"gopkg.in/yaml.v3"
)

func AddBool(node *yaml.Node, key string, value string) {
	addScalar(node, key, "!!bool", value)
}

func AddString(node *yaml.Node, key string, value string) {
	addScalar(node, key, "!!str", value)
}

func AddInt(node *yaml.Node, key string, value int) {
	addScalar(node, key, "!!int", strconv.Itoa(value))
}

func addScalar(node *yaml.Node, key string, aKind string, value string) {
	// --- cas mapping ---
	if node.Kind == yaml.MappingNode {

		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
		}

		valueNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   aKind,
			Value: value,
		}

		node.Content = append(node.Content, keyNode, valueNode)
		return
	}

	// --- cas sequence ---
	if node.Kind == yaml.SequenceNode {

		valueNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   aKind,
			Value: value,
		}

		node.Content = append(node.Content, valueNode)
		return
	}

	// --- cas node vide → on le transforme en scalaire ---
	node.Kind = yaml.ScalarNode
	node.Tag = aKind
	node.Value = value
}

func AddContent(node *yaml.Node, key string, content []*yaml.Node) {
	node.Content = append(node.Content, &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}, &yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: content,
	})
}
