package server

import (
	"bufio"
	"net"
)
func main() {
	ln, _ := net.Listen("tcp", ":5000")
	conn, _ := ln.Accept()
	defer ln.Close()

	for {
		message, _ := bufio.NewReader(conn).ReadBytes('\n')
		conn.Write(append(message, '\n'))
	}
}
