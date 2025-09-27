## 🔎 Quick Reference (Concept → GoPipe)

| Concept / Feature          | GoPipe API                              | Example (Go) |
|----------------------------|------------------------------------------|--------------|
| Create pipeline            | `gopipe.NewPipeline()`                  | `p := gopipe.NewPipeline()` |
| Register a handler         | `RegisterHandler(name, handler, ...)`   | `p.RegisterHandler("send_email", sendEmailHandler)` |
| Submit async task          | `pipeline.Submit(task)`                 | `p.Submit(task)` |
| Submit & wait for result   | `pipeline.SubmitAndWait(task)`          | `res := p.SubmitAndWait(task)` |
| Create a task              | `gopipe.NewTask(name, payload, opts…)`  | `t := gopipe.NewTask("job", []byte("data"))` |
| Priority tasks             | `WithPriority(PriorityHigh)`            | `t := gopipe.NewTask("job", nil, gopipe.WithPriority(gopipe.PriorityHigh))` |
| Scheduled tasks            | `WithScheduledTime(time)`               | `t := gopipe.NewTask("job", nil, gopipe.WithScheduledTime(time.Now().Add(5*time.Minute)))` |
| Add metadata               | `WithMetadata(key, val)`                | `t := gopipe.NewTask("job", nil, gopipe.WithMetadata("user_id", 123))` |
| Logging middleware         | `gopipe.LoggingMiddleware()`            | `p.RegisterHandler("x", h, gopipe.LoggingMiddleware())` |
| Retry middleware           | `gopipe.RetryMiddleware(DefaultBackoff())` | `p.RegisterHandler("x", h, gopipe.RetryMiddleware(gopipe.DefaultBackoff()))` |
| Timeout middleware         | `gopipe.TimeoutMiddleware(d)`           | `p.RegisterHandler("x", h, gopipe.TimeoutMiddleware(10*time.Second))` |
| Circuit breaker middleware | `gopipe.CircuitBreakerMiddleware(cb)`   | `cb := gopipe.NewCircuitBreaker(5, 30*time.Second)` |
| Get metrics                | `pipeline.GetMetrics()`                 | `m := p.GetMetrics(); fmt.Println(m.TasksCompleted.Get())` |
| Graceful shutdown          | `pipeline.Stop()`                       | `p.Stop()` |

---

## 🌟 Benefits of GoPipe

- ✅ **Production-ready task scheduler**: priorities, retries, backoff, and scheduling are built-in.  
- ⚡ **Leverages Go’s strengths**: goroutines + channels → minimal overhead, high concurrency.  
- 🔗 **Flexible architecture**: supports WorkerPool, Actor-style, Gossip, and Manager modes.  
- 🛠 **Pluggable middleware**: logging, retries, rate limiting, circuit breaking, and more.  
- 📊 **Metrics-first**: track submitted, completed, failed tasks, and latency out of the box.  
- 🌍 **Scalable**: pipelines can connect to each other or form clusters for distributed task flow.  
- 🚀 **Developer friendly**: simple API, but extensible for advanced workflows.  

GoPipe = **Celery-like power + Go-native simplicity**.  
