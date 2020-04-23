package main

import (
	"flag"
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

	clientIP := os.Args[1]
	persistent := *flag.Bool("mode", false, "Whether to have sockets for entire duration.")
	if persistent {
		peristentConn(clientIP)
		return
	}

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
					data, _ := puller.RecvBytes(zmq.DONTWAIT)
					go handleData(data, clientIP, ctx)
				}
			}
		}
	}
}

func handleData(data []byte, clientIP string, ctx *zmq.Context) {
	start := time.Now()

	t1 := time.Now()
	pusher := createSocket(zmq.PUSH, ctx, fmt.Sprintf(PushTemplate, clientIP, 6000), false)
	t2 := time.Now()
	fmt.Printf("Socket Creation: %f\n", t2.Sub(t1).Seconds())
	defer pusher.Close()

	pusher.SendBytes(data, zmq.DONTWAIT)
	end := time.Now()
	fmt.Printf("Total Handler Time: %f\n", end.Sub(start).Seconds())
}

func peristentConn(clientIP string) {
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Error creating ZMQ context")
	}

	puller := createSocket(zmq.PULL, ctx, fmt.Sprintf(PullTemplate, 5000), true)
	pusher := createSocket(zmq.PUSH, ctx, fmt.Sprintf(PushTemplate, clientIP, 6000), false)
	poller := zmq.NewPoller()
	poller.Add(puller, zmq.POLLIN)

	for true {
		sockets, _ := poller.Poll(0)

		for _, socket := range sockets {
			switch s := socket.Socket; s {
			case puller:
				{
					data, _ := puller.RecvBytes(zmq.DONTWAIT)
					go handle(data, pusher)
				}
			}
		}
	}
}

func handle(data []byte, pusher *zmq.Socket) {
	pusher.SendBytes(data, zmq.DONTWAIT)
}
