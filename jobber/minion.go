package jobber

type Minion interface {
	// Inbound is called before a new task is added.
	Inbound()
	// Done is called when a task is done
	Done()
	// Timedout is called when no response was received on time for a task
	Timedout()
}
