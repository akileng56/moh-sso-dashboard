package reportscheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// WorkerService is the scheduler's background execution boundary.
type WorkerService interface {
	UpdateWorkerHeartbeat(context.Context, string, time.Duration, time.Time, *time.Time, *time.Time, string) error
	PollGenerating(context.Context, int) error
	ProcessOneRetry(context.Context) (bool, error)
	ProcessOneDue(context.Context) (bool, error)
	ProcessOneDelivery(context.Context) (bool, error)
}

type Worker struct {
	mu       sync.Mutex
	started  bool
	service  WorkerService
	interval time.Duration
	log      func(args ...any)
	workerID string
}

func NewWorker(service WorkerService, interval time.Duration, log func(args ...any)) *Worker {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Worker{service: service, interval: interval, log: log, workerID: uuid.NewString()}
}

// begin checks persistence readiness before any execution is claimed. A worker
// instance may start once; re-enabling a module constructs a fresh registry.
func (w *Worker) begin(ctx context.Context) error {
	if w == nil || w.service == nil {
		return errors.New("report scheduler worker service is required")
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.started {
		return errors.New("report scheduler worker already started")
	}
	w.started = true
	readyCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := readyCtx.Err(); err != nil {
		return err
	}
	if err := w.service.UpdateWorkerHeartbeat(readyCtx, w.workerID, w.interval, time.Now().UTC(), nil, nil, ""); err != nil {
		return fmt.Errorf("report scheduler persistence readiness: %w", err)
	}
	return readyCtx.Err()
}

// Initialize returns after readiness. Cleanup cancels in-flight operations and
// joins the worker before the host closes its database and delivery resources.
func (w *Worker) Initialize(ctx context.Context) (func(context.Context) error, error) {
	workerCtx, cancel := context.WithCancel(ctx)
	if err := w.begin(workerCtx); err != nil {
		cancel()
		return nil, err
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = w.run(workerCtx)
	}()
	return func(shutdownCtx context.Context) error {
		cancel()
		select {
		case <-done:
			return nil
		default:
		}
		select {
		case <-done:
			return nil
		case <-shutdownCtx.Done():
			return shutdownCtx.Err()
		}
	}, nil
}

// Start is the blocking form for standalone callers.
func (w *Worker) Start(ctx context.Context) error {
	if err := w.begin(ctx); err != nil {
		return err
	}
	return w.run(ctx)
}

func (w *Worker) run(ctx context.Context) error {
	// Process immediately on startup so schedules due during a restart do not
	// wait for the first polling interval.
	w.process(ctx)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *Worker) process(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	started := time.Now().UTC()
	_ = w.service.UpdateWorkerHeartbeat(ctx, w.workerID, w.interval, started, &started, nil, "")
	cycleErr := ""
	defer func() {
		finished := time.Now().UTC()
		_ = w.service.UpdateWorkerHeartbeat(ctx, w.workerID, w.interval, finished, nil, &finished, cycleErr)
	}()
	if err := w.service.PollGenerating(ctx, 50); err != nil {
		cycleErr = err.Error()
		w.logf("report scheduler: poll Health BI jobs failed: ", err)
	}
	for processed := 0; processed < 100; processed++ {
		if ctx.Err() != nil {
			return
		}
		found, err := w.service.ProcessOneRetry(ctx)
		if err != nil {
			cycleErr = err.Error()
			w.logf("report scheduler: processing generation retry failed: ", err)
			return
		}
		if !found {
			break
		}
	}
	for processed := 0; processed < 100; processed++ {
		if ctx.Err() != nil {
			return
		}
		found, err := w.service.ProcessOneDue(ctx)
		if err != nil {
			cycleErr = err.Error()
			w.logf("report scheduler: processing due schedule failed: ", err)
			return
		}
		if !found {
			break
		}
	}
	for processed := 0; processed < 200; processed++ {
		if ctx.Err() != nil {
			return
		}
		found, err := w.service.ProcessOneDelivery(ctx)
		if err != nil {
			cycleErr = err.Error()
			w.logf("report scheduler: processing delivery failed: ", err)
			return
		}
		if !found {
			return
		}
	}
}

func (w *Worker) logf(args ...any) {
	if w.log != nil {
		w.log(args...)
	}
}
