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

type PusherCache struct {
	socketMap map[string]*zmq.Socket
}

func (cache *PusherCache) getPusher(address string, ctx *zmq.Context) (*zmq.Socket) {
	if pusher, ok := cache.socketMap[address]; !ok {
		pusher = createSocket(zmq.PUSH, ctx, fmt.Sprintf(PushTemplate, address, 6000), false)
		cache.socketMap[address] = pusher
		return pusher
	} else {
		return pusher
	}
}

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
	mode := os.Args[2]

	if mode == "pers" {
		persistent(clientIP)
		return
	} else if mode == "rr" {
		reqAndRep()
		return
	} else if mode == "persRR" {
		persistentRR(clientIP)
	}

	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Error creating ZMQ context")
	}

	cache := &PusherCache{
		socketMap:   make(map[string]*zmq.Socket),
	}

	puller := createSocket(zmq.PULL, ctx, fmt.Sprintf(PullTemplate, 5000), true)
	defer puller.Close()
	poller := zmq.NewPoller()
	poller.Add(puller, zmq.POLLIN)

	for true {
		sockets, _ := poller.Poll(0)

		for _, socket := range sockets {
			switch s := socket.Socket; s {
			case puller:
				{
					data, _ := puller.RecvBytes(zmq.DONTWAIT)
					go handleData(data, clientIP, ctx, cache)
				}
			}
		}
	}
}

func handleData(data []byte, clientIP string, ctx *zmq.Context, cache *PusherCache) {
	//start := time.Now()

	//t1 := time.Now()
	//	//pusher := createSocket(zmq.PUSH, ctx, fmt.Sprintf(PushTemplate, clientIP, 6000), false)
	//	//t2 := time.Now()
	//	//fmt.Printf("Socket Creation: %f\n", t2.Sub(t1).Seconds())
	//	//defer pusher.Close()
	pusher := cache.getPusher(clientIP, ctx)
	pusher.SendBytes(data, zmq.DONTWAIT)
	//end := time.Now()
	//fmt.Printf("Total Handler Time: %f\n", end.Sub(start).Seconds())
}

func persistent(clientIP string) {
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Error creating ZMQ context")
	}

	puller := createSocket(zmq.PULL, ctx, fmt.Sprintf(PullTemplate, 5000), true)
	pusher := createSocket(zmq.PUSH, ctx, fmt.Sprintf(PushTemplate, clientIP, 6000), false)
	poller := zmq.NewPoller()
	poller.Add(puller, zmq.POLLIN)
	defer puller.Close()
	defer pusher.Close()

	for true {
		sockets, _ := poller.Poll(0)

		for _, socket := range sockets {
			switch s := socket.Socket; s {
			case puller:
				{
					data, _ := puller.RecvBytes(zmq.DONTWAIT)
					pusher.SendBytes(data, zmq.DONTWAIT)
				}
			}
		}
	}
}

func persistentRR(clientIP string) {
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Error creating ZMQ context")
	}

	rep := createSocket(zmq.REP, ctx, fmt.Sprintf(PullTemplate, 5000), true)
	defer rep.Close()
	poller := zmq.NewPoller()
	poller.Add(rep, zmq.POLLIN)

	for true {
		sockets, _ := poller.Poll(0)

		for _, socket := range sockets {
			switch s := socket.Socket; s {
			case rep:
				{
					data, _ := rep.RecvBytes(zmq.DONTWAIT)
					rep.SendBytes(data, zmq.DONTWAIT)
				}
			}
		}
	}
}

func handle(data []byte, pusher *zmq.Socket) {
	pusher.SendBytes(data, zmq.DONTWAIT)
}

func reqAndRep() {
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Error creating ZMQ context")
	}
	rep, _ := ctx.NewSocket(zmq.REP)
	rep.Connect(fmt.Sprintf(PullTemplate, 5000))
	defer rep.Close()

	for {
		data, _ := rep.RecvBytes(0)
		rep.SendBytes(data, zmq.DONTWAIT)
	}
}
