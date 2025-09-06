package template

import (
	"bytes"
	"io"
	"net"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/iodasolutions/xbee-common/cmd"
)

func Output(s *string, data interface{}, funcMap map[string]interface{}) (err *cmd.XbeeError) {
	templateString := *s
	buf := &bytes.Buffer{}
	err = OutputWithTemplate(templateString, buf, data, funcMap)
	if err == nil {
		*s = buf.String()
	}
	return
}

func OutputWithTemplate(templateString string, wr io.Writer, data interface{}, funcMap map[string]interface{}) *cmd.XbeeError {

	appliedFuncMap := DefaultFunctions()
	if funcMap != nil {
		for key, value := range funcMap {
			appliedFuncMap[key] = value
		}
	}
	splitted := SplitNonParsedSections(templateString)
	for i, section := range splitted {
		if i%2 == 0 {
			tmpl, err := template.New("test").
				Funcs(appliedFuncMap).
				Parse(section)
			if err != nil {
				return cmd.Error("cannot load template to parse %s: %v", section, err)
			}
			buf := bytes.Buffer{}
			err = tmpl.Execute(&buf, data)
			if err != nil {
				return cmd.Error("cannot parse template to parse %s with data %v: %v", section, data, err)
			}
			if _, err := wr.Write(buf.Bytes()); err != nil {
				return cmd.Error("cannot write data: %v", err)
			}
		} else {
			if _, err := wr.Write([]byte(section)); err != nil {
				return cmd.Error("cannot write section: %v", err)
			}
		}
	}
	return nil
}

func SplitNonParsedSections(s string) (splitted []string) {
	tokenStart := "##xbee-start##"
	tokenEnd := "##xbee-end##"

	for len(s) > 0 {
		indexStart := strings.Index(s, tokenStart)
		if indexStart != -1 {
			indexEnd := strings.Index(s[indexStart+len(tokenStart):], tokenEnd)
			splitted = append(splitted, s[:indexStart])
			splitted = append(splitted, s[indexStart+len(tokenStart):indexEnd+indexStart+len(tokenStart)])
			s = s[indexStart+len(tokenStart)+indexEnd+len(tokenEnd):]
		} else {
			splitted = append(splitted, s)
			s = ""
		}
	}
	return
}

func DefaultFunctions() template.FuncMap {
	return template.FuncMap{
		"lower":        strings.ToLower,
		"basePath":     filepath.Base,
		"dirPath":      dirPath,
		"ip":           ip,
		"majorVersion": majorVersion,
		"minorVersion": minorVersion,
		"patchVersion": patchVersion,
		"gt":           gt,
		"mapArch":      mapArch,
	}
}

func ErrorFromCLI(errors []string, usage string) *cmd.XbeeError {
	return cmd.Error("\nERROR(s):\n%s\n\nUSAGE:\n%s", strings.Join(errors, "\n"), usage)
}

func dirPath(value interface{}) string {
	s := value.(string)
	return filepath.ToSlash(filepath.Dir(s))
}

func ip(value interface{}) string {
	host := value.(string)
	addr, err := net.LookupIP(host)
	if err != nil {
		panic(err)
	}
	return addr[0].String()
}

func majorVersion(value interface{}) string {
	version := value.(string)
	if !strings.Contains(version, ".") {
		return version
	}
	result := version[:strings.Index(version, ".")]
	return result
}
func minorVersion(value interface{}) string {
	version := value.(string)
	if !strings.Contains(version, ".") {
		return ""
	}
	return strings.Split(version, ".")[1]
}
func patchVersion(value interface{}) string {
	version := value.(string)
	if !strings.Contains(version, ".") {
		return ""
	}
	splitted := strings.Split(version, ".")
	if len(splitted) < 3 {
		return ""
	}
	return splitted[2]
}

func gt(a, b interface{}) bool {
	switch v1 := a.(type) {
	case int:
		switch v2 := b.(type) {
		case int:
			return v1 > v2
		case float64:
			return float64(v1) > v2
		}
	case float64:
		switch v2 := b.(type) {
		case int:
			return v1 > float64(v2)
		case float64:
			return v1 > v2
		}
	}
	return false
}

func mapArch(arch string) string {
	switch arch {
	case "amd64":
		return "x64"
	case "arm64":
		return "aarch64"
	}
	return "unknown"
}
