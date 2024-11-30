package newfs

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func CWD() Folder {
	s, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	s = filepath.ToSlash(s)
	return NewFolder(s)
}

var Home Folder

func init() {
	homeS, _ := os.UserHomeDir()
	homeS = filepath.ToSlash(homeS)
	Home = NewFolder(homeS)
}

type Folder struct {
	Path2
}

func NewFolder(path string) Folder {
	return Folder{Path2(path)}
}

func (fd Folder) EnsureEmpty() {
	fd.EnsureExists()
	fd.DeleteDirContent()
}

func (fd Folder) EnsureExists() bool {
	if !fd.Exists() {
		fd.Create()
		return true
	}
	return false
}

func (fd Folder) Create() Folder {
	if !fd.Exists() {
		err := os.MkdirAll(fd.String(), 0755)
		if err != nil {
			panic(fmt.Errorf("Cannot create folder %s : %v\n", fd.String(), err))
		}
	}
	return fd
}

func (fd Folder) DeleteDirContent() *cmd.XbeeError {
	if !fd.Exists() {
		return nil
	}
	dir, err := os.Open(fd.String())
	if err != nil {
		return cmd.Error("cannot open %s : %v", fd, err)
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return cmd.Error("cannot read folder content %s : %v", fd, err)
	}
	for _, name := range names {
		child := fd.ChildFile(name).String()
		err = os.RemoveAll(child)
		if err != nil {
			return cmd.Error("cannot removeAll from %s : %v", child, err)
		}
	}
	return nil
}

func (fd Folder) Delete() *cmd.XbeeError {
	if fd.Exists() {
		if err := fd.DeleteDirContent(); err != nil {
			return err
		}
		if err := os.Remove(fd.String()); err != nil {
			return cmd.Error("cannot remove empty folder %s : %v", fd, err)
		}
	}
	return nil
}

func (fd Folder) ChildFile(name string) File {
	path := fd.Child(name)
	return File{Path2: path}
}
func (fd Folder) ChildFolder(name string) Folder {
	path := fd.Child(name)
	return Folder{Path2: path}
}
func (fd Folder) ChildFileJson(name string) File {
	return fd.ChildFile(name + jsonExt)
}
func (fd Folder) ChildFileYml(name string) File {
	return fd.ChildFile(name + YamlExt)
}
func (fd Folder) ChildFilesYml() (result []File) {
	return fd.ChildrenFilesEndingWith(YamlExt)
}
func (fd Folder) ChildrenFilesEndingWith(end string) (result []File) {
	theFiles, _ := fd.ChildrenFilesAndFolders()
	for _, child := range theFiles {
		if strings.HasSuffix(child.String(), end) {
			result = append(result, child)
		}
	}
	return
}
func (fd Folder) Empty() bool {
	if !fd.Exists() {
		return true
	}
	fis, err := os.ReadDir(fd.String())
	if err != nil {
		log.Panicf("Cannot not read children files for folder : %s : %v\n", fd, err)
	}
	return len(fis) == 0
}

func (fd Folder) ChildrenFilesAndFolders() (theFiles []File, theFolders []Folder) {
	if !fd.Exists() {
		return
	}
	fis, err := os.ReadDir(fd.String())
	if err != nil {
		panic(fmt.Errorf("Cannot not read children files for folder %s : %v\n", fd, err))
	}
	for _, fi := range fis {
		name := fi.Name()
		if !fi.IsDir() {
			theFiles = append(theFiles, fd.ChildFile(name))
		} else {
			theFolders = append(theFolders, fd.ChildFolder(name))
		}
	}
	return
}

func (fd Folder) Children() (result []Path2) {
	fis, err := os.ReadDir(fd.String())
	if err != nil {
		panic(fmt.Errorf("Cannot not read children files for folder %s : %v\n", fd, err))
	}
	for _, fi := range fis {
		name := fi.Name()
		result = append(result, fd.Child(name))
	}
	return
}

func (fd Folder) ChildrenFilesAndFoldersRelativePaths() (result []string) {
	theFiles, theFolders := fd.ChildrenFilesAndFolders()
	for _, f := range theFiles {
		s := strings.TrimPrefix(f.String(), fd.String())
		s = strings.TrimPrefix(s, "/")
		result = append(result, s)
	}
	for _, f := range theFolders {
		s := strings.TrimPrefix(f.String(), fd.String())
		s = strings.TrimPrefix(s, "/")
		result = append(result, s)
	}
	return
}

func (fd Folder) ChildPath(name string) Path2 {
	return fd.Child(name)
}

func (fd Folder) SubFolderForLocation(location string) (Folder, string) {
	anUrl := Parse(location)
	host, path, name := anUrl.Split()
	return fd.ChildFolder(host).ChildFolder(path), name
}

func (fd Folder) ResolvePath(s string) string {
	if s == "" {
		return fd.String()
	}
	if strings.HasPrefix(s, "~") {
		s = Home.String() + s[1:]
	} else if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "..") {
		s = fd.ChildFolder(s).String()
	} else if strings.HasPrefix(s, "$HOME") {
		s = Home.ChildFile(strings.TrimPrefix(s, "$HOME")).String()
	} else {
		return s
	}
	return filepath.Clean(s)
}