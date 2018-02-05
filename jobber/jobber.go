package jobber

import (
	"time"

	"github.com/kavehmz/jobber/payload"
	"google.golang.org/grpc"
)

type Jobber struct {
	options
	do chan task
}

type options struct {
	maxConcurrentInvitees uint32
	callTimeout           time.Duration
	job                   Minion
}

var defaultJobberOptions = options{
	job: &Goroutine{},
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
		options: opts,
		do:      make(chan task),
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

// Minion set the minion system.
func MinionScheduler(m Minion) JobberOption {
	return func(o *options) {
		o.job = m
	}
}

// RegisterGRPC registers jobber service and its implementations to the gRPC
func (j *Jobber) RegisterGRPC(srv *grpc.Server) {
	payload.RegisterPayloadServer(srv, j)
}
