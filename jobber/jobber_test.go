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
	end := make(chan bool)
	j := NewJobber(Scheduler(&dummyScheduler{}), CallTimeout(time.Millisecond*300), MaxMinionLifetime(time.Second*1))
	d := &dummyJoinServer{send: make(chan error), receive: make(chan response)}
	go func() {
		j.Join(d)
	}()

	go func() {
		r, e := j.Do(&payload.Task{})
		if e != nil {
			t.Error("All was fine so expcet no error.", e, r)
		}
		end <- true
	}()
	d.send <- nil
	d.receive <- response{result: &payload.Result{}, err: nil}
	<-end

	// wait 2 *CallTimeout to make sure call will timeout
	go func() {
		r, e := j.Do(&payload.Task{})
		if e == nil || e.Error() != "timeout" {
			t.Error("Call should have been timedout.", r)
		}
		end <- true
	}()
	time.Sleep(time.Millisecond * 600)
	d.send <- nil
	d.receive <- response{result: &payload.Result{}, err: nil}
	<-end

	// error on send
	go func() {
		_, e := j.Do(&payload.Task{})
		if e.Error() != "send_error" {
			t.Error("expected error.", e)
		}
		end <- true
	}()
	d.send <- errors.New("send_error")
	<-end

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
		end <- true
	}()
	d1.send <- nil
	d1.receive <- response{result: &payload.Result{}, err: errors.New("receive_error")}
	<-end

	j.maxMinionLifetime = time.Millisecond * 100
	// check if MaxMinionLifetime take effect
	d2 := &dummyJoinServer{send: make(chan error), receive: make(chan response)}
	go func() {
		j.Join(d2)
		end <- true
	}()

	select {
	case <-end:
	case <-time.After(time.Second * 1):
		t.Error("Join was still running even after MaxMinionLifetime")
	}
}

func TestJobber_QuitOrignore(t *testing.T) {
	end := make(chan bool)
	go func() {
		quitOrignore(nil, nil)
		end <- true
	}()

	select {
	case <-end:
	case <-time.After(time.Second * 1):
		t.Error("QuitOrignore did not ignore the nil channel")
	}
}

func TestJobber_SendOrignore(t *testing.T) {
	end := make(chan bool)
	go func() {
		sendOrignore(nil, response{})
		end <- true
	}()

	select {
	case <-end:
	case <-time.After(time.Second * 1):
		t.Error("SendOrignore did not ignore the nil channel")
	}
}
