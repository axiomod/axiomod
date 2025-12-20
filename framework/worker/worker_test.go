package worker

import (
	"context"
	"testing"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/stretchr/testify/assert"
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
