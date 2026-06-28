package yaml2

import (
	"encoding/json"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// JSONNode wraps *yaml.Node with proper JSON marshaling/unmarshaling.
// Instead of dumping the yaml.Node struct fields, it serializes as the
// logical value (map, list, scalar) the node represents.
type JSONNode struct {
	n *yaml.Node
}

func NewJSONNode(n *yaml.Node) *JSONNode {
	if n == nil {
		return nil
	}
	return &JSONNode{n: n}
}

func (j *JSONNode) Node() *yaml.Node {
	if j == nil {
		return nil
	}
	return j.n
}

func (j *JSONNode) MarshalJSON() ([]byte, error) {
	if j == nil || j.n == nil {
		return []byte("null"), nil
	}
	return json.Marshal(ConvertYamlNode(j.n))
}

func (j *JSONNode) UnmarshalJSON(data []byte) error {
	var iface interface{}
	if err := json.Unmarshal(data, &iface); err != nil {
		return err
	}
	j.n = interfaceToYamlNode(iface)
	return nil
}

func (j *JSONNode) UnmarshalYAML(value *yaml.Node) error {
	j.n = value
	return nil
}

func interfaceToYamlNode(v interface{}) *yaml.Node {
	if v == nil {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: ""}
	}
	switch val := v.(type) {
	case map[string]interface{}:
		node := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		for k, child := range val {
			node.Content = append(node.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k},
				interfaceToYamlNode(child),
			)
		}
		return node
	case []interface{}:
		node := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for _, child := range val {
			node.Content = append(node.Content, interfaceToYamlNode(child))
		}
		return node
	case string:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: val}
	case bool:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(val)}
	case float64:
		if val == float64(int64(val)) {
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", int64(val))}
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: strconv.FormatFloat(val, 'f', -1, 64)}
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: fmt.Sprintf("%v", val)}
	}
}
