package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

const (
	PullTemplate = "tcp://*:%d"
	PushTemplate = "tcp://%s:%d"
)

func createSocket(tp zmq.Type, context *zmq.Context, address string, bind bool) *zmq.Socket {
	sckt, err := context.NewSocket(tp)
	if err != nil {
		fmt.Println("Unexpected error while creating new socket:\n", err)
		os.Exit(1)
	}
	if bind {
		err = sckt.Bind(address)
	} else {
		err = sckt.Connect(address)
	}
	if err != nil {
		fmt.Println("Unexpected error while binding/connecting socket:\n", err)
		os.Exit(1)
	}
	return sckt
}

func main() {
	serverIP := os.Args[1]
	mode := os.Args[2]

	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Error creating ZMQ context")
	}

	start := time.Now()
	pusher := createSocket(zmq.PUSH, ctx, fmt.Sprintf(PushTemplate, serverIP, 5000), false)
	end := time.Now()
	fmt.Printf("Push Socket Creation Time: %f\n", end.Sub(start).Seconds())

	start = time.Now()
	puller := createSocket(zmq.PULL, ctx, fmt.Sprintf(PullTemplate, 6000), true)
	end = time.Now()
	fmt.Printf("Pull Socket Creation Time: %f\n", end.Sub(start).Seconds())


	defer pusher.Close()
	defer puller.Close()

	switch mode {
	case "simple":
		{
			simpleClient(ctx, pusher, puller, serverIP)
		}
	case "size":
		{
			varySize(ctx, pusher, puller, serverIP)
		}
	}
}

func simpleClient(ctx *zmq.Context, pusher *zmq.Socket, puller *zmq.Socket, serverIP string) {
	start := time.Now()
	pusher.SendBytes([]byte("hello"), zmq.DONTWAIT)
	data, _ := puller.RecvBytes(0)
	end := time.Now()
	fmt.Printf("Round Trip: %f ms\n", 1000 * end.Sub(start).Seconds())
	fmt.Println(string(data))
}

func varySize(ctx *zmq.Context, pusher *zmq.Socket, puller *zmq.Socket, serverIP string) {
	for i := 1; i <= 20; i++ {
		size := math.Pow(2, float64(i))
		data := make([]byte, int(size))
		rand.Read(data)

		start := time.Now()
		pusher.SendBytes(data, zmq.DONTWAIT)
		data, _ = puller.RecvBytes(0)
		end := time.Now()

		fmt.Printf("Round Trip for Size %f: %f ms\n", size, 1000 * end.Sub(start).Seconds())
	}

}