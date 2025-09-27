package gopipe

import (
	"context"
	"gopipe/internal/utils"
	"sync"
)

// Pipeline orchestrates task processing with multiple worker pools
type Pipeline struct {
	workerPools map[string]*WorkerPool
	middlewares map[string][]Middleware
	handlers    map[string]Handler // base handlers (before middleware)

	router   *Router
	metrics  *Metrics
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	submitWg *utils.SizedWaitGroup
}

// NewPipeline creates a new pipeline
func NewPipeline() *Pipeline {
	ctx, cancel := context.WithCancel(context.Background())

	return &Pipeline{
		workerPools: make(map[string]*WorkerPool),
		middlewares: make(map[string][]Middleware),
		handlers:    make(map[string]Handler),
		router:      NewRouter(),
		metrics:     NewMetrics(),
		ctx:         ctx,
		cancel:      cancel,
		submitWg:    utils.NewSizedWaitGroup(100),
	}
}

// RegisterHandler registers a handler for a task type with middlewares
func (p *Pipeline) RegisterHandler(taskType string, handler Handler, middlewares ...Middleware) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// store base handler
	p.handlers[taskType] = handler

	// build chain
	chained := Chain(middlewares...)(handler)
	wp := NewWorkerPool(10, chained)

	p.workerPools[taskType] = wp
	p.middlewares[taskType] = middlewares
	p.router.AddRoute(taskType, wp)
}

// Submit sends a task to the pipeline
func (p *Pipeline) Submit(task *Task) error {
	p.metrics.TasksSubmitted.Inc()
	p.submitWg.Add()
	defer p.submitWg.Done()

	// Route task to appropriate worker pool
	workerPool, err := p.router.Route(task)
	if err != nil {
		return err
	}
	return workerPool.Submit(task)
}

// SubmitAndWait sends a task and waits for result
func (p *Pipeline) SubmitAndWait(task *Task) TaskResult {
	if err := p.Submit(task); err != nil {
		return TaskResult{
			TaskID:  task.ID,
			Error:   err,
			Success: false,
		}
	}
	return task.WaitForResult()
}

// GetMetrics returns pipeline metrics
func (p *Pipeline) GetMetrics() *Metrics {
	return p.metrics
}

// Stop gracefully stops the pipeline
func (p *Pipeline) Stop() {
	p.cancel()

	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, wp := range p.workerPools {
		wp.Stop()
	}
}

// AddMiddleware adds global middleware to all handlers
func (p *Pipeline) AddMiddleware(middleware Middleware) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for taskType, handler := range p.handlers {
		// prepend global middleware
		p.middlewares[taskType] = append([]Middleware{middleware}, p.middlewares[taskType]...)
		chained := Chain(p.middlewares[taskType]...)(handler)

		if wp, ok := p.workerPools[taskType]; ok {
			wp.UpdateHandler(chained) // update in-place
		}
	}
}
