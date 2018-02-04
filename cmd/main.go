package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kavehmz/jobber/handover"
	"github.com/kavehmz/jobber/jobber"
	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type params struct {
	GrpcPort int `envconfig:"GRPCPORT" default:"50051"`
	HTTPPort int `envconfig:"HTTPPORT" default:"8080"`
}

func parseServer() params {
	var p params
	err := envconfig.Process("", &p)
	if err != nil {
		envconfig.Usage("", &p)
		log.Fatal(err.Error())
	}
	return p
}

func main() {
	param := parseServer()

	s := serverGRPC(param.GrpcPort)

	shutdown := make(chan int, 1)
	sign := make(chan os.Signal, 1)

	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sign
		fmt.Println("Shutting down...")
		s.Stop()
		shutdown <- 1
	}()
	<-shutdown
	return
}

func serverGRPC(port int) *grpc.Server {
	creds, err := credentials.NewServerTLSFromFile("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		log.Fatalf("Failed to setup tls: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panic("failed to listen: ", err)
	}
	s := grpc.NewServer(grpc.Creds(creds))
	reflection.Register(s)

	j := jobber.NewJobber()
	j.RegisterGRPC(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Panic("failed to serve: ", err)
		}
	}()
	log.Printf("Grpc Listening on port %d", port)

	go func() {
		time.Sleep(time.Second * 5)
		r, e := j.Do(&handover.Task{Data: "TESTTEST"})
		fmt.Println("Hey:", r, e)
	}()

	return s
}
