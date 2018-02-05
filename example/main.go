package main

import (
	"fmt"
	"log"
	"net"

	"github.com/kavehmz/jobber/jobber"
	"github.com/kavehmz/jobber/payload"
	"google.golang.org/grpc"
)

func main() {

	serverGRPC(50051)

}

func serverGRPC(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panic("failed to listen: ", err)
	}
	s := grpc.NewServer()
	j := jobber.NewJobber(jobber.MinionScheduler(&jobber.Goroutine{}))
	j.RegisterGRPC(s)

	log.Printf("Start listening gRPC at %d", port)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Panic("failed to serve: ", err)
		}
	}()

	// Add some jobs for test
	for i := 0; i < 15; i++ {
		log.Println("Example: Creating a new job", i)
		r, e := j.Do(&payload.Task{Data: "This is the payload I will send to Lambda."})
		log.Println("Example: Recevied", r, e)
	}
}
