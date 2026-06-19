package yaml2

import "gopkg.in/yaml.v3"

func PromoteScalarToSequence(node *yaml.Node) {
	if node.Kind != yaml.ScalarNode {
		panic("This should never happen")
	}
	oldValue := node.Value
	oldTag := node.Tag
	node.Kind = yaml.SequenceNode
	node.Tag = "!!seq"
	node.Value = ""
	node.Content = []*yaml.Node{
		{
			Kind:  yaml.ScalarNode,
			Tag:   oldTag,
			Value: oldValue,
		},
	}
}

func PromoteScalarToMap(node *yaml.Node, key string) {
	if node.Kind != yaml.ScalarNode {
		panic("This should never happen")
	}
	oldValue := node.Value
	oldTag := node.Tag
	node.Kind = yaml.MappingNode
	node.Tag = "!!map"
	node.Value = ""
	node.Content = []*yaml.Node{
		{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
		},
		{
			Kind:  yaml.ScalarNode,
			Tag:   oldTag,
			Value: oldValue,
		},
	}
}

func PromoteInPlace(node *yaml.Node, key string) {
	// 1. copier l'ancien node
	old := CloneNode(node)

	// 2. transformer le node courant en map
	node.Kind = yaml.MappingNode
	node.Tag = "!!map"

	// reset propre
	node.Value = ""
	node.Content = nil

	// 3. injecter key + ancienne valeur
	node.Content = []*yaml.Node{
		{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
		},
		old,
	}
}
