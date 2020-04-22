package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	serverIP := os.Args[1]

	t1 := time.Now()
	conn, _ := net.Dial("tcp", serverIP + ":5000")
	t2 := time.Now()
	fmt.Printf("TCP Setup: %f\n", t2.Sub(t1).Seconds())

	start := time.Now()
	fmt.Fprint(conn, "hello" + "\n")
	message, _ := bufio.NewReader(conn).ReadString('\n')
	end := time.Now()
	fmt.Printf("Round Trip Time: %f\n", end.Sub(start).Seconds())


	fmt.Println(message)
}
