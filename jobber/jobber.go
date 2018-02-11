// Package jobber provides functionality invoke worker and ask them to join
// back (e.x by grpc) in a bidirectional way.
// This way instead of one request, they caller and continuously send a stream
// of requests to worker.
package jobber

import (
	"time"

	"github.com/kavehmz/jobber/payload"
	"google.golang.org/grpc"
)

// Jobber defines main attributes of caller.
type Jobber struct {
	options
	do chan task
}

type options struct {
	callTimeout           time.Duration
	scheduler             Minion
	maxConcurrentInvitees uint32
	maxWaitingList        uint32
	maxMinionLifetime     time.Duration
}

var defaultOptions = options{
	callTimeout:           time.Second * 3,
	maxConcurrentInvitees: 10,
	maxWaitingList:        100,
	maxMinionLifetime:     time.Second * 12,
}

// Option defines the options type to pass to NewJobber
type Option func(*options)

/*NewJobber return an Jobber with specific settings.

  myJobber = jobber.NewJobber(jobber.Scheduler(
  	&awslambda.LambdaScheduler{
  		GrpcHost: os.Getenv("GRPC_HOST") + ":50051",
  		Lambda:   fn,
  		// Set the rate of calling Lambda to 1 calls a second
  		Limiter: rate.NewLimiter(rate.Limit(1), 1),
  		Ctx:     context.Background(),
  	}),
  	jobber.MaxMinionLifetime(time.Second*13),
  )
  myJobber.RegisterGRPC(s)

*/
func NewJobber(opt ...Option) *Jobber {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}

	s := &Jobber{
		options: opts,
		do:      make(chan task),
	}
	return s
}

// CallTimeout set the timeout for every single call
func CallTimeout(t time.Duration) Option {
	return func(o *options) {
		o.callTimeout = t
	}
}

// MaxConcurrentInvitees set maximum number of concurrent invitees.
func MaxConcurrentInvitees(n uint32) Option {
	return func(o *options) {
		o.maxConcurrentInvitees = n
	}
}

// Scheduler set the scheduler system.
func Scheduler(s Minion) Option {
	return func(o *options) {
		o.scheduler = s
	}
}

// MaxWaitingList set buffer size for tasks channel
func MaxWaitingList(n uint32) Option {
	return func(o *options) {
		o.maxWaitingList = n
	}
}

// MaxMinionLifetime set how long server can rely on a minion to sent tasks.
// Basically some implmenetations like Lambda function have maximum age (300s)
// In these cases, based on load type, it might work better if server disconnected on its own terms.
func MaxMinionLifetime(d time.Duration) Option {
	return func(o *options) {
		o.maxMinionLifetime = d
	}
}

// RegisterGRPC registers jobber service and its implementations to the gRPC
func (j *Jobber) RegisterGRPC(srv *grpc.Server) {
	payload.RegisterPayloadServer(srv, j)
}
