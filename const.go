package socks5

import (
	"fmt"
	"path"

	"github.com/mitchellh/go-homedir"
)

const (
	name = "socks5"

	ServerSideName = name + "-server"
	LocalSideName  = name + "-local"

	Perm0644 = 0644

	Tcp = "tcp"

	BufferSize = 1024
)

var (
	HomePath, _          = homedir.Dir()
	ServerSideConfigPath = path.Join(HomePath, fmt.Sprintf(".%s.json", ServerSideName))
	LocalSideConfigPath  = path.Join(HomePath, fmt.Sprintf(".%s.json", LocalSideName))
	ServerSidePidPath    = path.Join(HomePath, fmt.Sprintf(".%s.pid", ServerSideName))
	LocalSidePidPath     = path.Join(HomePath, fmt.Sprintf(".%s.pid", LocalSideName))
)
