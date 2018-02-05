package jobber

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
	jobs    int32
	running int32
}

func (g *Goroutine) Inbound() {
	log.Println("minion: job inbound XXXXXXXXXXXX", g.jobs, g.running)
	atomic.AddInt32(&g.jobs, 1)
	if g.jobs > g.running-1 {
		go g.worker()
	}
}
func (g *Goroutine) Done() {
	log.Println("minion: job done")
	atomic.AddInt32(&g.jobs, -1)
}
func (g *Goroutine) Timedout() {
	log.Println("minion: job timedout")
	atomic.AddInt32(&g.jobs, -1)
}

// This is a dummy worker that ignores the payloads and only sleeps a bit
func (g *Goroutine) worker() {
	log.Println("worker: Hi, I was invoked and I am trying to connect to accept jobs")
	atomic.AddInt32(&g.running, 1)

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Println("worker: Failed to dial: %v", err)
		atomic.AddInt32(&g.running, -1)
		return
	}
	defer conn.Close()

	c := payload.NewPayloadClient(conn)
	stream, err := c.Join(context.Background())
	if err != nil {
		log.Println("worker: Failed to join: %v", err)
		atomic.AddInt32(&g.running, -1)
		return
	}

	log.Println("worker: I Joined the workforce")
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				atomic.AddInt32(&g.running, -1)
				return
			}
			if err != nil {
				log.Println("worker: received error", err)
				atomic.AddInt32(&g.running, -1)
				return
			}
			log.Println("worker: reveiced a task from server", in.String())

			// This is what worker does. The rest is the template how to write and strean in grpc
			time.Sleep(time.Millisecond * 900)
			log.Println("worker: task is done")
			if err = stream.Send(&payload.Result{Data: time.Now().String()}); err != nil {
				log.Println("worker: Failed to send, ", err)
				atomic.AddInt32(&g.running, -1)
				return
			}
		}
	}()
	<-time.After(time.Second * 5)
	atomic.AddInt32(&g.running, -1)
	log.Println("worker: my time is over (5s) Bye!")
	stream.CloseSend()
	time.Sleep(time.Second)
}
