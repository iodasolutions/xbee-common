package newfs

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/iodasolutions/xbee-common/util"
	"hash"
)

type ShaAccumulator struct {
	h     hash.Hash
	empty bool
}

func NewShaAccumulator() (sa *ShaAccumulator) {
	return &ShaAccumulator{
		h:     sha1.New(),
		empty: true,
	}
}

func (sa *ShaAccumulator) AddObject(o interface{}) {
	if s, err := util.NewJsonIO(o).SaveAsString(); err != nil {
		panic(fmt.Errorf("unexpected exception when serializing object to json2 string : %s\n", err))
	} else {
		if s == "" {
			return
		}
		sa.AddString(s)
	}
	sa.empty = false
}
func (sa *ShaAccumulator) AddPath(p Path) {
	if p == "" {
		return
	}
	if _, err := sa.h.Write([]byte(p.Sha1())); err != nil {
		panic(err)
	}
	sa.empty = false
}
func (sa *ShaAccumulator) AddString(s string) {
	if s == "" {
		return
	}
	if _, err := sa.h.Write([]byte(s)); err != nil {
		panic(err)
	}
	sa.empty = false
}
func (sa *ShaAccumulator) AddBytes(b []byte) {
	if len(b) == 0 {
		return
	}
	if _, err := sa.h.Write(b); err != nil {
		panic(err)
	}
	sa.empty = false
}

func (sa *ShaAccumulator) Sha() string {
	if sa.empty {
		return ""
	}
	hashBytes := sa.h.Sum(nil)
	s := hex.EncodeToString(hashBytes)
	return s[:10]
}
