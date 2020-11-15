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
	ErrListLength      = errors.New("Error Length list")
	ErrFullConList     = errors.New("List is full")
	ErrNullByteRead    = errors.New("Connection Null byte read")
	ErrEmtyConn        = errors.New("Attemt send to empty conn")
	ErrOverSizeBuf     = errors.New("Buffer Oversize")
	ErrRemoveByTag     = errors.New("Error remove connection by Tag Not Found Conection")
	ErrRemoveByIndx    = errors.New("Error remove connection by Index Not Found Conection")
	ErrNumConn         = errors.New("Error set connection numbers")
	ErrSetSizeBuf      = errors.New("Error set size buffer")
	ErrSetTimeOut      = errors.New("Error set timeout")
	ErrEmptyErrHandler = errors.New("Error Errorhandler is empty!!")
)

type TcpConfig struct {
	TCPport         string        `json:"tcpport"`
	NumConnetions   int           `json:"numconn"`
	ReadBufSize     int           `json:"readbufsize"`
	WriteBufSize    int           `json:"writebufsize"`
	TCPWriteTimeOut time.Duration `json:"tcpwritetime"`
	TCPReadTimeOut  time.Duration `json:"tcpreadtime"`
	ReadBufSizeTest bool          `json:"testreadbufsize"`
}

func CheckTcpConfig(conf TcpConfig) error {
	if conf.NumConnetions <= 0 {
		return ErrNumConn
	}
	if conf.ReadBufSize <= 0 || conf.WriteBufSize <= 0 {
		return ErrSetSizeBuf
	}
	if conf.TCPReadTimeOut <= 0 || conf.TCPWriteTimeOut <= 0 {
		return ErrSetTimeOut
	}
	return nil
}

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
	if list.numActiveConn >= len(list.connList) {
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
	for key, _ := range list.connList {
		if list.connList[key].ConnTag == contag {
			list.lock.Lock()
			defer list.lock.Unlock()
			list.connList[key].ConnTag = 0
			list.connList[key].Empty = true
			list.numActiveConn--
			return list.connList[key].Connection.Close()
		}
	}
	return ErrRemoveByTag
}
func (list *NetList) RemoveConnByIndex(index int) error {
	for key, _ := range list.connList {
		if key == index {
			list.lock.Lock()
			defer list.lock.Unlock()
			list.connList[key].ConnTag = 0
			list.connList[key].Empty = true
			list.numActiveConn--
			return list.connList[key].Connection.Close()
		}
	}
	return ErrRemoveByIndx
}
func (list *NetList) CloseAll() {
	for key, _ := range list.connList {
		list.connList[key].ConnTag = 0
		list.connList[key].Empty = true
		list.connList[key].Connection.Close()
	}
}

type ReadHandler func(sock ConectItem, buf []byte, size int, errorop error)
type WriteHandler func(sock ConectItem, buf []byte, size int, errorop error)
type AfterConnectHandler func(sock ConectItem)

type RListenTCP interface {
	ListenHandler(sock net.Conn, readfunc ReadHandler, err error)
	WriteBufSize() int
	ReadBufSize() int
	SetReadBufSize(size int)
	SetWriteBufSize(size int)
	Write(ConTag int, writefunc WriteHandler, byte_buf []byte)
	SetAfterConnectEvent(doafterconect AfterConnectHandler)
	SetErrorHandler(errhandler RNetError)
	SetNumConnections(numcon int)
	CloseConectionByTag(Tag int)
	SetReadDeadLine(rdtime time.Duration)
	SetWriteDeadLine(wrtime time.Duration)
	CloseConnections()
	Stop()
	Start() error
	CanWork() bool
}

