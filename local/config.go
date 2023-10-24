package local

import (
	"os"

	"github.com/pkg/errors"

	"github.com/liruonian/socks5"
)

type Config struct {
	RemoteAddress string `json:"remote_address"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

func (c *Config) ReadFrom(configFilePath string) error {
	return socks5.ReadConfig(c, configFilePath)
}

func (c *Config) WriteTo(configFilePath string) error {
	return socks5.WriteConfig(c, configFilePath)
}

func (c *Config) PrettyPrint(out *os.File) error {
	return socks5.PrettyPrint(c, out)
}

func (c *Config) Precheck() error {
	if len(c.RemoteAddress) == 0 {
		return errors.New("Remote address should not be nil")
	}

	if c.Port < 1024 {
		return errors.New("Port must be greater than 1024")
	}
	return nil
}
