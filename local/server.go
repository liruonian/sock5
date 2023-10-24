package local

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/liruonian/socks5/proxy"

	"github.com/liruonian/socks5"

	"github.com/sirupsen/logrus"
)

type server struct {
	config   *Config
	listener net.Listener
	remote   *net.TCPAddr
}

var singleton *server
var once sync.Once

func InitServer(config *Config) {
	once.Do(func() {
		singleton = &server{config: config}
	})
	proxy.InitProxy()
	_ = socks5.RecordPid(socks5.LocalSidePidPath)
}

func GetServer() *server {
	return singleton
}

func (s *server) StartServer() {
	// 校验配置文件的参数，是否存在不合理的配置
	err := s.config.Precheck()
	if err != nil {
		logrus.Errorf("Invalid configuration: %s", err.Error())
		return
	}

	// 解析socks5服务端地址
	remote, err := net.ResolveTCPAddr(socks5.Tcp, s.config.RemoteAddress)
	if err != nil {
		logrus.Errorf("Invalid remote address[%s]: %s", s.config.RemoteAddress, err.Error())
		return
	}
	s.remote = remote

	// 监听kill信号，用于graceful shutdown
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go s.waitingSignal(channel, cancel)

	// 开始监听被代理到本地端口的流量
	s.listen(ctx)
}

func (s *server) waitingSignal(channel chan os.Signal, cancel context.CancelFunc) {
	<-channel
	cancel()
	if s.listener != nil {
		_ = s.listener.Close()
	}
}

func (s *server) listen(ctx context.Context) {
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, socks5.Tcp, fmt.Sprintf(":%v", s.config.Port))
	if err != nil {
		logrus.Infof("Error occured: %s", err.Error())
		return
	}
	s.listener = listener

	for {
		select {
		case <-ctx.Done():
			logrus.Infof("Stopping socks5 local service...")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				logrus.Errorf("Error occured while accept tcp: %s", err.Error())
				continue
			}

			go s.handle(conn)
		}
	}
}

func (s *server) handle(localConn net.Conn) {
	remoteConn, err := net.DialTCP(socks5.Tcp, nil, s.remote)
	if err != nil {
		logrus.Errorf("Error occured while dial remote addr[%s]: %s", s.config.RemoteAddress, err.Error())
		return
	}
	logrus.Infof("proxy chain %s -> %s -> %s -> %s",
		localConn.RemoteAddr().String(), localConn.LocalAddr().String(), remoteConn.LocalAddr().String(), remoteConn.RemoteAddr().String())

	errCh := make(chan error, 2)
	go proxy.Proxy(localConn, remoteConn, errCh)
	go proxy.Proxy(remoteConn, localConn, errCh)

	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			logrus.Errorf("Error occured: %s", e.Error())
			return
		}
	}

}

func (s *server) StopServer() {
	socks5.Suicide(socks5.LocalSidePidPath)
}
