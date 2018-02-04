package jobber

type Minion interface {
	inbound()
	done()
	timedout()
}
