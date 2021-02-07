package internal

import "time"

type Semaphore struct {
	permits int
	channel chan int
}

func NewSemaphore(permits int) *Semaphore {
	return &Semaphore{channel: make(chan int, permits), permits: permits}
}

func (s *Semaphore) Acquire() {
	s.channel <- 0
}

func (s *Semaphore) Release() {
	<-s.channel
}

func (s *Semaphore) TryAcquire() bool {
	select {
	case s.channel <- 0:
		return true
	default:
		return false
	}
}

func (s *Semaphore) TryAcquireOnTime(timeout time.Duration) bool {
	for {
		select {
		case s.channel <- 0:
			return true
		case <-time.After(timeout):
			return false
		}
	}
}

func (s *Semaphore) AvailablePermits() int {
	return s.permits - len(s.channel)
}
