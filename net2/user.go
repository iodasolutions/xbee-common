package net2

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"os"
	"strconv"
)

type User struct {
	uid string
	gid string
}

func (u *User) Uid() string { return u.uid }
func (u *User) Gid() string { return u.gid }

func HostUser() *User {
	user := cmd.UserID()
	if user != "" {
		return &User{
			uid: user,
			gid: cmd.GroupID(),
		}
	}
	return nil
}

func (u *User) ChangeOwner(aPath string) *cmd.XbeeError {
	uid, _ := strconv.Atoi(u.Uid())
	gid, _ := strconv.Atoi(u.Gid())
	if err := os.Chown(aPath, uid, gid); err != nil {
		return cmd.Error(fmt.Sprintf("Cannot set ownership to %s:%s for file %s", u.Uid(), u.Gid(), aPath))
	}
	return nil
}
func (u *User) ChangeOwnerForAnscestorsOf(f newfs.File, root newfs.Folder) *cmd.XbeeError {
	ancestor := f.Dir()
	ancestor.EnsureExists()
	for ancestor.String() != root.String() && ancestor != "/" {
		if err := u.ChangeOwner(ancestor.String()); err != nil {
			return err
		}
		ancestor = ancestor.Dir()
	}
	return nil
}
