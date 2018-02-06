# jobber

Jobber is an idea and a sample implementation.

Note: there is no unit test in this repo because it is only a demo for a concept.

# Background

Past - Dealing with rigid bare-matal servers. minimum granuality : -
Recently - Setting up an auto-scaling environment, using Kubernetes or other tools. Scaling up and down based on incoming requersts. minimum granuality : one VM
Today - Using Lambda or Google functions we can scale up to 10,000 cpu in few seconds and then scale down to nothing. minimum granuality: one core, 128GB ram

For both AWS Lambda and Google cloud functions the catch is their mininum 100ms time granuality.
It means if I want to serve my http requests, and they only take 10ms, I will pay 10x more
because both platforms will charge me for 100ms of time. Based on current pricing it is expensive.

# Solution

Here the idea is to invoke a Lambda function but instead of asking it to do one reuqest we will keep it around to serve many more but sending back the results without delay.

Method is easy.

We create an http server and also a gRPC server. When we have traffic we invoke one or more Lambda fuctions.

Those Lambda functions create a bidirectional connection to our gRPC server which is running along with http server.

Now http server relays the load to Lambda functions through grpc and gets the result back.

We just need a mechanism to invoke enough Lambda function to handle our traffic.

This repo has one sample implementation just to show the concept (no real production code).

# Tools

Solution does not depend on any what I picked, but I took gRPC as RPC framework and protobuf as data interchange format and Go to implement it.

# Data interchange format

Protobuf is a simple [format](https://developers.google.com/protocol-buffers/docs/proto3).

Payload definition is at `payload/payload.proto`.

# Lambda scheduler

To invoke Lambda function you need to pass the scheduler to `NewJobber`. `main.go` is passing a dummy one like this:

```Go
	s := grpc.NewServer()
	taskMachine = jobber.NewJobber(jobber.MinionScheduler(&jobber.Goroutine{GrpcHost: "localhost:50051"}))
	taskMachine.RegisterGRPC(s)
```

`Goroutine` implements :

```Go
type Minion interface {
	// Inbound is called before a new task is added.
	Inbound()
	// Done is called when a task is done
	Done()
	// Timedout is called when no response was received on time for a task
	Timedout()
}
```

![flow](https://kavehmz.github.io/static/images/lambda_for_micro_jobs.png "Lambda for micro requests")

This solution does not depend on protobuf, grpc or aws lambda but in this implementation I picked those tools,

# Test run

`example` includes a dummy test case. It invokes Goroutines instead of Lambda but concept is identical.


