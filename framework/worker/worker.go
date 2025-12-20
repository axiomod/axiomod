package worker

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/axiomod/axiomod/platform/observability"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides the fx options for the worker module
var Module = fx.Options(
	fx.Provide(New),
	fx.Invoke(RegisterWorker),
)

// RegisterWorker registers the worker with the fx lifecycle
func RegisterWorker(lc fx.Lifecycle, w *Worker) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			w.StopAll()
			return nil
		},
	})
}

// Common errors
var (
	ErrWorkerStopped = errors.New("worker has been stopped")
	ErrJobNotFound   = errors.New("job not found")
)

// Job represents a background job
type Job struct {
	ID       string
	Name     string
	Func     func(ctx context.Context) error
	Interval time.Duration
	Timeout  time.Duration
}

// Worker manages background jobs
type Worker struct {
	jobs       map[string]*Job
	cancelFunc map[string]context.CancelFunc
	mu         sync.RWMutex
	logger     *observability.Logger
}

// New creates a new Worker
func New(logger *observability.Logger) *Worker {
	return &Worker{
		jobs:       make(map[string]*Job),
		cancelFunc: make(map[string]context.CancelFunc),
		logger:     logger,
	}
}

// RegisterJob registers a new job
func (w *Worker) RegisterJob(job *Job) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if job.ID == "" {
		return errors.New("job ID cannot be empty")
	}

	if job.Func == nil {
		return errors.New("job function cannot be nil")
	}

	w.jobs[job.ID] = job
	w.logger.Info("Registered job", zap.String("id", job.ID), zap.String("name", job.Name))
	return nil
}

// StartJob starts a job
func (w *Worker) StartJob(jobID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	job, exists := w.jobs[jobID]
	if !exists {
		return ErrJobNotFound
	}

	// Check if job is already running
	if _, running := w.cancelFunc[jobID]; running {
		return nil // Job is already running
	}

	// Create a context with cancel function
	ctx, cancel := context.WithCancel(context.Background())
	w.cancelFunc[jobID] = cancel

	// Start the job
	go w.runJob(ctx, job)

	w.logger.Info("Started job", zap.String("id", job.ID), zap.String("name", job.Name))
	return nil
}

// StopJob stops a job
func (w *Worker) StopJob(jobID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	cancel, exists := w.cancelFunc[jobID]
	if !exists {
		return ErrJobNotFound
	}

	// Cancel the job
	cancel()
	delete(w.cancelFunc, jobID)

	w.logger.Info("Stopped job", zap.String("id", jobID))
	return nil
}

// StopAll stops all jobs
func (w *Worker) StopAll() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for jobID, cancel := range w.cancelFunc {
		cancel()
		delete(w.cancelFunc, jobID)
		w.logger.Info("Stopped job", zap.String("id", jobID))
	}
}

// runJob runs a job at the specified interval
func (w *Worker) runJob(ctx context.Context, job *Job) {
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	// Run the job immediately
	w.executeJob(ctx, job)

	// Run the job at the specified interval
	for {
		select {
		case <-ticker.C:
			w.executeJob(ctx, job)
		case <-ctx.Done():
			w.logger.Info("Job context canceled", zap.String("id", job.ID), zap.String("name", job.Name))
			return
		}
	}
}

// executeJob executes a job with timeout
func (w *Worker) executeJob(ctx context.Context, job *Job) {
	w.logger.Debug("Executing job", zap.String("id", job.ID), zap.String("name", job.Name))

	// Create a context with timeout
	jobCtx := ctx
	if job.Timeout > 0 {
		var cancel context.CancelFunc
		jobCtx, cancel = context.WithTimeout(ctx, job.Timeout)
		defer cancel()
	}

	// Execute the job
	if err := job.Func(jobCtx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			w.logger.Error("Job timed out", zap.String("id", job.ID), zap.String("name", job.Name), zap.Duration("timeout", job.Timeout))
		} else {
			w.logger.Error("Job failed", zap.String("id", job.ID), zap.String("name", job.Name), zap.Error(err))
		}
	} else {
		w.logger.Debug("Job completed successfully", zap.String("id", job.ID), zap.String("name", job.Name))
	}
}
