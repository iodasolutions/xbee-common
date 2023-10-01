package template

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"text/template"
)

func Output(s *string, data interface{}, funcMap map[string]interface{}) (err error) {
	templateString := *s
	buf := &bytes.Buffer{}
	err = OutputWithTemplate(templateString, buf, data, funcMap)
	if err == nil {
		*s = buf.String()
	}
	return
}

func OutputWithTemplate(templateString string, wr io.Writer, data interface{}, funcMap map[string]interface{}) error {

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
				return err
			}
			buf := bytes.Buffer{}
			err = tmpl.Execute(&buf, data)
			if err != nil {
				return err
			}
			if _, err := wr.Write(buf.Bytes()); err != nil {
				return err
			}
		} else {
			if _, err := wr.Write([]byte(section)); err != nil {
				return err
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
	}
}

func ErrorFromCLI(errors []string, usage string) error {
	return fmt.Errorf("\nERROR(s):\n%s\n\nUSAGE:\n%s", strings.Join(errors, "\n"), usage)
}

func dirPath(s string) string {
	return filepath.ToSlash(filepath.Dir(s))
}

func ip(host string) string {
	addr, err := net.LookupIP(host)
	if err != nil {
		panic(err)
	}
	return addr[0].String()
}

func majorVersion(version string) string {
	if !strings.Contains(version, ".") {
		return version
	}
	result := version[:strings.Index(version, ".")]
	return result
}
func minorVersion(version string) string {
	if !strings.Contains(version, ".") {
		return ""
	}
	return strings.Split(version, ".")[1]
}
func patchVersion(version string) string {
	if !strings.Contains(version, ".") {
		return ""
	}
	splitted := strings.Split(version, ".")
	if len(splitted) < 3 {
		return ""
	}
	return splitted[2]
}
