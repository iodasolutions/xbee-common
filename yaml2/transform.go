package yaml2

import "gopkg.in/yaml.v3"

func AddOrReplace(parent *yaml.Node, key string, child *yaml.Node) {
	if parent.Kind != yaml.MappingNode {
		panic("not a yaml mapping")
	}
	found := false
	for i := 0; i < len(parent.Content); i += 2 {
		theKey := parent.Content[i].Value
		if theKey == key {
			found = true
			parent.Content[i+1] = child
			break
		}
	}
	if !found {
		parent.Content = append(parent.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: key,
		})
		parent.Content = append(parent.Content, child)
	}
}
