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

	resp := make(chan response, 1)
	var req task
	quit := make(chan error, 1)

	// Quit when the minion is too old to rely on anymore.
	go func() {
		<-time.After(j.maxMinionLifetime)
		log.Println("server: minion is too old to reply on. returning the task back to channel")
		select {
		case quit <- nil:
		default:
		}
	}()
	for {
		go func() {
			for {
				res, err := stream.Recv()
				log.Println("server: received the response")
				resp <- response{result: res, err: err}
				if err != nil {
					log.Println("server: received an error", err)
					// signal to quit
					select {
					case quit <- err:
					default:
					}
					return
				}
			}
		}()

		select {
		case req = <-j.do:
			log.Println("server: got a job")
			if err := stream.Send(req.task); err != nil {
				log.Println("server: not able to send any message", err)
				j.job.Inbound()
				return err
			}
			r := <-resp
			select {
			case req.back <- r:
				log.Println("server: send the response back to client")
			default:
				log.Println("server: channel closed. discarding the response")
			}
		case err := <-quit:
			return err
		}
	}
}
