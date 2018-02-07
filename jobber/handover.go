package jobber

import (
	"errors"
	"io"
	"log"
	"time"

	"github.com/kavehmz/jobber/payload"
)

func (j *Jobber) Do(t *payload.Task) (*payload.Result, error) {
	r := task{t, make(chan response)}
	go func() {
		j.job.Inbound()
		j.do <- r
	}()
	for {
		select {
		case res := <-r.back:
			j.job.Done()
			return res.result, res.err
		case <-time.After(j.callTimeout):
			j.job.Timedout()
			return &payload.Result{}, errors.New("timeout")
		}
	}
}

func (j *Jobber) Join(stream payload.Payload_JoinServer) error {
	log.Println("server: A new minion joined to help")
	putCh := make(chan task)
	getCh := make(chan response)
	quitCh := make(chan error)
	quitSend := make(chan error)
	for {
		select {
		case putCh := <-j.do:
		case resp := <-getCh:
		case err := <-quitCh:
			return err
		case <-time.After(j.maxMinionLifetime):
			log.Println("server: minion is too old to reply on. returning the task back to channel")
		}

		go func() {
			select {
			case r := <-putCh:
				if err := stream.Send(r.task); err != nil {
					log.Println("server: not able to send any message", err)
					r.back <- response{result: &payload.Result{}, err: err}
					quitCh <- err
				}
			case <-quitSend:
				return
			}

		}()

		res, err := stream.Recv()
		if err == io.EOF {
			log.Println("server: received io.EOF")
			request.back <- response{result: &payload.Result{}, err: err}
			return nil
		}
		if err != nil {
			log.Println("server: received error", err)
			request.back <- response{result: &payload.Result{}, err: err}
			return err
		}
		log.Println("server: received the response")

		select {
		case request.back <- response{result: res}:
			log.Println("server: send the response back to client")
		default:
			log.Println("server: channel closed. discarding the response")
		}
	}
}
