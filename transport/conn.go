package transport

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/yushimeng/rock/sip"
	// "github.com/yushimeng/rock/sip"
)

type Connection interface {
	// WriteMsg marshals message and sends to socket
	WriteMsg(msg sip.Message) error
	// Reference of connection can be increased/decreased to prevent closing to earlyss
	Ref(i int)
	// Close decreases reference and if ref = 0 closes connection. Returns last ref. If 0 then it is closed
	TryClose() (int, error)
}

var bufPool = sync.Pool{
	New: func() interface{} {
		// The Pool's New function should generally only return pointer
		// types, since a pointer can be put into the return interface
		// value without an allocation:
		b := new(bytes.Buffer)
		// b.Grow(2048)
		return b
	},
}

type conn struct {
	net.Conn
}

func (c *conn) Ref(i int) {
	// Not used so far
}

func (c *conn) TryClose() (int, error) {
	// Not used so far
	return 0, c.Conn.Close()
}

func (c *conn) String() string {
	return c.LocalAddr().Network() + ":" + c.LocalAddr().String()
}

func (c *conn) WriteMsg(msg sip.Message) error {
	return c.WriteMsgTo(msg, msg.Destination())
}

func (c *conn) WriteMsgTo(msg sip.Message, raddr string) error {
	buf := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(buf)
	buf.Reset()
	msg.StringWrite(buf)
	data := buf.Bytes()

	n, err := c.Write(data)
	if err != nil {
		return fmt.Errorf("conn %s write err=%w", c, err)
	}

	if n == 0 {
		return fmt.Errorf("wrote 0 bytes")
	}

	if n != len(data) {
		return fmt.Errorf("fail to write full message")
	}
	return nil
}
