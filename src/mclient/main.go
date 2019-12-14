// testdial1 project main.go
package main

import (
	"fmt"
	"mclient"
	"net"
	"time"
	//"mtcpserver"
)

func RHandler(sock *net.Conn, buf []byte, size int, errorop error) {
	if errorop == nil {
		fmt.Println("Read : ", buf)

	} else {
		fmt.Println("Read : ", errorop)
	}
}
func WHandler(sock *net.Conn, buf []byte, size int, errorop error) {
	if errorop == nil {
		fmt.Println("Write : ", buf)

	} else {
		fmt.Println("Write : ", errorop)
	}
}

type Errhandle struct{}

func (e *Errhandle) CreateError(conn *net.Conn, ConnTag int, err error) { //обработчик ошибок
	fmt.Println("Error Op", err)
	(*conn).Close()
}

var TClient = mclient.NewClientHelper(new(mclient.RclientHelper))

func k() {
	for {
		var source string
		fmt.Print("Введите слово: ")
		fmt.Scanln(&source)
		TClient.Write([]byte(source), WHandler)

	}
}
func main() {

	go k()
	TClient.SetErrorHandler(new(Errhandle))
	TClient.SetReadDeadLine(60 * time.Second)
	TClient.SetWriteDeadLine(60 * time.Second)
	TClient.SetReadBufSize(500)
	TClient.SetWriteBufSize(500)
	if err := TClient.Start("127.0.0.1:1179", RHandler); err != nil {
		fmt.Println(err)
	}

}
