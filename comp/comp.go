package comp

import (
	"os/user"
	"path/filepath"
)

func MustBeNil(err error) {
	if err != nil {
		panic(err)
	}
}

func Expanduser(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path[:2] == "~/" {
		path = filepath.Join(dir, path[2:])
	}
	return path
}
