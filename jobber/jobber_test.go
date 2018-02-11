package jobber

import (
	"errors"
	"testing"
	"time"

	"github.com/kavehmz/jobber/payload"
	"google.golang.org/grpc"
)

func TestJobber(t *testing.T) {
	s := &dummyScheduler{}
	j := NewJobber(CallTimeout(time.Second), MaxConcurrentInvitees(1), MaxWaitingList(1), MaxMinionLifetime(time.Second), Scheduler(s))
	if j.callTimeout != time.Second {
		t.Error("wrong callTimeout.")
	}
	if j.maxConcurrentInvitees != 1 {
		t.Error("wrong callTimeout.")
	}
	if j.maxWaitingList != 1 {
		t.Error("wrong callTimeout.")
	}
	if j.scheduler != s {
		t.Error("wrong scheduler.")
	}

	g := grpc.NewServer()
	j.RegisterGRPC(g)
}

func TestJobber_Join(t *testing.T) {
	j := NewJobber(Scheduler(&dummyScheduler{}), CallTimeout(time.Millisecond*300))
	d := &dummyJoinServer{send: make(chan error), receive: make(chan response)}
	go func() {
		j.Join(d)
	}()

	go func() {
		r, e := j.Do(&payload.Task{})
		if e != nil {
			t.Error("All was fine so expcet no error.", e, r)
		}
	}()
	d.send <- nil
	d.receive <- response{result: &payload.Result{}, err: nil}

	// wait 2 *CallTimeout to make sure call will timeout
	go func() {
		r, e := j.Do(&payload.Task{})
		if e == nil || e.Error() != "timeout" {
			t.Error("Call should have been timedout.", r)
		}
	}()
	time.Sleep(time.Millisecond * 600)
	d.send <- nil
	d.receive <- response{result: &payload.Result{}, err: nil}
	// error on send
	go func() {
		_, e := j.Do(&payload.Task{})
		if e.Error() != "send_error" {
			t.Error("expected error.", e)
		}
	}()
	d.send <- errors.New("send_error")
	// d.receive <- response{result: &payload.Result{}, err: nil}

	// function will exit on send error
	d1 := &dummyJoinServer{send: make(chan error), receive: make(chan response)}
	go func() {
		j.Join(d1)
	}()

	// error on receive
	go func() {
		_, e := j.Do(&payload.Task{})
		if e.Error() != "receive_error" {
			t.Error("expected error.", e)
		}
	}()
	d1.send <- nil
	d1.receive <- response{result: &payload.Result{}, err: errors.New("receive_error")}

	time.Sleep(time.Second)

}
