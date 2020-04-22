package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"net"
	"os"
	"time"
)

func main() {
	serverIP := os.Args[1]
	mode := os.Args[2]

	t1 := time.Now()
	conn, _ := net.Dial("tcp", serverIP+":5000")
	t2 := time.Now()
	fmt.Printf("TCP Setup: %f\n", t2.Sub(t1).Seconds())
	defer conn.Close()

	switch mode {
	case "simple":
		{
			simpleClient(conn)
		}
	case "size":
		{
			varySize(conn)
		}
	}
}

func simpleClient(conn net.Conn) {
	data := []byte("hello")
	data = append(data, '\n')
	start := time.Now()
	conn.Write(data)
	message, _ := bufio.NewReader(conn).ReadBytes('\n')
	end := time.Now()
	fmt.Printf("Round Trip Time: %f ms\n", 1000 * end.Sub(start).Seconds())
	fmt.Println(string(message))
}

func varySize(conn net.Conn) {
	for i := 1; i <= 20; i++ {
		size := math.Pow(2, float64(i))
		data := make([]byte, int(size))
		rand.Read(data)
		data = append(data, '\n')

		start := time.Now()
		conn.Write(data)
		bufio.NewReader(conn).ReadBytes('\n')
		end := time.Now()

		fmt.Printf("Round Trip for Size %f: %f ms\n", size, 1000 * end.Sub(start).Seconds())
	}
}
