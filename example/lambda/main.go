// This example is using the aws lambda scheduler.
// It needs to run in an environment that lambda function can connect back to the gRPC server started by this app.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/kavehmz/jobber/jobber"
	"github.com/kavehmz/jobber/payload"
	"github.com/kavehmz/jobber/scheduler/awslambda"
	"golang.org/x/time/rate"
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
	log.Println("Example: received", resp, err)
	fmt.Fprint(w, resp.Data)
}

func serverGRPC(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Panic("failed to listen: ", err)
	}
	s := grpc.NewServer()

	// Prepare AWS and Lambda
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		log.Panic("failed to create session: ", err)
	}
	fn := lambda.New(sess, &aws.Config{Region: aws.String("us-east-1")})

	taskMachine = jobber.NewJobber(jobber.Scheduler(
		&awslambda.LambdaScheduler{
			GrpcHost: os.Getenv("GRPC_HOST") + ":50051",
			Lambda:   fn,
			// Set the rate of calling Lambda to 1 calls a second
			Limiter: rate.NewLimiter(rate.Limit(1), 1),
			Ctx:     context.Background(),
		}),
		jobber.MaxMinionLifetime(time.Second*13),
	)
	taskMachine.RegisterGRPC(s)

	log.Printf("Start listening gRPC at %d", port)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Panic("failed to serve: ", err)
		}
	}()

}

func serverHTTP(port int) {
	http.HandleFunc("/", hello)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
