package jobber

import (
	"github.com/kavehmz/jobber/payload"
)

type response struct {
	result *payload.Result
	err    error
}
type request struct {
	task *payload.Task
	back chan response
}

func setRequest(t *payload.Task) *request {
	return &request{task: t, back: make(chan response)}
}
