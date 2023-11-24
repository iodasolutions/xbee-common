package cmd

import (
	"fmt"
	"os"
	"os/user"
)

const (
	xbee_UID = "XBEE_UID"
	xbee_GID = "XBEE_GID"
)

func UserID() string {
	return os.Getenv(xbee_UID)
}
func UserIDAsEnv() string {
	if u, err := user.Current(); err != nil {
		panic(Error("unexpected error when trying to get user id : %v", err))
	} else {
		return fmt.Sprintf("%s=%s", xbee_UID, u.Uid)
	}
}
func GroupID() string {
	return os.Getenv(xbee_GID)
}
func GroupIDAsEnv() string {
	if u, err := user.Current(); err != nil {
		panic(Error("unexpected error when trying to get group id : %v", err))
	} else {
		return fmt.Sprintf("%s=%s", xbee_GID, u.Gid)
	}
}
