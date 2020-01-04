// serveexample project main.go
package main

import (
	"fmt"
	"mtcpserver"
	"net"
	//	"time"
)

/*type Errhandle struct{}

func (e *Errhandle) CreateError(conn *net.Conn, ConnTag int, err error, done chan bool) { //обработчик ошибок
	//fmt.Println("Error Op", err)
	errop := mtcpserver.GetErrOp(err)
	switch errop.Op {
	case "read":
		{
			fmt.Println("Operation "+errop.Op+"  : ", "ConnTAg : ", ConnTag, err.Error())
			(*conn).Close()

		}
	case "write":
		{
			fmt.Println("Operation "+errop.Op+"  : ", "ConnTAg : ", ConnTag, err.Error())
			(*conn).Close()
		}
	case "accept":
		{
			fmt.Println("Operation : " + err.Error())
		}
	}

	done <- true
}
func ReadHandle(sock mtcpserver.ConectItem, buf []byte, size int, errorop error, done chan bool) {
	if errorop == nil {
		fmt.Println("Num Conn: ", sock.ConnTag, " ReceivedData: ", string(buf[:size]))
		ready := "READY"
		MyTCPServer.Write(sock.ConnTag, WriteHandle, []byte(ready), done)
	} else {
		fmt.Println("Num Conn: ", sock.ConnTag, " Error: ", errorop)
		done <- true
	}

}
func WriteHandle(sock mtcpserver.ConectItem, buf []byte, size int, errorop error, done chan bool) {
	if errorop == nil {
		fmt.Println("Num Conn: ", sock.ConnTag, " DATA Sended")
	} else {
		fmt.Println("Num Conn: ", sock.ConnTag, " Error: ", errorop)
	}
	done <- true
}

var MyTCPServer = mtcpserver.NewTcpWorker(new(mtcpserver.RTCPhelper))*/
type MTCPServer struct {
	Serv mtcpserver.NET_TCPWorker
}

func (s *MTCPServer) CreateError(conn *net.Conn, ConnTag int, err error, done chan bool) { //обработчик ошибок
	//fmt.Println("Error Op", err)
	errop := mtcpserver.GetErrOp(err)
	switch errop.Op {
	case "read":
		{
			fmt.Println("Operation "+errop.Op+"  : ", "ConnTAg : ", ConnTag, err.Error())
			go s.Serv.CloseConnectionByTag(ConnTag, done)
			return
		}
	case "write":
		{
			fmt.Println("Operation "+errop.Op+"  : ", "ConnTAg : ", ConnTag, err.Error())
			go s.Serv.CloseConnectionByTag(ConnTag, done)
			return
		}
	case "accept":
		{
			fmt.Println("Operation : " + err.Error())
			done <- true
		}
	}

	//	done <- true
}
func (s *MTCPServer) ReadHandle(sock mtcpserver.ConectItem, buf []byte, size int, errorop error, done chan bool) {
	if errorop == nil {
		fmt.Println("Num Conn: ", sock.ConnTag, " ReceivedData: ", string(buf[:size]))
		ready := "READY"
		go s.Serv.Write(sock.ConnTag, s.WriteHandle, []byte(ready), done)
	} else {
		//	fmt.Println("Num Conn: ", sock.ConnTag, " Error: ", errorop)
		go s.CreateError(&sock.Connection, sock.ConnTag, errorop, done)
	}

}
func (s *MTCPServer) WriteHandle(sock mtcpserver.ConectItem, buf []byte, size int, errorop error, done chan bool) {
	if errorop == nil {
		fmt.Println("Num Conn: ", sock.ConnTag, " DATA Sended")
		done <- true
	} else {
		//	fmt.Println("Num Conn: ", sock.ConnTag, " Error: ", errorop)
		go s.CreateError(&sock.Connection, sock.ConnTag, errorop, done)
	}

}

var MyTCPServer MTCPServer

func main() {
	MyTCPServer.Serv = mtcpserver.NewTcpWorker(new(mtcpserver.RTCPhelper))
	MyTCPServer.Serv.SetErrorHandler(&MyTCPServer)
	MyTCPServer.Serv.SetNumConnections(10)
	MyTCPServer.Serv.SetReadBufSize(1024)
	MyTCPServer.Serv.SetWriteBufSize(1024)
	MyTCPServer.Serv.SetReadDeadLine(180)
	MyTCPServer.Serv.SetWriteDeadLine(180)
	fmt.Println("Start Listen Port")
	MyTCPServer.Serv.Start(":1179", MyTCPServer.ReadHandle)
}
