package main

import (
	"log"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/liruonian/socks5/local"

	"github.com/liruonian/socks5"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = socks5.LocalSideName

	app.Commands = []cli.Command{
		configCmd,
		startCmd,
		stopCmd,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var configCmd = cli.Command{
	Name:  "config",
	Usage: "View and modify socks5 local configuration",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "v",
			Usage: "Print socks5 local configuration",
		},
		cli.StringFlag{
			Name:  "r",
			Usage: "Remote socks5 server address. eg: example.com:15680",
		},
		cli.IntFlag{
			Name:  "p",
			Usage: "Port of local socks5, must be greater than 1024. eg: 15678",
		},
	},
	Action: func(context *cli.Context) {
		config := &local.Config{}

		// 获取原始配置
		err := config.ReadFrom(socks5.LocalSideConfigPath)
		if err != nil && err != socks5.ConfigFileNotExist {
			logrus.Errorf("Error occoured: %s", err.Error())
			return
		}

		// 如果请求查看配置文件，则忽略其他参数
		if context.Bool("v") {
			err := config.PrettyPrint(os.Stdout)
			if err != nil {
				logrus.Errorf("Error occoured: %s", err.Error())
				return
			}
			return
		}

		// 将其它配置项写入配置文件
		if len(context.String("r")) > 0 {
			config.RemoteAddress = context.String("r")
		}
		if context.Int("p") > 1024 {
			config.Port = context.Int("p")
		}
		err = config.WriteTo(socks5.LocalSideConfigPath)
		if err != nil {
			logrus.Errorf("Error occoured: %s", err.Error())
			return
		}

		logrus.Infof("Successful modification of the configuration file: %s", socks5.LocalSideConfigPath)
	},
}

var startCmd = cli.Command{
	Name:  "start",
	Usage: "StartServer socks5 local service",
	Action: func(context *cli.Context) {
		config := &local.Config{}

		err := config.ReadFrom(socks5.LocalSideConfigPath)
		if err != nil {
			logrus.Errorf("Error occoured: %s", err.Error())
			logrus.Errorf("Failed to read the configuration file, if the file does not exist, " +
				"please create it initially with the command: socks5-local config, " +
				"or check if the configuration file permissions can be accessed properly")
			return
		}

		logrus.Infof("Try to initialize socks local service...")
		local.InitServer(config)

		logrus.Infof("Starting socks5 local service...")
		local.GetServer().StartServer()
	},
}

var stopCmd = cli.Command{
	Name:  "stop",
	Usage: "StopServer socks5 local service",
	Action: func(context *cli.Context) {
		local.GetServer().StopServer()
	},
}
