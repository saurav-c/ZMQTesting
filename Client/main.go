package main

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"
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
		pusher = createSocket(zmq.PUSH, ctx, fmt.Sprintf(PushTemplate, address, 5000), false)
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
	serverIP := os.Args[1]
	mode := os.Args[2]

	if mode == "rr" {
		reqAndRep(serverIP)
		return
	} else if mode == "newRR" {
		newRR(serverIP)
		return
	} else if mode == "pnp" {
		ctx, err := zmq.NewContext()
		if err != nil {
			log.Fatalf("Error creating ZMQ context")
		}

		channel := make(chan int)
		var wg sync.WaitGroup

		go pushAndPull(serverIP, ctx, channel)
		wg.Add(1)
		go sender(&wg, serverIP, ctx, channel)
		wg.Wait()
		return
	}

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
	for i := 0; i < 20; i++ {
		start := time.Now()
		pusher.SendBytes([]byte("hello"), zmq.DONTWAIT)
		puller.RecvBytes(0)
		end := time.Now()
		fmt.Printf("Round Trip: %f ms\n", 1000*end.Sub(start).Seconds())
	}
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

func reqAndRep(serverIP string) {
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Error creating ZMQ context")
	}
	req, _ := ctx.NewSocket(zmq.REQ)
	defer req.Close()
	req.Connect(fmt.Sprintf(PushTemplate, serverIP, 5000))
	data := make([]byte, 512)
	rand.Read(data)

	for i := 0; i < 20; i++ {
		start := time.Now()
		req.SendBytes(data, zmq.DONTWAIT)
		req.RecvBytes(0)
		end := time.Now()
		fmt.Printf("Round Trip Time: %f ms\n", 1000 * end.Sub(start).Seconds())
	}
}

func newRR(serverIP string) {
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Error creating ZMQ context")
	}

	data := make([]byte, 512)
	rand.Read(data)

	for i := 0; i < 20; i++ {
		req, _ := ctx.NewSocket(zmq.REQ)
		defer req.Close()
		req.Connect(fmt.Sprintf(PushTemplate, serverIP, 5000))

		start := time.Now()
		req.SendBytes(data, zmq.DONTWAIT)
		req.RecvBytes(0)
		end := time.Now()
		fmt.Printf("Round Trip Time: %f ms\n", 1000 * end.Sub(start).Seconds())
	}
}

func pushAndPull(serverIP string, ctx *zmq.Context, channel chan int) {
	puller := createSocket(zmq.PULL, ctx, fmt.Sprintf(PullTemplate, 6000), true)
	defer puller.Close()
	poller := zmq.NewPoller()
	poller.Add(puller, zmq.POLLIN)

	for true {
		sockets, _ := poller.Poll(0)

		for _, socket := range sockets {
			switch s := socket.Socket; s {
			case puller:
				{
					puller.RecvBytes(zmq.DONTWAIT)
					channel <- 1
				}
			}
		}
	}
}

func sender(wg *sync.WaitGroup, serverIP string, ctx *zmq.Context, channel chan int) {
	defer wg.Done()
	cache := PusherCache{socketMap: make(map[string]*zmq.Socket)}
	data := make([]byte, 512)
	rand.Read(data)
	for i:=0; i < 20; i++ {
		pusher := cache.getPusher(serverIP, ctx)
		start := time.Now()
		pusher.SendBytes(data, zmq.DONTWAIT)

		<-channel

		end := time.Now()
		fmt.Printf("Round Trip Time: %f ms\n", 1000 * end.Sub(start).Seconds())
	}
	//end := time.Now()
	//fmt.Printf("Total Handler Time: %f\n", end.Sub(start).Seconds())
}