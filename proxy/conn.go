package proxy

import "net"

type SecureConn struct {
	Conn net.Conn
}

func (c *SecureConn) Write(buf []byte) (n int, err error) {
	// TODO
	return c.Conn.Write(buf)
}

func (c *SecureConn) Read(buf []byte) (n int, err error) {
	// TODO
	return c.Conn.Read(buf)
}

func (c *SecureConn) Close() {
	_ = c.Conn.Close()
}
