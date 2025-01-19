package main

import (
	"encoding/json"
	"fmt"
	"github.com/iodasolutions/xbee-common/util"
)

type PElement struct {
	Provider map[string]interface{} `json:"provider,omitempty"`
}
type XbeeVolume struct {
	PElement
	Name string `json:"name,omitempty"`
	Size int    `json:"size,omitempty"`
}

func main() {
	a := new(XbeeVolume)
	a.Provider = map[string]interface{}{
		"cpus": 2,
	}
	a.Name = "name"
	fmt.Println(util.EncodeAsStringJson(a))
	s := `
{
    "provider": {
        "cpus": 2
    },
    "name": "name"
}
`
	var b XbeeVolume
	if err := json.Unmarshal([]byte(s), &b); err != nil {
		panic(err)
	}
	fmt.Println(b)
}
