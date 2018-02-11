package jobber

import (
	"context"

	"github.com/kavehmz/jobber/payload"
	"google.golang.org/grpc/metadata"
)

type dummyJoinServer struct {
	receive chan response
	send    chan error
}

func (d *dummyJoinServer) Context() context.Context     { return context.Background() }
func (d *dummyJoinServer) SendMsg(m interface{}) error  { return nil }
func (d *dummyJoinServer) RecvMsg(m interface{}) error  { return nil }
func (d *dummyJoinServer) SetHeader(metadata.MD) error  { return nil }
func (d *dummyJoinServer) SendHeader(metadata.MD) error { return nil }
func (d *dummyJoinServer) SetTrailer(metadata.MD)       {}

func (d *dummyJoinServer) Send(*payload.Task) error {
	return <-d.send
}
func (d *dummyJoinServer) Recv() (*payload.Result, error) {
	b := <-d.receive
	return b.result, b.err
}

type dummyScheduler struct{}

func (d *dummyScheduler) Inbound()  {}
func (d *dummyScheduler) Done()     {}
func (d *dummyScheduler) Timedout() {}
