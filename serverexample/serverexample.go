// serveexample project main.go
package main

import (
	"fmt"
	"mtcpserver"
	"net"
	//	"time"
)

type MTCPServer struct {
	Serv mtcpserver.NET_TCPWorker
}

func (s *MTCPServer) CreateError(conn *net.Conn, ConnTag int, err error) { //обработчик ошибок
	//fmt.Println("Error Op", err)
	errop := mtcpserver.GetErrOp(err)
	switch errop.Op {
	case "read":
		{
			fmt.Println("Operation "+errop.Op+"  : ", "ConnTAg : ", ConnTag, err.Error())
			go s.Serv.CloseConnectionByTag(ConnTag)
			return
		}
	case "write":
		{
			fmt.Println("Operation "+errop.Op+"  : ", "ConnTAg : ", ConnTag, err.Error())
			go s.Serv.CloseConnectionByTag(ConnTag)
			return
		}
	case "accept":
		{
			fmt.Println("Operation : " + err.Error())

		}
	}

	//	done <- true
}
func (s *MTCPServer) ReadHandle(sock mtcpserver.ConectItem, buf []byte, size int, errorop error) {
	if errorop == nil {
		fmt.Println("Num Conn: ", sock.ConnTag, " ReceivedData: ", string(buf[:size]))
		ready := "READY"
		go s.Serv.Write(sock.ConnTag, s.WriteHandle, []byte(ready))
	} else {
		//	fmt.Println("Num Conn: ", sock.ConnTag, " Error: ", errorop)
		go s.CreateError(&sock.Connection, sock.ConnTag, errorop)
	}

}

func (s *MTCPServer) WriteHandle(sock mtcpserver.ConectItem, buf []byte, size int, errorop error) {
	if errorop == nil {
		fmt.Println("Num Conn: ", sock.ConnTag, " DATA Sended")

	} else {
		//	fmt.Println("Num Conn: ", sock.ConnTag, " Error: ", errorop)
		go s.CreateError(&sock.Connection, sock.ConnTag, errorop)
	}

}
func (s *MTCPServer) AfterConnect(sock mtcpserver.ConectItem) {
	fmt.Println("Num Conn: ", sock.ConnTag, " Started work")
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
	MyTCPServer.Serv.SetAfterConnectEvent(MyTCPServer.AfterConnect)
	fmt.Println("Start Listen Port")
	err := MyTCPServer.Serv.Start(":3001", MyTCPServer.ReadHandle)
	if err != nil {
		fmt.Println("START ERROR :", err)
		var s string
		fmt.Scanln(&s)
	}
}
