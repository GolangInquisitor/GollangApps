// mtcpserver project mtcpserver.go
package mtcpserver

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	ErrListLength   = errors.New("Error Length list")
	ErrFullConList  = errors.New("List is full")
	ErrNullByteRead = errors.New("Connection Null byte read")
	ErrEmtyConn     = errors.New("Attemt send to empty conn")
	ErrOverSizeBuf  = errors.New("Buffer Oversize")
)

type RNetError = interface {
	CreateError(conn *net.Conn, ConnTag int, err error)
}
type ConectItem struct {
	ConnTag    int
	Empty      bool
	Connection net.Conn
}
type NetList struct {
	connList      []ConectItem
	numActiveConn int
	lock          sync.RWMutex
}

func (list *NetList) SetNumConnections(numconn int) error {
	if numconn < 0 {

		return ErrListLength
	}
	list.connList = make([]ConectItem, numconn)
	list.numActiveConn = 0
	for key, _ := range list.connList {
		list.connList[key].Empty = true
	}
	return nil
}
func (list *NetList) AddConnection(sock net.Conn) (ConectItem, error) {
	var l = len(list.connList)
	if list.numActiveConn >= l /* len(list.connList)*/ {
		return *new(ConectItem), ErrFullConList
	}
	for key, _ := range list.connList {
		if list.connList[key].Empty {
			list.lock.Lock()
			defer list.lock.Unlock()
			list.connList[key].Connection = sock
			list.connList[key].ConnTag = key
			list.connList[key].Empty = false
			list.numActiveConn++
			return list.connList[key], nil
		}
	}
	return *new(ConectItem), ErrFullConList
}
func (list *NetList) GetConnetionByTag(tag int) ConectItem {
	list.lock.RLock()
	defer list.lock.RUnlock()
	for key, value := range list.connList {
		if (value.ConnTag == tag) && (!value.Empty) {
			return list.connList[key]
		}
	}
	return ConectItem{Empty: true, ConnTag: -1, Connection: nil}
}
func (list *NetList) RemoveConnByConTag(contag int) error {
	for _, value := range list.connList {
		if value.ConnTag == contag {
			list.lock.Lock()
			defer list.lock.Unlock()
			value.ConnTag = 0
			value.Empty = true
			list.numActiveConn--
			cl := value.Connection
			return cl.Close()
		}
	}
	return errors.New("Error remove connection by Tag")
}
func (list *NetList) RemoveConnByIndex(index int) error {
	for key, value := range list.connList {
		if key == index {
			list.lock.Lock()
			defer list.lock.Unlock()
			value.ConnTag = 0
			value.Empty = true
			list.numActiveConn--
			cl := value.Connection
			return cl.Close()
		}
	}
	return errors.New("Error remove connection by Index")
}

type ReadHandler func(sock ConectItem, buf []byte, size int, errorop error)
type WriteHandler func(sock ConectItem, buf []byte, size int, errorop error)

type RListenTCP interface {
	ListenHandler(sock net.Conn, readfunc ReadHandler, err error)
	WriteBufSize() int
	ReadBufSize() int
	SetReadBufSize(size int)
	SetWriteBufSize(size int)
	Write(ConTag int, writefunc WriteHandler, byte_buf []byte)
	SetErrorHandler(errhandler RNetError)
	SetNumConnections(numcon int)
}

func NewTcpWorker(worker RListenTCP) NET_TCPWorker {
	return NET_TCPWorker{tcpserver: worker}
}

type NET_TCPWorker struct {
	tcpserver RListenTCP
}

func (tserver *NET_TCPWorker) ReadBufSize() int {
	return tserver.tcpserver.ReadBufSize()
}
func (tserver *NET_TCPWorker) WriteBufSize() int {
	return tserver.tcpserver.WriteBufSize()
}
func (tserver *NET_TCPWorker) SetReadBufSize(size int) {
	tserver.tcpserver.SetReadBufSize(size)
}
func (tserver *NET_TCPWorker) SetWriteBufSize(size int) {
	tserver.tcpserver.SetWriteBufSize(size)
}
func (tserver *NET_TCPWorker) SetErrorHandler(errhandler RNetError) {
	tserver.tcpserver.SetErrorHandler(errhandler)
}
func (tserver *NET_TCPWorker) SetNumConnections(numcon int) {
	tserver.tcpserver.SetNumConnections(numcon)
}
func (tserver *NET_TCPWorker) Start(port string, readfunc ReadHandler) error {
	listener, err := net.Listen("tcp", port)
	defer listener.Close()
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		tserver.tcpserver.ListenHandler(conn, readfunc, err)

	}

	return err
}
func (tserver *NET_TCPWorker) Write(ConTag int, writefunc WriteHandler, byte_buf []byte) {
	go tserver.tcpserver.Write(ConTag, writefunc, byte_buf)
}

