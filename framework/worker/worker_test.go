package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorker(t *testing.T) {
	// Setup logger
	cfg := &config.Config{}
	logger, _ := observability.NewLogger(cfg)

	w := New(logger)

	t.Run("Register and Start Job", func(t *testing.T) {
		jobChan := make(chan bool, 1)
		job := &Job{
			ID:       "test-job",
			Name:     "Test Job",
			Interval: 100 * time.Millisecond,
			Func: func(ctx context.Context) error {
				jobChan <- true
				return nil
			},
		}

		err := w.RegisterJob(job)
		assert.NoError(t, err)

		err = w.StartJob("test-job")
		assert.NoError(t, err)

		// Wait for job to execute
		select {
		case <-jobChan:
			// Success
		case <-time.After(500 * time.Millisecond):
			t.Fatal("Job did not execute in time")
		}

		err = w.StopJob("test-job")
		assert.NoError(t, err)
	})

	t.Run("Job Timeout", func(t *testing.T) {
		job := &Job{
			ID:       "timeout-job",
			Name:     "Timeout Job",
			Interval: 1 * time.Second,
			Timeout:  100 * time.Millisecond,
			Func: func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(500 * time.Millisecond):
					return nil
				}
			},
		}

		err := w.RegisterJob(job)
		assert.NoError(t, err)

		err = w.StartJob("timeout-job")
		assert.NoError(t, err)

		// Give it some time to run and timeout
		time.Sleep(200 * time.Millisecond)

		err = w.StopJob("timeout-job")
		assert.NoError(t, err)
	})

	t.Run("Stop All", func(t *testing.T) {
		w.StopAll()
		assert.Empty(t, w.cancelFunc)
	})
}

func TestWorkerErrors(t *testing.T) {
	cfg := &config.Config{}
	logger, _ := observability.NewLogger(cfg)
	w := New(logger)

	t.Run("Register Invalid Job", func(t *testing.T) {
		err := w.RegisterJob(&Job{ID: ""})
		assert.Error(t, err)
		assert.Equal(t, "job ID cannot be empty", err.Error())

		err = w.RegisterJob(&Job{ID: "valid", Func: nil})
		assert.Error(t, err)
		assert.Equal(t, "job function cannot be nil", err.Error())
	})

	t.Run("Start Non-Existent Job", func(t *testing.T) {
		err := w.StartJob("missing")
		assert.Error(t, err)
		assert.Equal(t, ErrJobNotFound, err)
	})

	t.Run("Stop Non-Existent Job", func(t *testing.T) {
		err := w.StopJob("missing")
		assert.Error(t, err)
		assert.Equal(t, ErrJobNotFound, err)
	})

	t.Run("Start Already Running Job", func(t *testing.T) {
		job := &Job{
			ID:       "running",
			Name:     "Running Job",
			Interval: 1 * time.Hour,
			Func:     func(ctx context.Context) error { return nil },
		}
		_ = w.RegisterJob(job)
		_ = w.StartJob("running")
		err := w.StartJob("running")
		assert.NoError(t, err) // Should be no-op/nil
		_ = w.StopJob("running")
	})
}

func newTestWorker(t *testing.T) *Worker {
	t.Helper()
	logger, err := observability.NewLogger(&config.Config{})
	require.NoError(t, err)
	return New(logger)
}

func TestWorkerJobTimeout(t *testing.T) {
	w := newTestWorker(t)

	var sawDeadline atomic.Bool
	job := &Job{
		ID:       "slow",
		Name:     "Slow Job",
		Interval: time.Hour, // only the immediate run matters
		Timeout:  30 * time.Millisecond,
		Func: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				sawDeadline.Store(true)
				return ctx.Err()
			case <-time.After(2 * time.Second):
				return nil
			}
		},
	}

	require.NoError(t, w.RegisterJob(job))
	require.NoError(t, w.StartJob(job.ID))

	assert.Eventually(t, func() bool { return sawDeadline.Load() }, 2*time.Second, 10*time.Millisecond,
		"job context must be cancelled by the per-job timeout")
	require.NoError(t, w.StopJob(job.ID))
}

func TestWorkerLifecycleErrors(t *testing.T) {
	w := newTestWorker(t)

	t.Run("start unknown job", func(t *testing.T) {
		assert.ErrorIs(t, w.StartJob("missing"), ErrJobNotFound)
	})

	t.Run("stop unknown job", func(t *testing.T) {
		assert.ErrorIs(t, w.StopJob("missing"), ErrJobNotFound)
	})

	t.Run("stop job that is not running", func(t *testing.T) {
		job := &Job{ID: "idle", Name: "Idle", Interval: time.Hour,
			Func: func(ctx context.Context) error { return nil }}
		require.NoError(t, w.RegisterJob(job))
		// Stopping a registered-but-not-started job must not panic; any
		// returned error is acceptable as long as the worker stays usable.
		_ = w.StopJob(job.ID)
		assert.ErrorIs(t, w.StartJob("still-missing"), ErrJobNotFound)
	})
}

func TestWorkerStopAllAndShutdown(t *testing.T) {
	w := newTestWorker(t)

	var runs atomic.Int32
	for _, id := range []string{"a", "b"} {
		job := &Job{
			ID: id, Name: id, Interval: 20 * time.Millisecond, Timeout: time.Second,
			Func: func(ctx context.Context) error { runs.Add(1); return nil },
		}
		require.NoError(t, w.RegisterJob(job))
		require.NoError(t, w.StartJob(job.ID))
	}

	assert.Eventually(t, func() bool { return runs.Load() >= 2 }, 2*time.Second, 10*time.Millisecond)

	w.StopAll()
	settled := runs.Load()
	time.Sleep(80 * time.Millisecond)
	assert.LessOrEqual(t, runs.Load(), settled+2, "jobs must stop ticking after StopAll")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	assert.NoError(t, w.Shutdown(ctx))
}
