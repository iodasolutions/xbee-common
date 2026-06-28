package yaml2

import "gopkg.in/yaml.v3"

// YAMLNode wraps *yaml.Node with proper YAML marshaling/unmarshaling.
// A bare *yaml.Node field in a struct is not handled correctly by yaml.v3:
// the decoder allocates an empty Node struct and tries to fill it via struct
// field matching instead of capturing the raw node. This wrapper fixes that
// by implementing UnmarshalYAML/MarshalYAML.
type YAMLNode struct {
	n *yaml.Node
}

func NewYAMLNode(n *yaml.Node) *YAMLNode {
	if n == nil {
		return nil
	}
	return &YAMLNode{n: n}
}

func (y *YAMLNode) Node() *yaml.Node {
	if y == nil {
		return nil
	}
	return y.n
}

func (y *YAMLNode) MarshalYAML() (interface{}, error) {
	if y == nil || y.n == nil {
		return nil, nil
	}
	return y.n, nil
}

func (y *YAMLNode) UnmarshalYAML(value *yaml.Node) error {
	y.n = value
	return nil
}
