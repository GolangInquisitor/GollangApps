# GollangApps
сервер TCP
package main

import (
	"fmt"
	"mtcpserver"
	"net"
)

type Errhandle struct{}

func (e *Errhandle) CreateError(conn *net.Conn, ConnTag int, err error) {
	fmt.Println("Error Op", err)
	(*conn).Close()

}
func CreateReadHandle() mtcpserver.ReadHandler {
	return func(sock mtcpserver.ConectItem, buf []byte, size int, errorop error) {

		if errorop == nil {
			fmt.Println("Num Conn: ", sock.ConnTag, " ReceivedData: ", string(buf))
			ready := "READY"
			MyTCPServer.Write(sock.ConnTag, CreateWriteHandle(), []byte(ready))
		} else {
			fmt.Println("Num Conn: ", sock.ConnTag, " Error: ", errorop)
		}

	}
}
func CreateWriteHandle() mtcpserver.WriteHandler {
	return func(sock mtcpserver.ConectItem, buf []byte, size int, errorop error) {
		if errorop == nil {
			fmt.Println("Num Conn: ", sock.ConnTag, " DATA Sended")
		} else {
			fmt.Println("Num Conn: ", sock.ConnTag, " Error: ", errorop)
		}
	}
}

var MyTCPServer = mtcpserver.NewTcpWorker(new(mtcpserver.RTCPhelper))

func main() {
	MyTCPServer.SetErrorHandler(new(Errhandle))
	MyTCPServer.SetNumConnections(10)
	MyTCPServer.SetReadBufSize(1024)
	MyTCPServer.SetWriteBufSize(1024)
	fmt.Println("Start Listen Port")
	MyTCPServer.Start(":1179", CreateReadHandle())

}
