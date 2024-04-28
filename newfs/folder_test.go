package newfs

import (
	"testing"
)

func Test_TarToFolder(t *testing.T) {
	fdA := TmpDir().ChildFolder("a").Create()
	fdA.ChildFile("b.txt").SetContent("b")
	fdA.ChildFile("c.txt").SetContent("c")
	if err := fdA.TarToFile("/tmp/a.tar"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

}

func Test_TarToFolderUsrBin(t *testing.T) {
	fdA := Folder("/home/eric/CLionProjects")
	if fdA.Exists() {
		if err := fdA.TarToFile("/tmp/a.tar"); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	} else {
		t.Skip("cannot run this test: folder /home/eric/CLionProjects do not exist")
	}

}
