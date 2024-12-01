package newfs

import "path/filepath"

func (p Path2) String() string {
	return filepath.ToSlash(string(p))
}
