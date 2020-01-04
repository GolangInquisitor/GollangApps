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
	ErrRemoveByTag  = errors.New("Error remove connection by Tag Not Found Conection")
	ErrRemoveByIndx = errors.New("Error remove connection by Index Not Found Conection")
)

type RNetError = interface {
	CreateError(conn *net.Conn, ConnTag int, err error, ready_ch chan bool)
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

type ReadHandler func(sock ConectItem, buf []byte, size int, errorop error, done chan bool)
type WriteHandler func(sock ConectItem, buf []byte, size int, errorop error, done chan bool)

type RListenTCP interface {
	ListenHandler(sock net.Conn, readfunc ReadHandler, err error)
	WriteBufSize() int
	ReadBufSize() int
	SetReadBufSize(size int)
	SetWriteBufSize(size int)
	Write(ConTag int, writefunc WriteHandler, byte_buf []byte, done chan bool)
	SetErrorHandler(errhandler RNetError)
	SetNumConnections(numcon int)
	CloseConectionByTag(Tag int, done chan bool)
	SetReadDeadLine(rdtime time.Duration)
	SetWriteDeadLine(wrtime time.Duration)
	CloseConnections()
	Stop()
	Start()
	CanWork() bool
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
func (tserver *NET_TCPWorker) SetReadDeadLine(rdtime time.Duration) {
	tserver.tcpserver.SetReadDeadLine(rdtime)
}
func (tserver *NET_TCPWorker) SetWriteDeadLine(wrtime time.Duration) {
	tserver.tcpserver.SetWriteDeadLine(wrtime)
}
func (tserver *NET_TCPWorker) Stop() {
	tserver.tcpserver.Stop()
}
func (tserver *NET_TCPWorker) CloseConnectionByTag(tag int, done chan bool) {

	tserver.tcpserver.CloseConectionByTag(tag, done)
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
	tserver.tcpserver.Start()
	for tserver.tcpserver.CanWork() {
		listener.SetDeadline(time.Now().Add(time.Second * 10))
		conn, err := listener.Accept()
		tserver.tcpserver.ListenHandler(conn, readfunc, err)

	}
	tserver.tcpserver.CloseConnections()
	return nil
}
func (tserver *NET_TCPWorker) Write(ConTag int, writefunc WriteHandler, byte_buf []byte, done chan bool) {
	go tserver.tcpserver.Write(ConTag, writefunc, byte_buf, done)
}

type RTCPhelper struct {
	ConectionList NetList
	ErrHandler    RNetError
	readbufsize   int
	writebufsize  int
	readtimeout   time.Duration
	writetimeout  time.Duration
	stop          bool
}

func (serv *RTCPhelper) ListenHandler(sock net.Conn, readfunc ReadHandler, err error) {
	if serv.ErrHandler != nil {
		var res error
		if err == nil {
			var citem ConectItem
			if citem, res = serv.ConectionList.AddConnection(sock); res != nil {
				done := make(chan bool)
				go serv.ErrHandler.CreateError(&sock, 0, res, done)
				<-done
				return
			}

			fmt.Println("New conect : " + sock.RemoteAddr().String())
			go serv.read(citem, readfunc)
		} else {
			done := make(chan bool)
			go serv.ErrHandler.CreateError(&sock, 0, err, done)
			<-done
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
	done := make(chan bool)
	for !conn.Empty && !serv.stop {
		if res := conn.Connection.SetReadDeadline(time.Now().Add(time.Second * serv.readtimeout)); res != nil {
			go serv.ErrHandler.CreateError(&conn.Connection, 0, res, done)
			<-done
			continue
		}
		n, err := conn.Connection.Read(input)
		go readfunc(conn, input, n, err, done)
		<-done
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
func (serv *RTCPhelper) Write(ConTag int, writefunc WriteHandler, byte_buf []byte, done chan bool) {
	var res error
	var sendedlen int = 0
	var sock ConectItem
	ready := make(chan bool)
	if serv.writebufsize >= len(byte_buf) {
		sock = serv.ConectionList.GetConnetionByTag(ConTag)
		if !sock.Empty {
			if res = sock.Connection.SetWriteDeadline(time.Now().Add(time.Second * serv.writetimeout)); res != nil {
				//	done := make(chan bool)
				go serv.ErrHandler.CreateError(&sock.Connection, 0, res, done)
				<-done
				return
			}
			sendedlen, res = sock.Connection.Write(byte_buf)
		} else {
			res = ErrEmtyConn
		}
	} else {
		res = ErrOverSizeBuf
	}
	go writefunc(sock, byte_buf, sendedlen, res, ready)
	<-ready
	done <- true
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
func (serv *RTCPhelper) Start() {
	serv.stop = false
}
func (serv *RTCPhelper) CanWork() bool {
	return !serv.stop
}
func (serv *RTCPhelper) CloseConectionByTag(Tag int, done chan bool) {
	if err := serv.ConectionList.RemoveConnByConTag(Tag); err != nil {
		c := serv.ConectionList.GetConnetionByTag(Tag).Connection
		donem := make(chan bool)
		go serv.ErrHandler.CreateError(&c, Tag, err, donem)
		var d bool = false
		d = <-donem
		done <- d
	}
	done <- true
	return
}
func GetErrOp(err error) net.OpError {
	errop := err.(*net.OpError)
	return *errop
}
