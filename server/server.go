package server

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/pkg/errors"

	"github.com/liruonian/socks5/server/auth"

	"github.com/liruonian/socks5/proxy"

	"github.com/liruonian/socks5"

	"github.com/sirupsen/logrus"
)

var (
	addressTypeNotSupportedError = errors.New("Address type not supported")
)

type server struct {
	config               *Config
	listener             net.Listener
	supportedAuthMethods map[uint8]auth.Authenticator
}

var singleton *server
var once sync.Once

func InitServer(config *Config) {
	once.Do(func() {
		// 将server初始化为单例
		singleton = &server{config: config}

		// 基于配置，判断当前server端支持的socks5的认证模式
		singleton.supportedAuthMethods = make(map[uint8]auth.Authenticator)
		singleton.supportedAuthMethods[auth.NoAuthenticationMethod] = &auth.NoAuthenticator{}
		if len(config.Username) != 0 && len(config.Password) != 0 {
			singleton.supportedAuthMethods[auth.UsernamePasswordAuthenticationMethod] = &auth.UsernamePasswordAuthenticator{}
		}

		// 记录当前进程的pid，当执行stop命令时，向该pid发送sigterm信号
		_ = socks5.RecordPid(socks5.ServerSidePidPath)

		// 初始化连接代理的buffer池
		proxy.InitProxy()
	})
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

	// 监听kill信号，用于graceful shutdown
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go s.waitingSignal(channel, cancel)

	// 开始监听连接到代理服务器的流量
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
			logrus.Infof("Stopping socks5 server service...")
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

func (s *server) handle(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()
	reader := bufio.NewReader(conn)

	// 协商socks版本
	version := []byte{0}
	if _, err := reader.Read(version); err != nil {
		logrus.Errorf("Error occoured while get version byte: %s", err.Error())
		return
	}
	if version[0] != socks5Version {
		logrus.Errorf("Unsupported socks version, expect %v, get %v", socks5Version, version[0])
		return
	}

	// 协商认证机制，如果未匹配到合适的认证方式，则采用无认证模式
	nmethods := []byte{0}
	if _, err := reader.Read(nmethods); err != nil {
		logrus.Errorf("Error occoured while get method number bytes: %s", err.Error())
		return
	}
	methods := make([]byte, nmethods[0])
	_, err := io.ReadAtLeast(reader, methods, int(nmethods[0]))
	if err != nil {
		logrus.Errorf("Error occoured while get method bytes: %s", err.Error())
		return
	}
	var authenticator auth.Authenticator
	for _, method := range methods {
		if item, exist := s.supportedAuthMethods[method]; exist {
			authenticator = item
		}
	}
	if authenticator == nil {
		authenticator = s.supportedAuthMethods[auth.NoAuthenticationMethod]
	}
	switch authenticator.GetMethod() {
	case auth.NoAuthenticationMethod:
		err := s.noAuthNegotiation(authenticator, conn)
		if err != nil {
			logrus.Errorf("Error occoured while initial socks connection setup: %s", err.Error())
			return
		}
	case auth.UsernamePasswordAuthenticationMethod:
		err := s.usernamePasswordNegotiation(authenticator, reader, conn)
		if err != nil {
			logrus.Errorf("Error occoured while initial socks connection setup: %s", err.Error())
			return
		}
	}

	// 解析本次请求类型
	request, err := s.newRequest(reader)
	if err != nil && err != addressTypeNotSupportedError {
		logrus.Errorf("Error occoured while process request: %s", err.Error())
		return
	} else if err == addressTypeNotSupportedError {
		if err := sendReply(conn, addressTypeNotSupported, nil); err != nil {
			logrus.Errorf("Address not supported error: %s", err.Error())
			return
		}
	}
	if request == nil {
		return
	}
	if client, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		request.RemoteAddr = &AddrSpec{IP: client.IP, Port: client.Port}
	}

	// 处理请求
	if err := s.handleRequest(request, conn); err != nil {
		logrus.Errorf("Error occoured while handle request: %s", err.Error())
		return
	}

}

func (s *server) usernamePasswordNegotiation(authenticator auth.Authenticator, reader *bufio.Reader, writer io.Writer) error {
	// 首先告知客户端，采用USERNAME/PASSWORD的方式进行认证
	_, err := writer.Write([]byte{socks5Version, authenticator.GetMethod()})
	if err != nil {
		return err
	}

	// 读取header部分，包含VER & ULEN，参考RFC 1929
	header := []byte{0, 0}
	if _, err := io.ReadAtLeast(reader, header, 2); err != nil {
		return err
	}

	// 确认认证协议的版本一致
	if header[0] != authVersion {
		return errors.New(fmt.Sprintf("Unsupported auth version, expect %v, get %v", authVersion, header[0]))
	}

	// 读取用户名
	ulen := int(header[1])
	username := make([]byte, ulen)
	if _, err := io.ReadAtLeast(reader, username, ulen); err != nil {
		return err
	}

	// 读取密码，复用header字节数组
	if _, err := reader.Read(header[:1]); err != nil {
		return err
	}
	plen := int(header[0])
	password := make([]byte, plen)
	if _, err := io.ReadAtLeast(reader, password, plen); err != nil {
		return err
	}

	// 校验认证结果，并写回给客户端
	err = authenticator.Authenticate(auth.Authentication{
		Principle:   string(username),
		Credentials: string(password),
	}, auth.Authentication{
		Principle:   s.config.Username,
		Credentials: s.config.Password,
	})
	if err != nil {
		_, err := writer.Write([]byte{authVersion, authFailure})
		return err
	}
	_, err = writer.Write([]byte{authVersion, authSuccess})
	return err
}