func NewTcpWorker(worker RListenTCP) NET_TCPWorker {
	return NET_TCPWorker{tcpserver: worker}
}
func NewDefuaultTcpWorker() NET_TCPWorker {
	return NewTcpWorker(new(RTCPhelper))
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
func (tserver *NET_TCPWorker) SetReadDeadLine(rdtime time.Duration) {
	tserver.tcpserver.SetReadDeadLine(rdtime)
}
func (tserver *NET_TCPWorker) SetWriteDeadLine(wrtime time.Duration) {
	tserver.tcpserver.SetWriteDeadLine(wrtime)
}
func (tserver *NET_TCPWorker) Stop() {
	tserver.tcpserver.Stop()
}
func (tserver *NET_TCPWorker) CloseConnectionByTag(tag int) {
	tserver.tcpserver.CloseConectionByTag(tag)
}
func (tserver *NET_TCPWorker) SetAfterConnectEvent(doafterconect AfterConnectHandler) {
	tserver.tcpserver.SetAfterConnectEvent(doafterconect)
}
func (tserver *NET_TCPWorker) Start(port string, readfunc ReadHandler) error {
	a, err := net.ResolveTCPAddr("tcp", port)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", a)
	defer listener.Close()
	if err != nil {
		return err
	}
	if err := tserver.tcpserver.Start(); err != nil {
		return err
	}

	for tserver.tcpserver.CanWork() {
		listener.SetDeadline(time.Now().Add(time.Second * 10))
		conn, err := listener.Accept()
		tserver.tcpserver.ListenHandler(conn, readfunc, err)

	}
	tserver.tcpserver.CloseConnections()
	return nil
}
func (tserver *NET_TCPWorker) Write(ConTag int, writefunc WriteHandler, byte_buf []byte) {
	tserver.tcpserver.Write(ConTag, writefunc, byte_buf)
}

type RTCPhelper struct {
	ConectionList          NetList
	ErrHandler             RNetError
	readbufsize            int
	writebufsize           int
	readtimeout            time.Duration
	writetimeout           time.Duration
	stop                   bool
	needtestreadbufsize    bool
	startafterconnectevent AfterConnectHandler
}

func (serv *RTCPhelper) ListenHandler(sock net.Conn, readfunc ReadHandler, err error) {
	if serv.ErrHandler != nil {
		var res error
		if err == nil {
			var citem ConectItem
			if citem, res = serv.ConectionList.AddConnection(sock); res != nil {
				serv.ErrHandler.CreateError(&sock, 0, res)
				return
			}

			fmt.Println("New conect : " + sock.RemoteAddr().String())
			go serv.read(citem, readfunc)
			serv.startafterconnectevent(citem)
		} else {
			serv.ErrHandler.CreateError(&sock, 0, err)
		}
	} else {
		fmt.Println("ErrHandler interface is nil")
	}
}
func (serv *RTCPhelper) read(conn ConectItem, readfunc ReadHandler) {
	input := make([]byte, serv.readbufsize)
	if conn.Empty {
		fmt.Println("empty")
	}
	for !conn.Empty && !serv.stop {
		if res := conn.Connection.SetReadDeadline(time.Now().Add(time.Second * serv.readtimeout)); res != nil {
			serv.ErrHandler.CreateError(&conn.Connection, 999999999999999, res)

			continue
		}
		n, err := conn.Connection.Read(input)
		if serv.needtestreadbufsize {
			if n > serv.readbufsize {
				serv.ErrHandler.CreateError(&conn.Connection, 999999999999998, ErrOverSizeBuf)
				continue
			}
		}
		readfunc(conn, input, n, err)

	}
	return
}
func (serv *RTCPhelper) SetReadDeadLine(rdtime time.Duration) {
	serv.readtimeout = rdtime
}
func (serv *RTCPhelper) SetWriteDeadLine(wrtime time.Duration) {
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
			if res = sock.Connection.SetWriteDeadline(time.Now().Add(time.Second * serv.writetimeout)); res != nil {

				serv.ErrHandler.CreateError(&sock.Connection, 0, res)

				return
			}
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
func (serv *RTCPhelper) CloseConnections() {
	serv.ConectionList.CloseAll()
}
func (serv *RTCPhelper) Stop() {
	serv.stop = true
}
func (serv *RTCPhelper) Start() error {
	serv.stop = false
	if serv.ErrHandler == nil {
		return ErrEmptyErrHandler
	}

	return nil
}
func (serv *RTCPhelper) CanWork() bool {
	return !serv.stop
}
func (serv *RTCPhelper) CloseConectionByTag(Tag int) {
	if err := serv.ConectionList.RemoveConnByConTag(Tag); err != nil {
		c := serv.ConectionList.GetConnetionByTag(Tag).Connection
		serv.ErrHandler.CreateError(&c, Tag, err)
	}

}
func (serv *RTCPhelper) SetAfterConnectEvent(doafterconect AfterConnectHandler) {
	serv.startafterconnectevent = doafterconect
}
func GetErrOp(err error) net.OpError {
	errop := err.(*net.OpError)
	return *errop
}
