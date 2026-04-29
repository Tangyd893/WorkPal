package main

import (
	"sync"
	"time"
)

const (
	breakerClosed   = "closed"
	breakerOpen     = "open"
	breakerHalfOpen = "half_open"
)

type circuitBreaker struct {
	mu                sync.Mutex
	state             string
	failureThreshold  int
	coolDown          time.Duration
	consecutiveErrors int
	openedAt          time.Time
	probeInFlight     bool
	now               func() time.Time
}

type circuitBreakerSnapshot struct {
	State               string `json:"state"`
	FailureThreshold    int    `json:"failure_threshold"`
	ConsecutiveFailures int    `json:"consecutive_failures"`
	CoolDownMS          int64  `json:"cool_down_ms"`
	ProbeInFlight       bool   `json:"probe_in_flight"`
	NextRetryAt         string `json:"next_retry_at,omitempty"`
}

func newCircuitBreaker(failureThreshold int, coolDown time.Duration) *circuitBreaker {
	return &circuitBreaker{
		state:            breakerClosed,
		failureThreshold: maxInt(1, failureThreshold),
		coolDown:         coolDown,
		now:              time.Now,
	}
}

func (b *circuitBreaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case breakerClosed:
		return true
	case breakerOpen:
		if b.coolDown <= 0 || b.now().After(b.openedAt.Add(b.coolDown)) {
			b.state = breakerHalfOpen
			b.probeInFlight = true
			return true
		}
		return false
	case breakerHalfOpen:
		if b.probeInFlight {
			return false
		}
		b.probeInFlight = true
		return true
	default:
		b.state = breakerClosed
		return true
	}
}

func (b *circuitBreaker) OnSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.state = breakerClosed
	b.probeInFlight = false
	b.consecutiveErrors = 0
	b.openedAt = time.Time{}
}

func (b *circuitBreaker) OnFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch b.state {
	case breakerHalfOpen:
		b.trip()
		return
	case breakerOpen:
		b.trip()
		return
	default:
		b.consecutiveErrors++
		if b.consecutiveErrors >= b.failureThreshold {
			b.trip()
		}
	}
}

func (b *circuitBreaker) Snapshot() circuitBreakerSnapshot {
	b.mu.Lock()
	defer b.mu.Unlock()

	snapshot := circuitBreakerSnapshot{
		State:               b.state,
		FailureThreshold:    b.failureThreshold,
		ConsecutiveFailures: b.consecutiveErrors,
		CoolDownMS:          b.coolDown.Milliseconds(),
		ProbeInFlight:       b.probeInFlight,
	}
	if b.state == breakerOpen && !b.openedAt.IsZero() {
		snapshot.NextRetryAt = b.openedAt.Add(b.coolDown).Format(time.RFC3339)
	}
	return snapshot
}

func (b *circuitBreaker) trip() {
	b.state = breakerOpen
	b.openedAt = b.now()
	b.probeInFlight = false
	if b.consecutiveErrors < b.failureThreshold {
		b.consecutiveErrors = b.failureThreshold
	}
}
