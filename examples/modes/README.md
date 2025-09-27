## 🔎 Quick Reference (Execution Modes)

| Mode          | Description | When to Use | Example |
|---------------|-------------|-------------|---------|
| **WorkerPool** | Default GoPipe mode using goroutines + channels with priority queues. | General async task execution, background jobs, scheduled tasks. | `p := gopipe.NewPipelineWithMode(gopipe.ModeWorkerPool)` |
| **Actor**     | Actor-style concurrency. Each actor has its own mailbox, processes messages sequentially, and maintains state. | Stateful services, chat systems, gaming servers, IoT devices. | `p := gopipe.NewPipelineWithMode(gopipe.ModeActor)` |
| **Gossip**    | Distributed gossip protocol to share state & tasks between pipelines. | Clusters of services, load balancing, failure detection, distributed schedulers. | `p := gopipe.NewPipelineWithMode(gopipe.ModeGossip)` |
| **Manager**   | Master-worker mode. One manager distributes tasks to workers, central control. | Centralized scheduling, batch processing, monitoring-heavy workflows. | `p := gopipe.NewPipelineWithMode(gopipe.ModeManager)` |

---

## 🌟 Benefits of GoPipe

- ✅ **Production-ready task scheduler**: priorities, retries, backoff, and scheduling are built-in.  
- ⚡ **Leverages Go’s strengths**: goroutines + channels → minimal overhead, high concurrency.  
- 🔗 **Flexible architecture**: supports **WorkerPool, Actor, Gossip, and Manager modes**.  
- 🛠 **Pluggable middleware**: logging, retries, rate limiting, circuit breaking, and more.  
- 📊 **Metrics-first**: track submitted, completed, failed tasks, and latency out of the box.  
- 🌍 **Scalable**: pipelines can connect to each other or form clusters for distributed task flow.  
- 🚀 **Developer friendly**: simple API, but extensible for advanced workflows.  

GoPipe = **Celery-like power + Go-native simplicity** 🚀  
