package yaml2

import "gopkg.in/yaml.v3"

func EmptyMap() *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}
}