func (s *server) noAuthNegotiation(authenticator auth.Authenticator, writer io.Writer) error {
	_, err := writer.Write([]byte{socks5Version, authenticator.GetMethod()})
	return err
}

func (s *server) newRequest(reader *bufio.Reader) (*Request, error) {
	header := []byte{0, 0, 0}
	if _, err := io.ReadAtLeast(reader, header, 3); err != nil {
		logrus.Errorf("Error occoured while get request bytes: %s", err.Error())
		return nil, err
	}

	if header[0] != socks5Version {
		return nil, errors.New(fmt.Sprintf("Unsupported socks5 version, expect %v, get %v", socks5Version, header[0]))
	}

	// 根据不同的地址类型，读取目的地址信息
	dest := &AddrSpec{}
	atyp := []byte{0}
	if _, err := reader.Read(atyp); err != nil {
		return nil, err
	}
	switch atyp[0] {
	case ipv4Address:
		addr := make([]byte, 4)
		if _, err := io.ReadAtLeast(reader, addr, len(addr)); err != nil {
			return nil, err
		}
		dest.IP = net.IP(addr)
	case ipv6Address:
		addr := make([]byte, 16)
		if _, err := io.ReadAtLeast(reader, addr, len(addr)); err != nil {
			return nil, err
		}
		dest.IP = net.IP(addr)
	case fqdnAddress:
		if _, err := reader.Read(atyp); err != nil {
			return nil, err
		}
		addrLen := int(atyp[0])
		fqdn := make([]byte, addrLen)
		if _, err := io.ReadAtLeast(reader, fqdn, addrLen); err != nil {
			return nil, err
		}
		dest.FQDN = string(fqdn)
	default:
		return nil, addressTypeNotSupportedError
	}

	// 获取端口信息
	port := []byte{0, 0}
	if _, err := io.ReadAtLeast(reader, port, 2); err != nil {
		return nil, err
	}
	dest.Port = (int(port[0]) << 8) | int(port[1])

	return &Request{
		Version:  socks5Version,
		Command:  header[1],
		DestAddr: dest,
		reader:   reader,
	}, nil
}

func sendReply(w io.Writer, resp uint8, addr *AddrSpec) error {
	var addrType uint8
	var addrBody []byte
	var addrPort uint16
	switch {
	case addr == nil:
		addrType = ipv4Address
		addrBody = []byte{0, 0, 0, 0}
		addrPort = 0

	case addr.FQDN != "":
		addrType = fqdnAddress
		addrBody = append([]byte{byte(len(addr.FQDN))}, addr.FQDN...)
		addrPort = uint16(addr.Port)

	case addr.IP.To4() != nil:
		addrType = ipv4Address
		addrBody = []byte(addr.IP.To4())
		addrPort = uint16(addr.Port)

	case addr.IP.To16() != nil:
		addrType = ipv6Address
		addrBody = []byte(addr.IP.To16())
		addrPort = uint16(addr.Port)

	default:
		return errors.New(fmt.Sprintf("Failed to format address: %v", addr))
	}

	msg := make([]byte, 6+len(addrBody))
	msg[0] = socks5Version
	msg[1] = resp
	msg[2] = 0
	msg[3] = addrType
	copy(msg[4:], addrBody)
	msg[4+len(addrBody)] = byte(addrPort >> 8)
	msg[4+len(addrBody)+1] = byte(addrPort & 0xff)

	_, err := w.Write(msg)
	return err
}

func (s *server) handleRequest(request *Request, conn net.Conn) error {
	ctx := context.Background()

	dest := request.DestAddr
	if dest.FQDN != "" {
		ctx_, addr, err := s.resolve(ctx, dest.FQDN)
		if err != nil {
			if err := sendReply(conn, hostUnreachable, nil); err != nil {
				return err
			}
			return err
		}
		ctx = ctx_
		dest.IP = addr
	}

	switch request.Command {
	case ConnectCommand:
		return s.handleConnectRequest(conn, request)
	default:
		if err := sendReply(conn, commandNotSupported, nil); err != nil {
			return err
		}
	}

	return nil
}

func (s *server) resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	addr, err := net.ResolveIPAddr("ip", name)
	if err != nil {
		return ctx, nil, err
	}
	return ctx, addr.IP, err
}

func (s *server) handleConnectRequest(conn net.Conn, request *Request) error {
	target, err := net.Dial(socks5.Tcp, request.DestAddr.Address())
	if err != nil {
		msg := err.Error()
		resp := hostUnreachable
		if strings.Contains(msg, "refused") {
			resp = connectionRefused
		} else if strings.Contains(msg, "network is unreachable") {
			resp = networkUnreachable
		}
		if err := sendReply(conn, resp, nil); err != nil {
			return err
		}
		return err
	}
	defer func() {
		_ = target.Close()
	}()

	local := target.LocalAddr().(*net.TCPAddr)
	bind := AddrSpec{IP: local.IP, Port: local.Port}
	if err := sendReply(conn, succeeded, &bind); err != nil {
		return err
	}

	errCh := make(chan error, 2)
	go proxy.Proxy(target, request.reader, errCh)
	go proxy.Proxy(conn, target, errCh)

	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			return e
		}
	}
	return nil
}

func (s *server) StopServer() {
	socks5.Suicide(socks5.ServerSidePidPath)
}
