package gopipe

// Router routes tasks to appropriate worker pools
type Router struct {
	routes   map[string]*WorkerPool
	fallback *WorkerPool
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]*WorkerPool),
	}
}

// AddRoute adds a route for a task type
func (r *Router) AddRoute(taskType string, workerPool *WorkerPool) {
	r.routes[taskType] = workerPool
}

// SetFallback sets fallback worker pool
func (r *Router) SetFallback(workerPool *WorkerPool) {
	r.fallback = workerPool
}

// Route routes a task to a worker pool
func (r *Router) Route(task *Task) (*WorkerPool, error) {
	if wp, exists := r.routes[task.Name]; exists {
		return wp, nil
	}

	if r.fallback != nil {
		return r.fallback, nil
	}

	return nil, ErrHandlerNotFound
}
