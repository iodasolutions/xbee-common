package newfs

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/stringutils"
	"os"
	"path/filepath"
	"strings"
)

type Path string

func ResolvePath(s string) string {
	return CWD().ResolvePath(s)
}

func CleanPath(s string) Path {
	if strings.HasPrefix(s, "~") {
		s = Home.String() + s[1:]
	}
	return Path(filepath.Clean(s))
}

func (p Path) Exists() bool {
	_, err := os.Stat(string(p))
	if err == nil {
		return true
	}
	return false
}

func (p Path) String() string {
	return string(p)
}

func (p Path) Child(name string) Path {
	return Path(filepath.Join(string(p), name))
}

func (p Path) BaseWithoutExtension() string {
	base := filepath.Base(string(p))
	index := strings.LastIndex(base, ".")
	if index != -1 {
		return base[:index]
	}
	return base
}

func (p Path) Extension() string {
	base := filepath.Base(string(p))
	index := strings.LastIndex(base, ".")
	if index != -1 {
		return base[index+1:]
	}
	return ""
}

func (p Path) Base() string {
	return filepath.Base(string(p))
}

func (p Path) Hash() string {
	return p.Base() + "-" + stringutils.Sha1StringTruncated(filepath.Dir(string(p)))
}
func (p Path) WithoutHash() Path {
	return Path(p.String()[:len(p.String())-11])
}

func (p Path) IsDir() bool {
	fi, err := os.Lstat(string(p))
	if err == nil {
		return fi.IsDir()
	}
	return false
}
func (p Path) IsLink() bool {
	fi, err := os.Lstat(string(p))
	if err == nil {
		return fi.Mode()&os.ModeSymlink != 0
	}
	return false
}

func (p Path) IsAbs() bool {
	return filepath.IsAbs(string(p))
}
func (p Path) IsRelative() bool {
	return !p.IsAbs()
}
func (p Path) IsDescendantOf(another Path) bool {
	return strings.HasPrefix(string(p), string(another))
}

func (p Path) IsSymLink() bool {
	fi, err := os.Stat(string(p))
	if err == nil {
		return fi.Mode()&os.ModeSymlink == os.ModeSymlink
	}
	return false
}
func (p Path) RelativeFrom(another Path) Path {
	if p.IsRelative() {
		panic(fmt.Errorf("method Path.RelativeFrom : path %s is not an absolute path", p))
	}
	if p.IsDescendantOf(another) {
		folderS := another.String()
		pS := p.String()
		trimmed := strings.TrimPrefix(pS, folderS)
		trimmed = strings.TrimLeft(trimmed, "/")
		return Path(trimmed)
	} else {
		panic(fmt.Errorf("path %s is not a descendant of %s", p, another))
	}
}

func (p Path) Sha1() string {
	if p.IsDir() {
		return Folder(p).Sha1()
	} else {
		return File(p).Sha1()
	}
}
func (p Path) Sha1Path() Path {
	return Path(fmt.Sprintf("%s.sha1", p))
}

func (p Path) CopyToDir(fd Folder) {
	if p.IsDir() {
		srcDir := Folder(p)
		srcDir.CopyDirToDir(fd)
	} else {
		f := File(p)
		f.CopyToDir(fd)
	}
}

func (p Path) AddRelative(path string) Path {
	s := filepath.Join(p.String(), path)
	return CleanPath(s)
}

func (p Path) Dir() Folder {
	return Folder(filepath.Dir(p.String()))
}
