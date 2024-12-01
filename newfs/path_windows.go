package newfs

import "path/filepath"

func (p Path) String() string {
	return filepath.ToSlash(string(p))
}
