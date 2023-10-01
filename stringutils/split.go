package stringutils

import "strings"

func SplitComma(s string) (result []interface{}) {
	return split(s, ",")
}

func Split(s string) (result []interface{}) {
	return split(s, ".")
}

func split(s string, sep string) (result []interface{}) {
	if len(s) == 0 {
		return nil
	}
	splitteds := strings.Split(s, sep)
	for _, elt := range splitteds {
		elt = strings.TrimSpace(elt)
		result = append(result, elt)
	}
	return
}

func NameValuesToMap(list []string) map[string]interface{} {
	result := make(map[string]interface{})
	for _, elt := range list {
		index := strings.Index(elt, "=")
		result[elt[:index]] = elt[index+1:]
	}
	return result
}
