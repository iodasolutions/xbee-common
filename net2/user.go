package net2

import (
	"fmt"
	"github.com/iodasolutions/xbee-common/cmd"
	"github.com/iodasolutions/xbee-common/newfs"
	"github.com/iodasolutions/xbee-common/util"
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
	if cmd.UserId != "" {
		return &User{
			uid: cmd.UserId,
			gid: cmd.GroupId,
		}
	}
	return nil
}

func (u *User) ChangeOwner(aPath string) *util.XbeeError {
	uid, _ := strconv.Atoi(u.Uid())
	gid, _ := strconv.Atoi(u.Gid())
	if err := os.Chown(aPath, uid, gid); err != nil {
		return util.Error(fmt.Sprintf("Cannot set ownership to %s:%s for file %s", u.Uid(), u.Gid(), aPath))
	}
	return nil
}
func (u *User) ChangeOwnerForAnscestorsOf(f newfs.File, root newfs.Folder) *util.XbeeError {
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
