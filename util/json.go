package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func SourceToTarget(src interface{}, target interface{}) {
	s := EncodeAsStringJson(src)
	json.Unmarshal([]byte(s), target)
}

func EncodeAsStringJson(o interface{}) string {
	b := &bytes.Buffer{}
	enc := json.NewEncoder(b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	if err := enc.Encode(o); err != nil {
		panic(fmt.Errorf("cannot encode as string  : %v", err))
	}
	return b.String()
}

type JsonIO struct {
	Object interface{}
}

func NewJsonIO(o interface{}) *JsonIO {
	return &JsonIO{
		Object: o,
	}
}

func (y *JsonIO) loadFromReader(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, y.Object); err != nil {
		return err
	}
	return nil
}

func (y *JsonIO) LoadFromString(s string) error {
	r := strings.NewReader(s)
	return y.loadFromReader(r)
}

func (y *JsonIO) SaveAsString() (string, error) {
	b := &bytes.Buffer{}
	if err := y.SaveToWriter(b); err != nil {
		return "", err
	}
	return b.String(), nil
}

func (y *JsonIO) SaveAsBytes() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := y.SaveToWriter(b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (y *JsonIO) SaveToWriter(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false) // very important when used in conjunction with grpc
	enc.SetIndent("", "    ")
	return enc.Encode(y.Object)
}
