package server

import (
	"bufio"
	"fmt"
	"net"
)
func main() {
	ln, _ := net.Listen("tcp", ":5000")
	conn, _ := ln.Accept()
	defer ln.Close()

	for {
		message, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Println(message)
		conn.Write([]byte("world" + "\n"))
	}

}
