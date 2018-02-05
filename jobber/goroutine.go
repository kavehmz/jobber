package jobber

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/kavehmz/jobber/handover"
	"google.golang.org/grpc"
)

// Goroutine is simple Inviter which just spins up a grpc call in goroutine.
// I can imagine any use for it except testing the logic.
type Goroutine struct {
	jobs    uint
	running uint
	sync.Mutex
}

func (g *Goroutine) Inbound() {
	log.Println("job inbound")
	g.Lock()
	g.jobs++
	if g.jobs > g.running {
		go g.worker()
	}
	g.Unlock()
}
func (g *Goroutine) Done() {
	log.Println("job done")
	g.Lock()
	g.jobs--
	g.Unlock()
}
func (g *Goroutine) Timedout() {
	log.Println("job timedout")
	g.Lock()
	g.jobs--
	g.Unlock()
}

// This is a dummy worker that ignores the payloads and only sleeps a bit
func (g *Goroutine) worker() {
	g.Lock()
	g.running++
	g.Unlock()

	defer func() {
		g.Lock()
		g.running--
		g.Unlock()
	}()
	// creds, err := credentials.NewClientTLSFromFile("cert/server-cert.pem", "localhost")
	// if err != nil {
	// 	log.Fatalf("cert load error: %s", err)
	// }

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("Failed to connect: %v", err)
		return
	}
	defer conn.Close()

	c := handover.NewHandoverClient(conn)
	stream, err := c.Join(context.Background())
	if err != nil {
		log.Println("Failed to connect: %v", err)
		return
	}

	log.Println("worker: Joined")
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Println("worker: received error", err)
				return
			}
			log.Println("worker: reveiced a task from server", in.String())

			// This is what worker does. The rest is the template how to write and strean in grpc
			time.Sleep(time.Millisecond * 50)

			if err = stream.Send(&handover.Result{Data: time.Now().String()}); err != nil {
				log.Println("startWorker: Failed to send, ", err)
				return
			}
		}
	}()
	<-time.After(time.Second * 5)
	stream.CloseSend()
	time.Sleep(time.Second)
}
