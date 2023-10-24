package socks5

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

var (
	ConfigFileNotExist = errors.New("Config file not exist")
)

type Config interface {
	ReadFrom(configFilePath string) error
	WriteTo(configFilePath string) error
	PrettyPrint(out *os.File) error
	Precheck() error
}

func ReadConfig(config interface{}, configFilePath string) error {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return ConfigFileNotExist
	}

	configFile, err := os.Open(configFilePath)
	if err != nil {
		return errors.Wrapf(err, "Error occured while try to open config file[%s]", configFilePath)
	}
	defer func() {
		_ = configFile.Close()
	}()

	err = json.NewDecoder(configFile).Decode(config)
	if err != nil {
		return errors.Wrapf(err, "Incorrect json format[%s]", configFilePath)
	}

	return nil
}

func WriteConfig(config interface{}, configFilePath string) error {
	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "Json marshal failed: %s", err.Error())
	}

	err = ioutil.WriteFile(configFilePath, jsonBytes, Perm0644)
	if err != nil {
		return errors.Wrapf(err, "Write config file[%s] failed: %s", configFilePath, err.Error())
	}

	return nil
}

func PrettyPrint(config interface{}, out *os.File) error {
	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "Json marshal failed: %s", err.Error())
	}

	_, err = out.Write(jsonBytes)
	if err != nil {
		return errors.Wrapf(err, "Write config file to std failed: %s", err.Error())
	}

	return nil
}
