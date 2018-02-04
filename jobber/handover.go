package jobber

import (
	"errors"
	"io"
	"log"
	"time"

	"github.com/kavehmz/jobber/handover"
)

func (j *Jobber) Do(task *handover.Task) (*handover.Result, error) {
	payload := setPayload(task)

	go func() {
		j.job.inbound()
		j.do <- payload
	}()

	for {
		select {
		case res := <-payload.back:
			log.Println("http: received response")
			j.job.done()
			return res.result, res.err
		case <-time.After(time.Second * 1):
			j.job.timedout()
			log.Println("http: Timed out from http server side")
			return &handover.Result{}, errors.New("timeout")
		}
	}
}

func (j *Jobber) Join(stream handover.Handover_JoinServer) error {
	log.Println("Chat started")
	for {
		payload := <-j.do
		log.Println("new payload arrived")

		if err := stream.Send(payload.task); err != nil {
			log.Println("error: not able to send any message")
			payload.back <- result{result: &handover.Result{}, err: errors.New("Cant send")}
			return err
		}

		res, err := stream.Recv()
		if err == io.EOF {
			log.Println("Client send io.EOF")
			payload.back <- result{result: &handover.Result{}, err: errors.New("Client sent io.EOF")}
			return nil
		}
		if err != nil {
			log.Println("Client send error", err)
			payload.back <- result{result: &handover.Result{}, err: err}
			return err
		}
		log.Println("Client says:", res.Data)

		select {
		case payload.back <- result{result: res}:
			log.Println("Sent to http:", res.Data)
		default:
			log.Println("payload.ResCh Channel closed. Discarding the response")
		}
	}
}
