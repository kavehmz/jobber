package jobber

import (
	"time"

	"github.com/kavehmz/jobber/payload"
	"google.golang.org/grpc"
)

type Jobber struct {
	options
	do  chan *request
	job Minion
}

type options struct {
	maxConcurrentInvitees uint32
	callTimeout           time.Duration
}

var defaultJobberOptions = options{
	maxConcurrentInvitees: 10,
	callTimeout:           time.Second,
}

type JobberOption func(*options)

func NewJobber(opt ...JobberOption) *Jobber {
	opts := defaultJobberOptions
	for _, o := range opt {
		o(&opts)
	}

	s := &Jobber{
		job:     &Goroutine{},
		options: opts,
		do:      make(chan *request),
	}
	return s
}

// CallTimeout set the timeout for every single call
func CallTimeout(t time.Duration) JobberOption {
	return func(o *options) {
		o.callTimeout = t
	}
}

// MaxConcurrentInvitees set maximum number of concurrent invitees.
func MaxConcurrentInvitees(n uint32) JobberOption {
	return func(o *options) {
		o.maxConcurrentInvitees = n
	}
}

// RegisterGRPC registers jobber service and its implementations to the gRPC
func (j *Jobber) RegisterGRPC(srv *grpc.Server) {
	payload.RegisterPayloadServer(srv, j)
}
