package goroutine

import (
	"context"
	"io"
	"log"
	"sync/atomic"
	"time"

	"github.com/kavehmz/jobber/payload"
	"google.golang.org/grpc"
)

// Goroutine is simple Inviter which just spins up a grpc call in goroutine.
// I can't imagine any use for it except testing the logic.
type Goroutine struct {
	jobs     int32
	running  int32
	GrpcHost string
	number   int32
}

// Inbound is called before a new task is added.
func (g *Goroutine) Inbound() {
	log.Println("minion: job inbound", g.jobs, g.running)
	atomic.AddInt32(&g.jobs, 1)
	if g.jobs > g.running-1 {
		atomic.AddInt32(&g.number, 1)
		go g.worker(g.number)
	}
}

// Done is called when a task is done
func (g *Goroutine) Done() {
	log.Println("minion: job done")
	atomic.AddInt32(&g.jobs, -1)
}

// Timedout is called when no response was received on time for a task
func (g *Goroutine) Timedout() {
	log.Println("minion: job timedout")
	atomic.AddInt32(&g.jobs, -1)
}

// This is a dummy worker that ignores the payloads and only sleeps a bit
func (g *Goroutine) worker(n int32) {
	log.Printf("worker[%d]: Hi, I was invoked and I am trying to connect to accept jobs\n", n)
	atomic.AddInt32(&g.running, 1)

	conn, err := grpc.Dial(g.GrpcHost, grpc.WithInsecure())
	if err != nil {
		log.Printf("worker[%d]: Failed to dial: %v\n", n, err)
		atomic.AddInt32(&g.running, -1)
		return
	}
	defer conn.Close()

	c := payload.NewPayloadClient(conn)
	stream, err := c.Join(context.Background())
	if err != nil {
		log.Printf("worker[%d]: Failed to join: %v\n", n, err)
		atomic.AddInt32(&g.running, -1)
		return
	}

	log.Printf("worker[%d]: I Joined the workforce\n", n)
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				atomic.AddInt32(&g.running, -1)
				return
			}
			if err != nil {
				log.Printf("worker[%d]: received error %v\n", n, err)
				atomic.AddInt32(&g.running, -1)
				return
			}
			log.Printf("worker[%d]: reveiced a task from server %s\n", n, in.String())

			// This is what worker does. The rest is the template how to write and strean in grpc
			time.Sleep(time.Millisecond * 900)
			log.Printf("worker[%d]: task is done\n", g.number)
			if err = stream.Send(&payload.Result{Data: time.Now().String()}); err != nil {
				log.Printf("worker[%d]: Failed to send, %v\n", n, err)
				atomic.AddInt32(&g.running, -1)
				return
			}
		}
	}()
	<-time.After(time.Second * 15)
	log.Printf("worker[%d]: my time is over (15s) Bye!\n", n)
	stream.CloseSend()
	time.Sleep(time.Second)
}
