package gopipe

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/mani-s-tiwari/gopipe/internal/utils"
)

// --- Scheduling Strategies ---
type SchedulingStrategy int

const (
	ScheduleFIFO SchedulingStrategy = iota
	ScheduleStrictPriority
	ScheduleWeightedRandom
)

// PriorityQueue implements heap.Interface for prioritized tasks
type PriorityQueue []*Task

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority // Higher priority first
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*Task)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// WorkerPool manages a pool of workers with priority queue
type WorkerPool struct {
	maxWorkers    int
	taskQueue     chan *Task
	priorityQueue PriorityQueue
	queueMutex    sync.Mutex
	workersWg     sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	handler       Handler
	metrics       *Metrics
	rateLimiter   *utils.RateLimiter

	scheduling SchedulingStrategy
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers int, handler Handler) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	wp := &WorkerPool{
		maxWorkers:    maxWorkers,
		taskQueue:     make(chan *Task, 1000),
		priorityQueue: make(PriorityQueue, 0),
		ctx:           ctx,
		cancel:        cancel,
		handler:       handler,
		metrics:       NewMetrics(),
		scheduling:    ScheduleStrictPriority, // default
	}

	heap.Init(&wp.priorityQueue)
	wp.start()
	return wp
}

// SetScheduling sets the scheduling strategy
func (wp *WorkerPool) SetScheduling(strategy SchedulingStrategy) {
	wp.scheduling = strategy
}

// Submit adds a task to the pool
func (wp *WorkerPool) Submit(task *Task) error {
	select {
	case wp.taskQueue <- task:
		wp.metrics.TasksSubmitted.Inc()
		return nil
	case <-wp.ctx.Done():
		return context.Canceled
	default:
		// Queue full, use priority queue
		wp.queueMutex.Lock()
		heap.Push(&wp.priorityQueue, task)
		wp.queueMutex.Unlock()
		return nil
	}
}

// WithRateLimit sets rate limiter
func (wp *WorkerPool) WithRateLimit(rate float64, capacity float64) {
	wp.rateLimiter = utils.NewRateLimiter(rate, capacity)
}

// start the worker pool
func (wp *WorkerPool) start() {
	// Start workers
	for i := 0; i < wp.maxWorkers; i++ {
		wp.workersWg.Add(1)
		go wp.worker(i)
	}

	// Start priority queue processor
	go wp.processPriorityQueue()
}

func (wp *WorkerPool) worker(id int) {
	defer wp.workersWg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case task := <-wp.taskQueue:
			wp.processTask(task)
		}
	}
}

func (wp *WorkerPool) processPriorityQueue() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case <-ticker.C:
			wp.queueMutex.Lock()
			for wp.priorityQueue.Len() > 0 && len(wp.taskQueue) < cap(wp.taskQueue) {
				var task *Task

				switch wp.scheduling {
				case ScheduleFIFO:
					// pop oldest (like queue)
					task = heap.Pop(&wp.priorityQueue).(*Task)

				case ScheduleStrictPriority:
					// always pop highest priority
					task = heap.Pop(&wp.priorityQueue).(*Task)

				case ScheduleWeightedRandom:
					// pick weighted random
					task = wp.popTaskWeighted()
				}

				if task != nil {
					wp.taskQueue <- task
				}
			}
			wp.queueMutex.Unlock()
		}
	}
}

// --- Weighted Scheduling Helper ---
func (wp *WorkerPool) popTaskWeighted() *Task {
	if wp.priorityQueue.Len() == 0 {
		return nil
	}

	weights := make([]float64, wp.priorityQueue.Len())
	items := make([]*Task, wp.priorityQueue.Len())
	for i, task := range wp.priorityQueue {
		items[i] = task
		weights[i] = float64(task.Priority + 1) // higher priority = bigger weight
	}

	// Pick random index
	task := utils.WeightedRandom(items, weights)

	// Remove picked task from queue
	for i, t := range wp.priorityQueue {
		if t == task {
			heap.Remove(&wp.priorityQueue, i)
			break
		}
	}
	return task
}

func (wp *WorkerPool) processTask(task *Task) {
	if wp.rateLimiter != nil {
		wp.rateLimiter.Wait() // block until token available
	}
	start := time.Now()

	result := TaskResult{
		TaskID:   task.ID,
		Attempts: task.attempts,
		Success:  false,
	}

	defer func() {
		result.Duration = time.Since(start)
		task.Complete(result)
		wp.metrics.RecordTask(result)
	}()

	// Check if task is scheduled for future
	if !task.ScheduledFor.IsZero() && time.Now().Before(task.ScheduledFor) {
		time.Sleep(time.Until(task.ScheduledFor))
	}

	err := wp.handler(task.ctx, task)
	if err != nil {
		result.Error = err
	} else {
		result.Success = true
	}
}

// Stop gracefully stops the worker pool
func (wp *WorkerPool) Stop() {
	wp.cancel()
	wp.workersWg.Wait()
}

// UpdateHandler updates the handler in place
func (wp *WorkerPool) UpdateHandler(h Handler) {
	wp.queueMutex.Lock()
	defer wp.queueMutex.Unlock()
	wp.handler = h
}
