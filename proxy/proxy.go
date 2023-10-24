package proxy

import (
	"io"
	"sync"

	"github.com/liruonian/socks5"
)

var pool sync.Pool
var once sync.Once

func InitProxy() {
	once.Do(func() {
		pool.New = func() interface{} {
			return make([]byte, socks5.BufferSize)
		}
	})
}

type closeWriter interface {
	CloseWrite() error
}

func Proxy(dst io.Writer, src io.Reader, errCh chan error) {
	_, err := io.Copy(dst, src)
	if tcpConn, ok := dst.(closeWriter); ok {
		_ = tcpConn.CloseWrite()
	}
	errCh <- err

}
