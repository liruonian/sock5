package socks5

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
)

func RecordPid(pidPath string) error {
	err := ioutil.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), Perm0644)
	if err != nil {
		return errors.Wrapf(err, "Record pid failed: %v", os.Getpid())
	}

	return nil
}

func Suicide(pidPath string) {
	bytes, err := ioutil.ReadFile(pidPath)
	if err != nil {
		return
	}
	_ = exec.Command("kill", string(bytes)).Run()
}
