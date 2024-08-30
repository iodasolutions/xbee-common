package types

import (
	"fmt"
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
	Repo   string `json:"repo,omitempty"`
	Commit string `json:"commit,omitempty"`
}

// IdJson is used to identify one of three xbee abstractions: product, application, environment.
type IdJson struct {
	Origin string `json:"origin,omitempty"`
	Commit string `json:"commit,omitempty"`
	Alias  string `json:"alias,omitempty"`
}

// Colon should be used only for log purpose
func (m *IdJson) Colon() string { return m.withDelimiter(":") }

func (m *IdJson) ShortName() string {
	if m.Alias != "" {
		return m.Alias
	}
	return ShortNameFromOrigin(m.Origin)
}

func (m *IdJson) withDelimiter(delimiter string) string {
	var extension string
	if m.Commit != "" {
		extension = delimiter + m.Commit
	}
	return fmt.Sprintf("%s%s", m.Origin, extension)
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
