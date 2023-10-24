package main

import (
	"log"
	"os"

	"github.com/liruonian/socks5/server"

	"github.com/sirupsen/logrus"

	"github.com/liruonian/socks5"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = socks5.ServerSideName

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
	Usage: "View and modify socks5 server configuration",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "v",
			Usage: "Print socks5 server configuration",
		},
		cli.IntFlag{
			Name:  "p",
			Usage: "Port of server socks5, must be greater than 1024. eg: 15678",
		},
		cli.StringFlag{
			Name:  "u",
			Usage: "Username for authentication",
		},
		cli.StringFlag{
			Name:  "P",
			Usage: "Password for authentication",
		},
	},
	Action: func(context *cli.Context) {
		config := &server.Config{}

		// 获取原始配置
		err := config.ReadFrom(socks5.ServerSideConfigPath)
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
		if context.Int("p") > 1024 {
			config.Port = context.Int("p")
		}
		if len(context.String("u")) > 0 {
			config.Username = context.String("u")
		}
		if len(context.String("P")) > 0 {
			config.Password = context.String("P")
		}
		err = config.WriteTo(socks5.ServerSideConfigPath)
		if err != nil {
			logrus.Errorf("Error occoured: %s", err.Error())
			return
		}

		logrus.Infof("Successful modification of the configuration file: %s", socks5.ServerSideConfigPath)
	},
}

var startCmd = cli.Command{
	Name:  "start",
	Usage: "StartServer socks5 server service",
	Action: func(context *cli.Context) {
		config := &server.Config{}

		err := config.ReadFrom(socks5.ServerSideConfigPath)
		if err != nil {
			logrus.Errorf("Error occoured: %s", err.Error())
			logrus.Errorf("Failed to read the configuration file, if the file does not exist, " +
				"please create it initially with the command: socks5-server config, " +
				"or check if the configuration file permissions can be accessed properly")
			return
		}

		logrus.Infof("Try to initialize socks server service...")
		server.InitServer(config)

		logrus.Infof("Starting socks5 server service...")
		server.GetServer().StartServer()
	},
}

var stopCmd = cli.Command{
	Name:  "stop",
	Usage: "StopServer socks5 server service",
	Action: func(context *cli.Context) {
		server.GetServer().StopServer()
	},
}
