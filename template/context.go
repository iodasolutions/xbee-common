package template

import (
	"strconv"
	"strings"
)

type ContextMap map[string]interface{}

func (p *ContextMap) Eval(path string) interface{} {
	var result interface{}
	pathElements := strings.Split(path, ".")
	var value map[string]interface{} = *p
	for _, elt := range pathElements {
		result = value[elt]
		if _, ok := result.(map[string]interface{}); ok {
			value = result.(map[string]interface{})
		} else if _, ok := result.(*ContextMap); ok {
			ctx := result.(*ContextMap)
			value = *ctx
		}
	}
	return result
}

func (p *ContextMap) EvalAsString(path string) string {
	raw := p.Eval(path)
	switch x := raw.(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case bool:
		if x {
			return "true"
		} else {
			return "false"
		}
	default:
		return "???"
	}
}
