package jobber

import (
	"errors"
	"io"
	"log"
	"time"

	"github.com/kavehmz/jobber/payload"
)

func (j *Jobber) Do(task *payload.Task) (*payload.Result, error) {
	request := setRequest(task)

	go func() {
		j.job.Inbound()
		j.do <- request
	}()

	for {
		select {
		case res := <-request.back:
			log.Println("http: received response")
			j.job.Done()
			return res.result, res.err
		case <-time.After(time.Second * 1):
			j.job.Timedout()
			log.Println("http: Timed out from http server side")
			return &payload.Result{}, errors.New("timeout")
		}
	}
}

func (j *Jobber) Join(stream payload.Payload_JoinServer) error {
	log.Println("Chat started")
	for {
		request := <-j.do
		log.Println("new request arrived")

		if err := stream.Send(request.task); err != nil {
			log.Println("error: not able to send any message")
			request.back <- response{result: &payload.Result{}, err: errors.New("Cant send")}
			return err
		}

		res, err := stream.Recv()
		if err == io.EOF {
			log.Println("Client send io.EOF")
			request.back <- response{result: &payload.Result{}, err: errors.New("Client sent io.EOF")}
			return nil
		}
		if err != nil {
			log.Println("Client send error", err)
			request.back <- response{result: &payload.Result{}, err: err}
			return err
		}
		log.Println("Client says:", res.Data)

		select {
		case request.back <- response{result: res}:
			log.Println("Sent to http:", res.Data)
		default:
			log.Println("request.ResCh Channel closed. Discarding the response")
		}
	}
}
