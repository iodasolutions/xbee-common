package types

import (
	"path/filepath"
	"strings"
)

func ShortNameFromOrigin(origin string) string {
	var part string
	if strings.Contains(origin, "://") {
		part = strings.TrimSuffix(origin, ".git")
	} else {
		part = filepath.ToSlash(origin)
	}
	return part[strings.LastIndex(part, "/")+1:]
}

type Origin struct {
	Repo   string `yaml:"repo,omitempty"`
	Commit string `yaml:"commit,omitempty"`
	Ref    string `yaml:"ref,omitempty"`
}

// IdJson is used to identify one of three xbee abstractions: product, application, environment.
type IdJson struct {
	Origin
	Alias string `yaml:"alias,omitempty"`
}

func (m *IdJson) Clone() IdJson {
	return *m
}

func (m *IdJson) AsMap() map[string]interface{} {
	result := make(map[string]interface{})
	result["origin"] = m.Origin
	if m.Alias != "" {
		result["alias"] = m.Alias
	}
	return result
}
