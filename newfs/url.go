package newfs

import (
	"fmt"
	"net/url"
	"strings"
)

type Url struct {
	url.URL
}

func Parse(rawUrl string) *Url {
	if anURL, err := url.Parse(rawUrl); err != nil {
		panic(fmt.Errorf("Cannot parse url %s : %v", rawUrl, err))
	} else {
		return &Url{URL: *anURL}
	}
}

func (u *Url) LastPath() string {
	index := strings.LastIndex(u.Path, "/")
	if index != -1 {
		return u.Path[index+1:]
	} else {
		return u.Path
	}
}

func (u *Url) Split() (host string, path string, name string) {
	index := strings.LastIndex(u.Path, "/")
	host = u.Host
	if index != -1 {
		name = u.Path[index+1:]
		path = u.Path[:index+1]
	} else {
		name = u.Path
	}
	return
}
