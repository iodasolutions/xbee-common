package types

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/newfs"
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
	Version string `json:"version,omitempty"`
	Origin  string `json:"origin,omitempty"`
	Commit  string `json:"commit,omitempty"`
	Alias   string `json:"alias,omitempty"`
}

// Colon should be used only for log purpose
func (m *IdJson) Colon() string { return m.withDelimiter(":") }
func (m *IdJson) OriginVersion() string {
	var version string
	if m.Version != "" {
		version = fmt.Sprintf(":%s", m.Version)
	}
	return fmt.Sprintf("%s%s", m.Origin, version)
}
func (m *IdJson) ShortName() string {
	if m.Alias != "" {
		return m.Alias
	}
	return ShortNameFromOrigin(m.Origin)
}

func (m *IdJson) withDelimiter(delimiter string) string {
	shortOrigin := ShortNameFromOrigin(delimiter)
	var extension string
	if m.Commit != "" {
		extension = delimiter + m.Commit
	}
	return fmt.Sprintf("%s%s", shortOrigin, extension)
}

func (m *IdJson) Equals(other *IdJson) bool {
	if m == nil {
		return other == nil
	}
	return other != nil && m.Version == other.Version && m.Origin == other.Origin
}

func (m *IdJson) Clone() IdJson {
	return *m
}

func (m *IdJson) AsMap() map[string]interface{} {
	result := make(map[string]interface{})
	result["origin"] = m.Origin
	if m.Version != "" {
		result["version"] = m.Version
	}
	if m.Alias != "" {
		result["alias"] = m.Alias
	}
	return result
}

func (m *IdJson) Hash() string {
	result := m.ShortName()
	sa := newfs.NewShaAccumulator()
	sa.AddString(m.Commit)
	sa.AddString(m.Origin)
	sha := sa.Sha()
	if sha != "" {
		result = result + "-" + sha
	}
	return result
}
