package jobber

import (
	"github.com/kavehmz/jobber/payload"
)

type response struct {
	result *payload.Result
	err    error
}
type task struct {
	task *payload.Task
	back chan response
}
