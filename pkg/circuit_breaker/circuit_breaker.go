package circuit_breaker

import (
	"errors"
	"sync"
	"time"
)

type Status uint8

const (
	Closed   Status = 1
	Open     Status = 2
	HalfOpen Status = 3
)

type circuitBreaker struct {
	mu sync.Mutex
	// CLOSED - work!, OPEN - fail!, HALFOPEN - work until fail!
	state Status
	// Длинна отслеживаемого хвоста запросов
	recordLength int
	// Сколько времени у CB восстановиться
	timeout time.Duration

	lastAttemptedAt time.Time
	// Процент запросов после которого открывается CB
	percentile float64
	// Buffer хранит данные о результатах запроса
	buffer []bool
	// Pos увеличивается для каждого след запроса, потом сбрасывает в 0
	pos int
	// Сколько успешных запросов надо сделать подряд, чтобы перейти в CLOSED
	recoveryRequests int
	// Сколько успешных запросов в HALFOPEN уже сделано
	successCount int
}

type CircuitBreaker interface {
	Call(service func() error) error
	Reset()
}

func New(recordLength int, timeout time.Duration, percentile float64, recoveryRequests int) CircuitBreaker {
	return &circuitBreaker{
		state:            Closed,
		recordLength:     recordLength,
		timeout:          timeout,
		percentile:       percentile,
		buffer:           make([]bool, recordLength),
		recoveryRequests: recoveryRequests,
	}
}

var (
	ErrOpenCB = errors.New("CB IS OPEN")
)

func (cb *circuitBreaker) Call(service func() error) error {
	cb.mu.Lock()
	if cb.state == Open {
		if elapsed := time.Since(cb.lastAttemptedAt); elapsed > cb.timeout {
			// fmt.Printf("\nSWITCHING TO HALFOPEN\n")
			cb.state = HalfOpen
			cb.successCount = 0
		} else {
			cb.mu.Unlock()
			return ErrOpenCB
		}
	}
	cb.mu.Unlock()

	err := service()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.buffer[cb.pos] = err != nil
	cb.pos = (cb.pos + 1) % cb.recordLength

	if cb.state == HalfOpen {
		if err != nil {
			// fmt.Printf("\nSwitching back to open state due to an error\n")

			cb.successCount = 0
			cb.state = Open
			cb.lastAttemptedAt = time.Now()
		} else {
			cb.successCount++
			if cb.successCount > cb.recoveryRequests {
				// fmt.Printf("\nSwitching to closed state\n")
				cb.Reset()
			}
		}
		return err
	}

	// only CLOSED
	fails := 0
	for _, failed := range cb.buffer {
		if failed {
			fails++
		}
	}
	if float64(fails)/float64(cb.recordLength) >= cb.percentile {
		// 	fmt.Printf("\nSwitching to open state due to exceeding percentile\n\n")

		cb.state = Open
		cb.successCount = 0
		cb.lastAttemptedAt = time.Now()
	}

	return err
}

func (cb *circuitBreaker) Reset() {
	for i := range cb.buffer {
		cb.buffer[i] = false
	}
	// cb.buffer = make([]bool, cb.recordLength)
	cb.successCount = 0
	cb.pos = 0
	cb.state = Closed
}
