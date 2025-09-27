package gopipe

import (
	"context"
	"sync"

	"github.com/mani-s-tiwari/gopipe/internal/utils"
)

// ExecutionMode determines how the pipeline runs
type ExecutionMode int

const (
	ModeWorkerPool ExecutionMode = iota
	ModeActor
	ModeGossip
	ModeManager
)

// Pipeline orchestrates task processing with multiple strategies
type Pipeline struct {
	mode        ExecutionMode
	workerPools map[string]*WorkerPool
	middlewares map[string][]Middleware
	handlers    map[string]Handler // base handlers

	actorSystem   *ActorSystem
	gossipCluster *GossipCluster
	manager       *ManagerWorker

	router      *Router
	metrics     *Metrics
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	submitWg    *utils.SizedWaitGroup
	connections []*Pipeline // 🔗 downstream pipelines
	rrIndex     int         // round-robin index for forwarding
}

// --- Constructors ---
func NewPipeline() *Pipeline {
	return NewPipelineWithMode(ModeWorkerPool)
}

func NewPipelineWithMode(mode ExecutionMode) *Pipeline {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pipeline{
		mode:        mode,
		workerPools: make(map[string]*WorkerPool),
		middlewares: make(map[string][]Middleware),
		handlers:    make(map[string]Handler),
		router:      NewRouter(),
		metrics:     NewMetrics(),
		ctx:         ctx,
		cancel:      cancel,
		submitWg:    utils.NewSizedWaitGroup(100),
		connections: []*Pipeline{},
		actorSystem: NewActorSystem(),
	}
}

// --- Handler Registration ---
func (p *Pipeline) RegisterHandler(taskType string, handler Handler, middlewares ...Middleware) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers[taskType] = handler
	chained := Chain(middlewares...)(handler)

	switch p.mode {
	case ModeWorkerPool:
		wp := NewWorkerPool(10, chained)
		p.workerPools[taskType] = wp
		p.middlewares[taskType] = middlewares
		p.router.AddRoute(taskType, wp)

	case ModeActor:
		p.actorSystem.RegisterActor(taskType, chained)

	case ModeManager:
		if p.manager == nil {
			p.manager = NewManagerWorker(5, chained)
		}

	case ModeGossip:
		// Gossip uses connected pipelines; handlers live downstream
	}
}

// --- Task Submission ---
func (p *Pipeline) Submit(task *Task) error {
	p.metrics.TasksSubmitted.Inc()
	p.submitWg.Add()
	defer p.submitWg.Done()

	var err error
	switch p.mode {
	case ModeWorkerPool:
		workerPool, e := p.router.Route(task)
		if e != nil {
			return e
		}
		err = workerPool.Submit(task)

	case ModeActor:
		err = p.actorSystem.Submit(task)

	case ModeManager:
		if p.manager != nil {
			err = p.manager.Submit(task)
		}

	case ModeGossip:
		// Gossip: forward to connected pipelines
		p.forwardToConnections(task, false) // default: round-robin
	}

	return err
}

func (p *Pipeline) SubmitAndWait(task *Task) TaskResult {
	if err := p.Submit(task); err != nil {
		return TaskResult{TaskID: task.ID, Error: err, Success: false}
	}
	return task.WaitForResult()
}

// --- Forwarding (fan-out / round-robin) ---
func (p *Pipeline) forwardToConnections(task *Task, fanout bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.connections) == 0 {
		return
	}

	if fanout {
		// Send to all
		for _, conn := range p.connections {
			go conn.Submit(task)
		}
	} else {
		// Round-robin
		conn := p.connections[p.rrIndex%len(p.connections)]
		p.rrIndex++
		go conn.Submit(task)
	}
}

// --- Metrics ---
func (p *Pipeline) GetMetrics() *Metrics {
	return p.metrics
}

// --- Stop ---
func (p *Pipeline) Stop() {
	p.cancel()
	switch p.mode {
	case ModeWorkerPool:
		p.mu.RLock()
		for _, wp := range p.workerPools {
			wp.Stop()
		}
		p.mu.RUnlock()
	case ModeActor:
		p.actorSystem.Stop()
	case ModeManager:
		if p.manager != nil {
			p.manager.Stop()
		}
	case ModeGossip:
		for _, conn := range p.connections {
			conn.Stop()
		}
	}
}

