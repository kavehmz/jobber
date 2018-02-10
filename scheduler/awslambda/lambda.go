package awslambda

import (
	"context"
	"encoding/json"
	"log"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"golang.org/x/time/rate"
)

// LambdaScheduler is simple scheduler to invoke long running lambda functions in AWS and
// Inviting them to connect back by gRPC, bidrectional, to serve mutliple requests.
type LambdaScheduler struct {
	GrpcHost string
	Lambda   *lambda.Lambda
	Limiter  *rate.Limiter
	Ctx      context.Context
	jobs     int32
	running  int32
	number   int32
}

func (g *LambdaScheduler) Inbound() {
	log.Println("minion: job inbound", g.jobs, g.running)
	atomic.AddInt32(&g.jobs, 1)
	if g.jobs > g.running-1 {
		go g.invoke()
	}
}
func (g *LambdaScheduler) Done() {
	log.Println("minion: job done")
	atomic.AddInt32(&g.jobs, -1)
}
func (g *LambdaScheduler) Timedout() {
	log.Println("minion: job timedout")
	atomic.AddInt32(&g.jobs, -1)
}

type input struct {
	In string `json:"in"`
}

func (g *LambdaScheduler) invoke() {
	log.Printf("scheduler[%d]: Hi, I see you need more workers. I will invoke one now.\n", g.number)
	log.Printf("scheduler[%d]: Checking the rate.\n", g.number)
	err := g.Limiter.Wait(g.Ctx)
	if err != nil {
		log.Printf("scheduler[%d]: Limiter error: %v\n", g.number, err)
	}

	payload, err := json.Marshal(input{In: g.GrpcHost})
	if err != nil {
		log.Println(err)
		return
	}

	atomic.AddInt32(&g.running, 1)
	defer func() { atomic.AddInt32(&g.running, -1) }()
	_, err = g.Lambda.Invoke(&lambda.InvokeInput{FunctionName: aws.String("lambda-handler"), Payload: payload})
	if err != nil {
		log.Printf("scheduler[%d]: Invitation failed.\n", g.number)
		return
	}
}
