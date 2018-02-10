// This is a simple aws lambda function you can use for lambda scheduler.
// This one does nothing except a 50ms sleep.
// Lambda function connect back to caller (CallBackServer) as soon as it is invoked
// Connection is grpc and bidirectional.
package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/kavehmz/jobber/payload"

	"github.com/aws/aws-lambda-go/lambda"
	"google.golang.org/grpc"
)

type output struct {
	Out string
}

type input struct {
	CallBackServer string
	MaxServeTime   int
}

func handler(ctx context.Context, in input) (output, error) {
	join(in)
	return output{Out: "Bye"}, nil
}

func main() {
	lambda.Start(handler)
}

func join(in input) {
	conn, err := grpc.Dial(in.CallBackServer, grpc.WithInsecure())
	if err != nil {
		log.Println("Failed to dial the callbackserver: ", err)
		return
	}
	defer conn.Close()

	c := payload.NewPayloadClient(conn)
	stream, err := c.Join(context.Background())
	if err != nil {
		log.Println("Failed to join: ", err)
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

			if err = stream.Send(&payload.Result{Data: time.Now().String()}); err != nil {
				log.Println("startWorker: Failed to send, ", err)
				return
			}
		}
	}()
	<-time.After(time.Second * time.Duration(in.MaxServeTime))
	stream.CloseSend()
	time.Sleep(time.Second)
}
