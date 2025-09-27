package gopipe

import (
	"sync/atomic"
	"time"
)

// Counter is a simple atomic counter
type Counter struct {
	val int64
}

func (c *Counter) Inc() {
	atomic.AddInt64(&c.val, 1)
}

func (c *Counter) Add(n int64) {
	atomic.AddInt64(&c.val, n)
}

func (c *Counter) Get() int64 {
	return atomic.LoadInt64(&c.val)
}

// Gauge is a point-in-time value
type Gauge struct {
	val int64
}

func (g *Gauge) Set(n int64) {
	atomic.StoreInt64(&g.val, n)
}

func (g *Gauge) Inc() {
	atomic.AddInt64(&g.val, 1)
}

func (g *Gauge) Dec() {
	atomic.AddInt64(&g.val, -1)
}

func (g *Gauge) Get() int64 {
	return atomic.LoadInt64(&g.val)
}

// Histogram tracks simple latency buckets
type Histogram struct {
	count   int64
	totalNs int64
}

func (h *Histogram) Observe(d time.Duration) {
	atomic.AddInt64(&h.count, 1)
	atomic.AddInt64(&h.totalNs, d.Nanoseconds())
}

func (h *Histogram) Avg() time.Duration {
	count := atomic.LoadInt64(&h.count)
	if count == 0 {
		return 0
	}
	total := atomic.LoadInt64(&h.totalNs)
	return time.Duration(total / count)
}

// Metrics holds all pipeline stats
type Metrics struct {
	TasksSubmitted Counter
	TasksCompleted Counter
	TasksFailed    Counter
	TasksInFlight  Gauge
	TaskLatency    Histogram
}

// NewMetrics creates metrics struct
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordTask updates metrics after a task finishes
func (m *Metrics) RecordTask(result TaskResult) {
	if result.Success {
		m.TasksCompleted.Inc()
	} else {
		m.TasksFailed.Inc()
	}
	m.TaskLatency.Observe(result.Duration)
}
