package jobber

import (
	"github.com/kavehmz/jobber/handover"
)

type result struct {
	result *handover.Result
	err    error
}
type payload struct {
	task *handover.Task
	back chan result
}

func setPayload(t *handover.Task) *payload {
	return &payload{task: t, back: make(chan result)}
}
