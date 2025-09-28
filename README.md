# GoPipe <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 100 100">
  <!-- Pipe shape -->
  <rect x="10" y="40" width="80" height="20" rx="5" ry="5" fill="#00ADD8"/>
  
  <!-- Data flow arrows -->
  <polygon points="15,30 25,40 15,50" fill="white"/>
  <polygon points="45,30 55,40 45,50" fill="white"/>
  <polygon points="75,30 85,40 75,50" fill="white"/>
  
  <!-- Text -->
  <text x="50" y="85" font-size="14" text-anchor="middle" fill="#333" font-family="Arial, sans-serif">
    GoPipe
  </text>
</svg>


**GoPipe** is an experimental **task orchestration & communication library in Go**.  
It started with a simple worker pool, but grew into a playground for exploring:

- 🛠 **Worker pools** for scheduling and executing tasks.  
- 🎭 **Actor-like models** for stateful concurrency.  
- 🌍 **Gossip protocols** for sharing state across pipelines.  
- 👩‍✈️ **Manager-worker orchestration** for centralized control.  

The motivation is simple:  
👉 **Can we make Go’s goroutines & channels feel like a true pipeline system** that adapts to different workloads — from single-node task queues to distributed actor clusters?

---

## 🚀 Quick Start

```bash
go get github.com/mani-s-tiwari/gopipe
```

```go
import "github.com/mani-s-tiwari/gopipe/pkg/gopipe"

func main() {
    p := gopipe.NewPipeline()
    p.RegisterHandler("hello", func(ctx context.Context, t *gopipe.Task) error {
        fmt.Println("Hello from GoPipe!")
        return nil
    })

    task := gopipe.NewTask("hello", nil)
    p.Submit(task)
    p.Stop()
}

```

## Contributing

This project is still evolving.
- Ideas, discussions, and PRs are very welcome 💡✨
- Found a bug? Open an issue.
- Have an idea? Let’s discuss it.
- Want to experiment with a new mode (actors, gossip, DAG workflows)? Jump in!

## 📌 Vision

GoPipe is not "done" — it’s a research + learning project that could become a practical tool.
The goal is to explore:
- How far goroutines/channels can go as building blocks.
- What happens when we mix worker pools, actors, and gossip together.
- Whether Go can have a simple yet powerful task system without heavy dependencies.

### 🚀 Try it. Break it. Improve it.
Let’s see what pipelines in Go can really do.
