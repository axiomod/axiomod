package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the state of the circuit breaker
type State int

const (
	// StateClosed means the circuit breaker is closed and requests are allowed
	StateClosed State = iota
	// StateOpen means the circuit breaker is open and requests are not allowed
	StateOpen
	// StateHalfOpen means the circuit breaker is half-open and a limited number of requests are allowed
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name          string
	maxFailures   int
	resetTimeout  time.Duration
	 halfOpenLimit int // Number of successful requests required in half-open to close
	state         State
	 failures      int // Current consecutive failures in closed state
	 halfOpenCount int // Current successful requests in half-open state
	lastFailure   time.Time
	mutex         sync.RWMutex // Changed back to RWMutex for State() read optimization
}

// Options contains options for creating a new CircuitBreaker
type Options struct {
	// Name is the name of the circuit breaker
	Name string
	// MaxFailures is the number of failures that will trip the circuit breaker
	MaxFailures int
	// ResetTimeout is the time to wait before transitioning from open to half-open
	ResetTimeout time.Duration
	// HalfOpenLimit is the number of successful requests required in half-open state to close the circuit
	HalfOpenLimit int
}

// DefaultOptions returns the default options for a circuit breaker
func DefaultOptions() Options {
	return Options{
		Name:          "default",
		MaxFailures:   5,
		ResetTimeout:  10 * time.Second,
		HalfOpenLimit: 1, // Default: 1 successful request closes the circuit
	}
}

// New creates a new CircuitBreaker with the given options
func New(options Options) *CircuitBreaker {
	// Ensure HalfOpenLimit is at least 1
	 halfOpenLimit := options.HalfOpenLimit
	 if halfOpenLimit < 1 {
		 halfOpenLimit = 1
	 }
	return &CircuitBreaker{
		name:          options.Name,
		maxFailures:   options.MaxFailures,
		resetTimeout:  options.ResetTimeout,
		 halfOpenLimit: halfOpenLimit,
		state:         StateClosed,
	}
}

// Execute executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	 if !cb.AllowRequest() {
		 return errors.New("circuit breaker is open")
	 }

	 err := fn()
	 cb.RecordResult(err)
	 return err
}

// AllowRequest checks if a request is allowed based on the current state
// It handles state transitions from Open to HalfOpen.
func (cb *CircuitBreaker) AllowRequest() bool {
	 cb.mutex.Lock() // Use write lock as state transitions might occur
	 defer cb.mutex.Unlock()

	 now := time.Now()
	 state := cb.state

	 switch state {
	 case StateClosed:
		 return true
	 case StateOpen:
		 // Check if reset timeout has elapsed
		 if cb.lastFailure.IsZero() || now.Sub(cb.lastFailure) > cb.resetTimeout {
			 // Transition to half-open state
			 cb.state = StateHalfOpen
			 cb.halfOpenCount = 0 // Reset success counter for half-open
			 return true // Allow the first request in half-open
		 }
		 // Timeout not elapsed, still open
		 return false
	 case StateHalfOpen:
		 // Allow requests up to the limit. The actual counting happens in RecordResult.
		 // This check might seem redundant if RecordResult handles the state change, but it prevents
		 // excessive requests if RecordResult is slow or fails to be called.
		 // A simpler approach might be to always allow in HalfOpen and let RecordResult manage state.
		 // Let's allow and rely on RecordResult.
		 return true
	 default:
		 return false // Should not happen
	 }
}

// RecordResult records the result of a request and handles state transitions
func (cb *CircuitBreaker) RecordResult(err error) {
	 cb.mutex.Lock()
	 defer cb.mutex.Unlock()

	 now := time.Now()

	 switch cb.state {
	 case StateClosed:
		 if err != nil {
			 cb.failures++
			 if cb.failures >= cb.maxFailures {
				 // Trip the circuit breaker
				 cb.state = StateOpen
				 cb.lastFailure = now
			 }
		 } else {
			 // Reset failures on success
			 cb.failures = 0
		 }
	 case StateHalfOpen:
		 if err != nil {
			 // Failure in half-open state, transition back to open
			 cb.state = StateOpen
			 cb.lastFailure = now
		 } else {
			 // Success in half-open state
			 cb.halfOpenCount++
			 // Check if enough successful requests passed to close the circuit
			 if cb.halfOpenCount >= cb.halfOpenLimit {
				 cb.state = StateClosed
				 cb.failures = 0 // Reset failure count
			 }
		 }
	 // No action needed if StateOpen, as requests shouldn't reach RecordResult then.
	 }
}

// Reset resets the circuit breaker to the closed state
func (cb *CircuitBreaker) Reset() {
	 cb.mutex.Lock()
	 defer cb.mutex.Unlock()

	 cb.state = StateClosed
	 cb.failures = 0
	 cb.halfOpenCount = 0
	 cb.lastFailure = time.Time{} // Reset last failure time
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	 cb.mutex.RLock()
	 defer cb.mutex.RUnlock()
	 // Need to check if open state should transition to half-open based on time
	 // This makes State() potentially modify state, which is bad practice for a read-only func.
	 // Let's keep State() purely read-only and rely on AllowRequest/RecordResult for transitions.
	 return cb.state

	 /* Alternative: Check for transition in State() - requires write lock potentially
	 cb.mutex.Lock()
	 defer cb.mutex.Unlock()
	 if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.resetTimeout {
		 cb.state = StateHalfOpen
		 cb.halfOpenCount = 0
	 }
	 return cb.state
	 */
}

// Name returns the name of the circuit breaker
func (cb *CircuitBreaker) Name() string {
	 return cb.name
}

