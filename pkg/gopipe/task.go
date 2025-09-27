package gopipe

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mani-s-tiwari/gopipe/internal/utils"
)

// TaskPriority defines task priority levels
type TaskPriority int

const (
	PriorityLow TaskPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// Task represents an enhanced task unit
type Task struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Payload      json.RawMessage        `json:"payload"`
	Metadata     map[string]interface{} `json:"metadata"`
	Priority     TaskPriority           `json:"priority"`
	Retries      int                    `json:"retries"`
	MaxRetries   int                    `json:"max_retries"`
	Timeout      time.Duration          `json:"timeout"`
	CreatedAt    time.Time              `json:"created_at"`
	ScheduledFor time.Time              `json:"scheduled_for"`

	// Internal fields
	ctx        context.Context
	cancel     context.CancelFunc
	resultChan chan TaskResult
	attempts   int
}

// TaskResult represents task execution result
type TaskResult struct {
	TaskID   string        `json:"task_id"`
	Output   interface{}   `json:"output"`
	Error    error         `json:"error"`
	Duration time.Duration `json:"duration"`
	Attempts int           `json:"attempts"`
	Success  bool          `json:"success"`
}

// NewTask creates a new task with context and timeout
func NewTask(taskType string, payload []byte, opts ...TaskOption) *Task {
	task := &Task{
		ID:         utils.GenerateTaskID(),
		Type:       taskType, // used for routing
		Name:       taskType, // default name = type (can override with WithName)
		Payload:    payload,
		Metadata:   make(map[string]interface{}),
		Priority:   PriorityNormal,
		MaxRetries: 3,
		Timeout:    30 * time.Second,
		CreatedAt:  time.Now(),
		resultChan: make(chan TaskResult, 1),
	}

	for _, opt := range opts {
		opt(task)
	}

	task.ctx, task.cancel = context.WithTimeout(context.Background(), task.Timeout)
	return task
}

// TaskOption functional options for task configuration
type TaskOption func(*Task)

func RehydrateTask(t *Task) {
	if t.resultChan == nil {
		t.resultChan = make(chan TaskResult, 1)
	}
	if t.ctx == nil {
		t.ctx, t.cancel = context.WithTimeout(context.Background(), t.Timeout)
	}
}

func WithPriority(priority TaskPriority) TaskOption {
	return func(t *Task) {
		t.Priority = priority
	}
}

func WithName(name string) TaskOption {
	return func(t *Task) {
		t.Name = name
	}
}

func WithTimeout(timeout time.Duration) TaskOption {
	return func(t *Task) {
		t.Timeout = timeout
	}
}

func WithMetadata(key string, value interface{}) TaskOption {
	return func(t *Task) {
		if t.Metadata == nil {
			t.Metadata = make(map[string]interface{})
		}
		t.Metadata[key] = value
	}
}

func WithScheduledTime(t time.Time) TaskOption {
	return func(task *Task) {
		task.ScheduledFor = t
	}
}

// WaitForResult waits for task completion and returns result
func (t *Task) WaitForResult() TaskResult {
	select {
	case result := <-t.resultChan:
		return result
	case <-t.ctx.Done():
		return TaskResult{
			TaskID:  t.ID,
			Error:   t.ctx.Err(),
			Success: false,
		}
	}
}

// Complete signals task completion
func (t *Task) Complete(result TaskResult) {
	select {
	case t.resultChan <- result:
	default:
		// Already completed
	}
	t.cancel()
}
