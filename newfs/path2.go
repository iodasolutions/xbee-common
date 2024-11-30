package newfs

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/stringutils"
	"os"
	"path/filepath"
	"strings"
)

type Path2 string

func (p Path2) String() string {
	return string(p)
}

func (p Path2) Exists() bool {
	_, err := os.Stat(string(p))
	if err == nil {
		return true
	}
	return false
}

func (p Path2) Child(name string) Path2 {
	return Path2(filepath.ToSlash(filepath.Join(string(p), name)))
}

func (p Path2) BaseWithoutExtension() string {
	base := filepath.Base(string(p))
	index := strings.LastIndex(base, ".")
	if index != -1 {
		return base[:index]
	}
	return base
}

func (p Path2) Extension() string {
	base := filepath.Base(string(p))
	index := strings.LastIndex(base, ".")
	if index != -1 {
		return base[index+1:]
	}
	return ""
}

func (p Path2) Base() string {
	return filepath.Base(string(p))
}

func (p Path2) Hash() string {
	return p.Base() + "-" + stringutils.Sha1StringTruncated(filepath.Dir(string(p)))
}
func (p Path2) WithoutHash() Path2 {
	return Path2(p.String()[:len(p.String())-11])
}

func (p Path2) IsDir() bool {
	fi, err := os.Lstat(string(p))
	if err == nil {
		return fi.IsDir()
	}
	return false
}
func (p Path2) IsLink() bool {
	fi, err := os.Lstat(string(p))
	if err == nil {
		return fi.Mode()&os.ModeSymlink != 0
	}
	return false
}

func (p Path2) IsAbs() bool {
	return filepath.IsAbs(string(p))
}
func (p Path2) IsRelative() bool {
	return !p.IsAbs()
}
func (p Path2) IsDescendantOf(another Path2) bool {
	return strings.HasPrefix(string(p), string(another))
}

func (p Path2) IsSymLink() bool {
	fi, err := os.Stat(string(p))
	if err == nil {
		return fi.Mode()&os.ModeSymlink == os.ModeSymlink
	}
	return false
}

func (p Path2) RelativeFrom(another Path2) Path2 {
	if p.IsRelative() {
		panic(fmt.Errorf("method Path.RelativeFrom : path %s is not an absolute path", p))
	}
	if p.IsDescendantOf(another) {
		folderS := another.String()
		pS := p.String()
		trimmed := strings.TrimPrefix(pS, folderS)
		trimmed = strings.TrimLeft(trimmed, "/")
		return Path2(trimmed)
	} else {
		panic(fmt.Errorf("path %s is not a descendant of %s", p, another))
	}
}

func (p Path2) Dir() Folder {
	return NewFolder(filepath.Dir(p.String()))
}

func (p Path2) ChMod(mod os.FileMode) {
	err := os.Chmod(p.String(), mod)
	if err != nil {
		panic(err)
	}
}
