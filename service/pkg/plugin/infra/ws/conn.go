package ws

import (
	"bytes"
	"github.com/gorilla/websocket"
	"net"
	"time"
)

type conn struct {
	conn       *websocket.Conn
	readBuffer bytes.Buffer
}

func newConn(wsc *websocket.Conn) *conn {
	return &conn{
		conn: wsc,
	}
}

func (c *conn) Read(b []byte) (n int, err error) {
	if c.readBuffer.Len() > 0 {
		return c.readBuffer.Read(b)
	}
	_, msg, err := c.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	n = copy(b, msg)
	if n < len(msg) {
		c.readBuffer.Write(msg[n:])
	}
	return n, nil

}
func (c *conn) Write(b []byte) (n int, err error) {
	return len(b), c.conn.WriteMessage(websocket.BinaryMessage, b)
}
func (c *conn) Close() error {
	return c.conn.Close()
}
func (c *conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}
func (c *conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}
func (c *conn) SetDeadline(t time.Time) error {
	err := c.conn.SetReadDeadline(t)
	if err != nil {
		return err
	}
	return c.conn.SetWriteDeadline(t)
}
func (c *conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}
func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
