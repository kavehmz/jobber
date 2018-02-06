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
	joinTime := time.Now()
	for {
		request := <-j.do
		log.Println("server: new request arrived. Sending the task to the minion")
		if time.Since(joinTime) > j.maxMinionLifetime {
			log.Println("server: minion is too old to reply on. returning the task back to channel")
			j.do <- request
			return nil
		}

		if err := stream.Send(request.task); err != nil {
			log.Println("server: not able to send any message", err)
			// undelivered message goes back to the queue
			j.do <- request
			return err
		}

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
