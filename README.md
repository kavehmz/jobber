jobber
=========

[![Go Lang](http://kavehmz.github.io/static/gopher/gopher-front.svg)](https://golang.org/)
[![GoDoc](https://godoc.org/github.com/kavehmz/jobber?status.svg)](https://godoc.org/github.com/kavehmz/jobber)
[![Build Status](https://travis-ci.org/kavehmz/jobber.svg?branch=master)](https://travis-ci.org/kavehmz/jobber)
[![Coverage Status](https://coveralls.io/repos/kavehmz/jobber/badge.svg?branch=master&service=github)](https://coveralls.io/github/kavehmz/jobber?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/kavehmz/jobber)](https://goreportcard.com/report/github.com/kavehmz/jobber)

Jobber is an idea and a sample implementation.

# Background

Past - Dealing with rigid bare-matal servers. minimum granuality : -
Recently - Setting up an auto-scaling environment, using Kubernetes or other tools. Scaling up and down based on incoming requersts. minimum granuality : one VM
Today - Using Lambda or Google functions we can scale up to 10,000 cpu in few seconds and then scale down to nothing. minimum granuality: one core, 128GB ram

For both AWS Lambda and Google cloud functions the catch is their mininum 100ms time granuality.
It means if I want to serve my http requests, and they only take 10ms, I will pay 10x more
because both platforms will charge me for 100ms of time. Based on current pricing it is expensive.

# Solution

Here the idea is to invoke a Lambda function but instead of asking it to do one reuqest we will keep it around to serve many more.

Method is easy.

We create a http server and also a gRPC server. When we have traffic we invoke one or more Lambda fuctions.

Those Lambda functions create a bidirectional connection to our gRPC server which is running along with http server.

Now http server relays the load to Lambda functions through grpc and gets the result back.

We just need a mechanism to invoke enough Lambda function to handle our traffic.

This repo has one sample implementation to show the concept.

# Tools

Solution does not depend on these, but I picked gRPC as RPC framework and protobuf as data interchange format and Go to implement it.

# Data interchange format

Protobuf is a simple [format](https://developers.google.com/protocol-buffers/docs/proto3).

Payload definition is at `payload/payload.proto`.

If you are gonna encode/decode your data what is there is enough, otherwise edit `payload.proto` and regenereate the go file.

# Lambda scheduler

To invoke Lambda function you need to pass the scheduler to `NewJobber`.

I implemented to scheduler.

- goroutine.Goroutine: a dummy scheduler which is there only for test purposes.
- awslambda.LambdaScheduler: A simple scheduler which can invoke a lambda function to send the jobs to it.

```Go
	s := grpc.NewServer()
	taskMachine = jobber.NewJobber(jobber.Scheduler(&goroutine.Goroutine{GrpcHost: "localhost:50051"}))
	taskMachine.RegisterGRPC(s)
```

Scheduler need to implements the following inteface:

```Go
interface {
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

`example` includes a dummy test case and also an example of using lamba scheduler.

just to see how it all works simply do the following

In one terminal run the example
```bash
$go run example/goroutine/main.go 
2018/02/11 14:48:57 Start listening gRPC at 50051
2018/02/11 14:49:17 minion: job inbound 0 0
2018/02/11 14:49:17 worker[1]: Hi, I was invoked and I am trying to connect to accept jobs
2018/02/11 14:49:17 worker[1]: I Joined the workforce
2018/02/11 14:49:17 server: A new minion joined to help
2018/02/11 14:49:17 server: got a job
2018/02/11 14:49:17 worker[1]: reveiced a task from server data:"This is the payload I will send to Lambda." 
2018/02/11 14:49:18 worker[1]: task is done
2018/02/11 14:49:18 server: received the response
2018/02/11 14:49:18 server: send the response
2018/02/11 14:49:18 server: send the response back to client
2018/02/11 14:49:18 minion: job done
2018/02/11 14:49:18 Example: Recevied data:"2018-02-11 14:49:18.623911211 +0100 CET m=+21.589500545"  <nil>
```

 In another terminal send a request
```
$ curl 'http://localhost:8000/'
2018-02-11 14:49:18.623911211 +0100 CET m=+21.589500545
```

In the code, your request goes to http server. Your handler will call Do and wait for response. Managing Lambda functions and sending and receiving message is done by jobber

```go
resp, err := myJobber.Do(&payload.Task{Data: "This is the payload I will send to Lambda."})
if err != nil {
	resp = &payload.Result{Data: "Because of error result was returned as nil"}
}
log.Println("Example: Recevied", resp, err)
fmt.Fprint(w, resp.Data)
```

## Test Lambda
If you are familiar with Lambda function, setting up one is easy.

But notice lambda functions need to connect back you the grpc server which Jobber depends on. So they must be in the same network (VPC), or somehow they need to have access you your grpc port.

You can see a simple example as `example/lambda` and a sample Lambda function which does nothing at `exmaple/aws_func`.

