package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
	"os"
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

	clientIP := os.Args[1]

	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Error creating ZMQ context")
	}

	puller := createSocket(zmq.PULL, ctx, fmt.Sprintf(PullTemplate, 5000), true)
	poller := zmq.NewPoller()
	poller.Add(puller, zmq.POLLIN)

	for true {
		sockets, _ := poller.Poll(0)

		for _, socket := range sockets {
			switch s := socket.Socket; s {
			case puller:
				{
					data, _ := puller.Recv(zmq.DONTWAIT)
					go handleData(data, clientIP, ctx)
				}
			}
		}
	}
}

func handleData(data string, clientIP string, ctx *zmq.Context) {
	fmt.Println(string(data))
	s := "world"

	pusher := createSocket(zmq.PUSH, ctx, fmt.Sprintf(PushTemplate, clientIP, 6000), false)
	defer pusher.Close()

	pusher.Send(s, zmq.DONTWAIT)
}
