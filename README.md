# socks5
基于socks5的实现，参考 [RFC1928](https://www.ietf.org/rfc/rfc1928.txt)
- 支持socks5代理协议
- 支持无鉴权方式和基于用户名和密码的鉴权
- 暂仅支持CONNECT命令

## 1.准备
### 1.1 编译
要求linux系统以及golang 1.17+，编译完成后会得到socks的可执行文件，包括服务端`socks5-server`和客户端`socks5-local`。
```bash
$ git clone https://github.com/liruonian/socks5.git
$ cd socks5
$ go build -o socks5-server cmd/server/main.go
$ go build -o socks5-local cmd/local/main.go
```

## 2.使用
### 2.1 服务端
支持对服务端的配置、启停功能。
```bash
$ socks5-server -h
NAME:
   socks5-server - A new cli application

USAGE:
   socks5-server [global options] command [command options] [arguments...]

COMMANDS:
   config   View and modify socks5 server configuration
   start    StartServer socks5 server service
   stop     StopServer socks5 server service
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

首先对服务端进行初始配置，包括监听的端口，以及用于认证的用户名和密码。
```bash
$ socks5-server config -p 12345 -u liruonian -P liruonian
INFO[0000] Successful modification of the configuration file: /root/.socks5-server.json
```

启动服务端程序。
```bash
$ socks5-server start
INFO[0000] Try to initialize socks server service...
INFO[0000] Starting socks5 server service...
```

### 2.2 客户端
支持对客户端的配置、启停功能。
```bash
$ socks5-local -h                                                                  14:22:49
NAME:
   socks5-local - A new cli application

USAGE:
   socks5-local [global options] command [command options] [arguments...]

COMMANDS:
   config   View and modify socks5 local configuration
   start    StartServer socks5 local service
   stop     StopServer socks5 local service
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

结合上面对服务端的配置，修改客户端配置，假设服务器域名为`example.com`。
```bash
$ socks5-local config -r example.com:12345 -p 1111                                 14:25:07
INFO[0000] Successful modification of the configuration file: /Users/lihao/.socks5-local.json
```

启动客户端程序。
```bash
$ socks5-local start                                                               14:26:07
INFO[0000] Try to initialize socks local service...
INFO[0000] Starting socks5 local service...
```

### 2.3 配置本地代理
#### 2.3.1 CURL
```bash
curl -v google.com -x socks5://liruonian:liruonian@127.0.0.1:1111                                                14:28:18
*   Trying 127.0.0.1:1111...
* SOCKS5 connect to IPv4 172.217.160.110:80 (locally resolved)
* SOCKS5 request granted.
* Connected to 127.0.0.1 (127.0.0.1) port 1111 (#0)
> GET / HTTP/1.1
> Host: google.com
> User-Agent: curl/7.84.0
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 301 Moved Permanently
< Location: http://www.google.com/
< Content-Type: text/html; charset=UTF-8
< Date: Sun, 26 Feb 2023 06:28:47 GMT
< Expires: Tue, 28 Mar 2023 06:28:47 GMT
< Cache-Control: public, max-age=2592000
< Server: gws
< Content-Length: 219
< X-XSS-Protection: 0
< X-Frame-Options: SAMEORIGIN
<
<HTML><HEAD><meta http-equiv="content-type" content="text/html;charset=utf-8">
<TITLE>301 Moved</TITLE></HEAD><BODY>
<H1>301 Moved</H1>
The document has moved
<A HREF="http://www.google.com/">here</A>.
</BODY></HTML>
* Connection #0 to host 127.0.0.1 left intact
```

#### 2.3.2 浏览器
示例中使用Chrome浏览器，浏览器使用系统代理，再将系统代理配置为使用socks5协议。
![setup](https://p.ipic.vip/5f675s.png)

配置完成即可正常使用代理服务。
![google](https://p.ipic.vip/zyxemi.png)
