package jobber

import (
	"errors"
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
	resp := make(chan response)
	var req task
	quit := make(chan error)
	for {

		go func() {
			for {
				res, err := stream.Recv()
				if err != nil {
					log.Println("server: received an error", err)
					res = &payload.Result{}
					quit <- err
				}
				log.Println("server: received the response")

				resp <- response{result: res, err: err}

				if err != nil {
					return
				}

			}
		}()

		// quit has prioiry
		select {
		case err := <-quit:
			return err
		default:
		}
		select {
		case req = <-j.do:
			log.Println("server: got a job")
			if err := stream.Send(req.task); err != nil {
				log.Println("server: not able to send any message", err)
				j.job.Inbound()
				j.do <- req
				return err
			}
			log.Println("server: send done")
			r := <-resp
			log.Println("server: got answer")
			select {
			case req.back <- r:
				log.Println("server: send the response back to client")
			default:
				log.Println("server: channel closed. discarding the response")
			}

		case err := <-quit:
			return err
		case <-time.After(j.maxMinionLifetime):
			log.Println("server: minion is too old to reply on. returning the task back to channel")
			return nil
		}

	}
}