type RTCPhelper struct {
	ConectionList NetList
	ErrHandler    RNetError
	readbufsize   int
	writebufsize  int
	readtimeout   time.Time
	writetimeout  time.Time
}

func (serv *RTCPhelper) ListenHandler(sock net.Conn, readfunc ReadHandler, err error) {
	if serv.ErrHandler != nil {
		var res error
		if err == nil {
			var citem ConectItem
			if citem, res = serv.ConectionList.AddConnection(sock); res != nil {
				go serv.ErrHandler.CreateError(&sock, 0, res)
				return
			}
			if res = sock.SetWriteDeadline(serv.writetimeout); res != nil {
				go serv.ErrHandler.CreateError(&sock, 0, res)
				return
			}
			if res = sock.SetReadDeadline(serv.readtimeout); res != nil {
				go serv.ErrHandler.CreateError(&sock, 0, res)
				return
			}
			fmt.Println("New conect : " + sock.RemoteAddr().String())
			go serv.read(citem, readfunc)
		} else {
			go serv.ErrHandler.CreateError(&sock, 0, err)
		}
	} else {
		fmt.Println("ErrHandler interface is nil")
	}
}
func (serv *RTCPhelper) read(conn ConectItem, readfunc ReadHandler) {
	defer conn.Connection.Close()
	input := make([]byte, serv.readbufsize)
	if conn.Empty {
		fmt.Println("empty")
	}
	for !conn.Empty {
		n, err := conn.Connection.Read(input)
		if n == 0 {
			err = ErrNullByteRead
		} else {
			//fmt.Println("NewData ;" + string(input))
			go readfunc(conn, input, n, err)
		}
	}
	return
}
func (serv *RTCPhelper) SetReadDeadLine(rdtime time.Time) {
	serv.readtimeout = rdtime
}
func (serv *RTCPhelper) SetWriteDeadLine(wrtime time.Time) {
	serv.writetimeout = wrtime
}
func (serv *RTCPhelper) WriteBufSize() int {
	return serv.writebufsize
}
func (serv *RTCPhelper) ReadBufSize() int {
	return serv.readbufsize
}
func (serv *RTCPhelper) SetReadBufSize(size int) {
	serv.writebufsize = size
}
func (serv *RTCPhelper) SetWriteBufSize(size int) {
	serv.readbufsize = size
}
func (serv *RTCPhelper) Write(ConTag int, writefunc WriteHandler, byte_buf []byte) {
	var res error
	var sendedlen int = 0
	var sock ConectItem
	if serv.writebufsize >= len(byte_buf) {
		sock = serv.ConectionList.GetConnetionByTag(ConTag)
		if !sock.Empty {
			sendedlen, res = sock.Connection.Write(byte_buf)
		} else {
			res = ErrEmtyConn
		}
	} else {
		res = ErrOverSizeBuf
	}
	writefunc(sock, byte_buf, sendedlen, res)
}
func (serv *RTCPhelper) SetErrorHandler(errhandler RNetError) {
	serv.ErrHandler = errhandler
}
func (serv *RTCPhelper) SetNumConnections(numcon int) {
	serv.ConectionList.SetNumConnections(numcon)
}

/*func R_StartListener(port string, lsnrhandlr RServTCP) error {
	listener, err := net.Listen("tcp", port)

	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		lsnrhandlr.ListenHandler(conn, err)

	}
	defer listener.Close()
	return err
}
func R_Read(conect ConectItem, lsnrhandlr RServTCP) error {
	return lsnrhandlr.ReadHandler(conect, lsnrhandlr.ReadFunc)

}

type TCP_Server struct {
	tcp_conections NetList
}

func (mServer *TCP_Server) ListenHandler(sock net.Conn, err error) {
	if !mServer.ErrorHandler(sock, err) {
		err := mServer.tcp_conections.AddConnection(&sock)
		if err != nil {

		}
	}

}
func (mServer *TCP_Server) ReadFunc(sock ConectItem, buf []byte, errorop error) {

}
func (mServer *TCP_Server) ReadHandler(conect ConectItem, readerfunc ReaderHandler) error {

}

func (mServer *TCP_Server) ErrorHandler(sock net.Conn, err error) bool {
	if err != nil {
		log.Println("SockADDR :" + sock.RemoteAddr().String() + "ListenHandler errr:" + err.Error())
		return true
	}
	return false
}

var RTcp_Server TCP_Server

/*func (mServer *TCP_Server) StartListener(port string) (res error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("error accepting connection %v", err)
			continue
		}
		log.Printf("accepted connection from %v", conn.RemoteAddr())

	}
	return nil
}
func (mServer *TCP_Server) readConnReadEventHandler(conn net.Conn) {
	defer conn.Close()
	var CanWork bool = true
	r := bufio.NewReader(conn)
	scanr := bufio.NewScanner(r)
	for CanWork {
		conn.Read()
	}
}

func (mServer *TCP_Server) SetupReaderHandler(readerFunc ReaderHandler) {
	mServer.readFunc = readerFunc
}*/
