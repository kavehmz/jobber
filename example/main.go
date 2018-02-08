package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/kavehmz/jobber/jobber"
	"github.com/kavehmz/jobber/payload"
	"google.golang.org/grpc"
)

var taskMachine *jobber.Jobber

func main() {
	serverGRPC(50051)
	serverHTTP(8000)

}

func hello(w http.ResponseWriter, r *http.Request) {
	resp, err := taskMachine.Do(&payload.Task{Data: "This is the payload I will send to Lambda."})
	if err != nil {
		resp = &payload.Result{Data: "Because of error result was returned as nil"}
	}
	log.Println("Example: Recevied", resp, err)
	fmt.Fprint(w, resp.Data)
}

func serverHTTP(port int) {
	http.HandleFunc("/", hello)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func serverGRPC(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panic("failed to listen: ", err)
	}
	s := grpc.NewServer()
	taskMachine = jobber.NewJobber(jobber.MinionScheduler(&jobber.Goroutine{GrpcHost: "localhost:50051"}), jobber.MaxMinionLifetime(time.Second*13))
	taskMachine.RegisterGRPC(s)

	log.Printf("Start listening gRPC at %d", port)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Panic("failed to serve: ", err)
		}
	}()

}
