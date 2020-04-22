package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
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

	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Error creating ZMQ context")
	}

	start := time.Now()
	pusher := createSocket(zmq.PUSH, ctx, fmt.Sprintf(PushTemplate, serverIP, 5000), false)
	puller := createSocket(zmq.PULL, ctx, fmt.Sprintf(PullTemplate, 6000), true)
	end := time.Now()
	fmt.Printf("Socket Creation Time: %f\n", end.Sub(start).Seconds())

	defer pusher.Close()
	defer puller.Close()

	start = time.Now()
	pusher.Send("hello", zmq.DONTWAIT)
	data, _ := puller.Recv(0)
	end = time.Now()
	fmt.Printf("Round Trip: %f\n", end.Sub(start).Seconds())

	fmt.Println(data)

}