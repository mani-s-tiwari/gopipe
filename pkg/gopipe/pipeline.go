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
		// Gossip uses connected pipelines; handlers are registered downstream
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
		for _, conn := range p.connections {
			go conn.Submit(task) // gossip broadcast
		}
	}
	return err
}

func (p *Pipeline) SubmitAndWait(task *Task) TaskResult {
	if err := p.Submit(task); err != nil {
		return TaskResult{TaskID: task.ID, Error: err, Success: false}
	}
	return task.WaitForResult()
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

// --- Connections (for Gossip / DAG pipelines) ---
func (p *Pipeline) Connect(other *Pipeline) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.connections = append(p.connections, other)
}

// --- Global Pipeline (optional, for demos) ---
var globalPipeline *Pipeline

func SetGlobalPipeline(p *Pipeline) { globalPipeline = p }
func GlobalPipeline() *Pipeline     { return globalPipeline }
