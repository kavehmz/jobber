# jobber

Jobber is an idea and a sample implementation.

# Background
Past - Dealing with rigid barematal servers. minimum granuality : -
Recently - Setting up a auto-scaling enviroment using Kubernetes or other tools to scale up and down based on incoming requersts.
Today - Using Lambda or Google functions I can scale up to 10,000 cpu in few seconds and then scale down in no time.ing 

But for both AWS Lambda and Google cloud functions the catch it their mininum 100ms granuality. 
It means if I want to serve my http requests there and they only take 10ms I will pay 10x more 
because both platforms will charge me for 100ms. Based on current prices that is expensive.

# Solution
Here the idea is to invoke a Lambda function but instead of asking it to do one reuqest we will keep it around to serve many more.

Method is easy.

We need to define what is the format of payload we want to send to Lambda `payload directory`.

We start our http server along with a grpc server `cmd directory`.

When http requests or any other micro job comes, we an scheduler invokes Lambda functions. Lamdba functions connect to the server (Bidirectional streaming).

Server starts to send stream of tasks to Lambda.

Lambda stops at 300s (platform limit) and now that 100ms resolution won't look too bad.

![flow](https://kavehmz.github.io/static/images/lambda_for_micro_jobs.png "Lambda for micro requests")

This solution does not depend on protobuf, grpc or aws lambda but in this implementation I picked those tools,

# Test run

`example` includes a dummy test case. It invokes Goroutines instead of Lambda but concept is identical and invoking lambda fucntions in just few lines of code.