// --- Middleware ---
func (p *Pipeline) AddMiddleware(middleware Middleware) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.mode != ModeWorkerPool {
		return // global middlewares only apply in WorkerPool mode
	}

	for taskType, handler := range p.handlers {
		p.middlewares[taskType] = append([]Middleware{middleware}, p.middlewares[taskType]...)
		chained := Chain(p.middlewares[taskType]...)(handler)
		if wp, ok := p.workerPools[taskType]; ok {
			wp.UpdateHandler(chained)
		}
	}
}

// --- Connections (for Gossip / DAG / Workflow) ---

// Connect: simple connection (round-robin or fanout based on forwardToConnections)
func (p *Pipeline) Connect(other *Pipeline) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connections = append(p.connections, other)
}

// ConnectTo: workflow chaining (fromTask → toTask in other pipeline)
func (p *Pipeline) ConnectTo(other *Pipeline, fromTask, toTask string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if handler, ok := p.handlers[fromTask]; ok {
		wrapped := func(ctx context.Context, task *Task) error {
			// run original handler
			err := handler(ctx, task)
			if err != nil {
				return err
			}

			// auto-forward
			next := NewTask(toTask, task.Payload)
			return other.Submit(next)
		}

		middlewares := p.middlewares[fromTask]
		chained := Chain(middlewares...)(wrapped)
		if wp, exists := p.workerPools[fromTask]; exists {
			wp.UpdateHandler(chained)
		}
	}
	p.connections = append(p.connections, other)
}

// ConnectIf: conditional connection (forward only if condition is true)
func (p *Pipeline) ConnectIf(other *Pipeline, fromTask, toTask string, cond func(*Task) bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if handler, ok := p.handlers[fromTask]; ok {
		wrapped := func(ctx context.Context, task *Task) error {
			err := handler(ctx, task)
			if err != nil {
				return err
			}
			if cond(task) {
				next := NewTask(toTask, task.Payload)
				return other.Submit(next)
			}
			return nil
		}
		middlewares := p.middlewares[fromTask]
		chained := Chain(middlewares...)(wrapped)
		if wp, exists := p.workerPools[fromTask]; exists {
			wp.UpdateHandler(chained)
		}
	}
	p.connections = append(p.connections, other)
}

// ConnectTransform: transform payload/metadata before forwarding
func (p *Pipeline) ConnectTransform(other *Pipeline, fromTask, toTask string, f func(*Task) *Task) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if handler, ok := p.handlers[fromTask]; ok {
		wrapped := func(ctx context.Context, task *Task) error {
			err := handler(ctx, task)
			if err != nil {
				return err
			}
			next := f(task)
			next.Type = toTask
			return other.Submit(next)
		}
		middlewares := p.middlewares[fromTask]
		chained := Chain(middlewares...)(wrapped)
		if wp, exists := p.workerPools[fromTask]; exists {
			wp.UpdateHandler(chained)
		}
	}
	p.connections = append(p.connections, other)
}

// MultiConnect: fan-out to multiple downstream pipelines
func (p *Pipeline) MultiConnect(fromTask string, links []struct {
	Other  *Pipeline
	ToTask string
}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if handler, ok := p.handlers[fromTask]; ok {
		wrapped := func(ctx context.Context, task *Task) error {
			err := handler(ctx, task)
			if err != nil {
				return err
			}
			// fan-out to all
			for _, link := range links {
				next := NewTask(link.ToTask, task.Payload)
				_ = link.Other.Submit(next)
			}
			return nil
		}
		middlewares := p.middlewares[fromTask]
		chained := Chain(middlewares...)(wrapped)
		if wp, exists := p.workerPools[fromTask]; exists {
			wp.UpdateHandler(chained)
		}
		for _, link := range links {
			p.connections = append(p.connections, link.Other)
		}
	}
}



// --- Global Pipeline (optional, for demos) ---
var globalPipeline *Pipeline

func SetGlobalPipeline(p *Pipeline) { globalPipeline = p }
func GlobalPipeline() *Pipeline     { return globalPipeline }
