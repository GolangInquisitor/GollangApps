// client project client.go
package mclient

import (
	"errors"
	"mtcpserver"
	"net"
	"time"
)

type ReadHandler func(sock *net.Conn, buf []byte, size int, errorop error)
type WriteHandler func(sock *net.Conn, buf []byte, size int, errorop error)
type RClient = interface {
	WriteBufSize() int
	ReadBufSize() int
	SetReadBufSize(size int)
	SetWriteBufSize(size int)
	Write(byte_buf []byte, writefunc WriteHandler)
	Read(readfunc ReadHandler)
	SetErrorHandler(errhandler mtcpserver.RNetError)
	CloseConection()
	SetReadDeadLine(rdtime time.Duration)
	SetWriteDeadLine(wrtime time.Duration)
	CanWork() bool
	Start() error
	Stop()
	CreateConnection(sock net.Conn)
}

func NewClientHelper(mclient RClient) RClientWorker {
	return RClientWorker{client: mclient}
}

type RClientWorker struct {
	client RClient
}

func (worker *RClientWorker) Start(ipportstr string, read ReadHandler) error {

	conn, res := net.Dial("tcp", ipportstr)
	if res != nil {
		conn = nil
		return res
	}
	worker.client.CreateConnection(conn)
	if res = worker.client.Start(); res != nil {
		return res
	}
	worker.client.Read(read)
	return nil
}
func (worker *RClientWorker) SetErrorHandler(ehand mtcpserver.RNetError) {
	worker.client.SetErrorHandler(ehand)
}
func (worker *RClientWorker) SetReadDeadLine(rtimeduration time.Duration) {
	worker.client.SetReadDeadLine(rtimeduration)
}
func (worker *RClientWorker) SetWriteDeadLine(wtimeduration time.Duration) {
	worker.client.SetWriteDeadLine(wtimeduration)
}
func (c *RClientWorker) Write(byte_buf []byte, writefunc WriteHandler) {
	c.client.Write(byte_buf, writefunc)
}
func (c *RClientWorker) SetWriteBufSize(size int) {
	c.client.SetWriteBufSize(size)
}
func (c *RClientWorker) SetReadBufSize(size int) {
	c.client.SetReadBufSize(size)
}

type RclientHelper struct {
	sock       net.Conn
	wbufsize   int
	rbufsize   int
	rdtime     time.Duration
	wrtime     time.Duration
	ErrHandler mtcpserver.RNetError
	canwork    bool
}

func (c *RclientHelper) WriteBufSize() int {
	return c.wbufsize
}
func (c *RclientHelper) ReadBufSize() int {
	return c.rbufsize
}
func (c *RclientHelper) SetReadBufSize(size int) {
	c.rbufsize = size
}
func (c *RclientHelper) SetWriteBufSize(size int) {
	c.wbufsize = size
}
func (c *RclientHelper) Write(byte_buf []byte, writefunc WriteHandler) {
	if c.sock == nil {
		go c.ErrHandler.CreateError(&c.sock, 0, mtcpserver.ErrEmtyConn)
		return
	}
	if !c.canwork {
		go c.ErrHandler.CreateError(&c.sock, 0, errors.New("Client Stoped"))
		return
	}
	var res error
	var sendedlen int = 0
	if c.wbufsize >= len(byte_buf) {
		if res = c.sock.SetWriteDeadline(time.Now().Add(time.Second * c.wrtime)); res != nil {
			go c.ErrHandler.CreateError(&c.sock, 0, res)
			return
		}
		sendedlen, res = c.sock.Write(byte_buf)
	} else {
		res = mtcpserver.ErrOverSizeBuf
	}
	writefunc(&c.sock, byte_buf, sendedlen, res)
}
func (c *RclientHelper) Read(readfunc ReadHandler) {
	input := make([]byte, c.rbufsize)
	if res := c.sock.SetReadDeadline(time.Now().Add(time.Second * c.rdtime)); res != nil {
		go c.ErrHandler.CreateError(&c.sock, 0, res)
		return
	}
	for c.canwork {
		n, err := c.sock.Read(input)
		go readfunc(&c.sock, input, n, err)
	}

}
func (c *RclientHelper) SetErrorHandler(errhandler mtcpserver.RNetError) {
	c.ErrHandler = errhandler
}
func (c *RclientHelper) CloseConection() {
	c.sock.Close()
}
func (c *RclientHelper) Start() error {
	if c.ErrHandler == nil {
		return errors.New("Empty ErrorHandler")
	}
	c.canwork = true
	return nil
}
func (c *RclientHelper) Stop() {
	c.canwork = false
}
func (c *RclientHelper) CanWork() bool {
	return c.canwork
}
func (c *RclientHelper) SetReadDeadLine(rdtime time.Duration) {
	c.rdtime = rdtime
}
func (c *RclientHelper) SetWriteDeadLine(wrtime time.Duration) {
	c.wrtime = wrtime
}
func (c *RclientHelper) CreateConnection(sock net.Conn) {
	c.sock = sock
}
