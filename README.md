<img width="500" height="500" alt="svgviewer-png-output" src="https://github.com/user-attachments/assets/438fa7da-3fcd-490a-8531-f1b6fe8d85a2" />
# GoPipe <img src="https://github.com/user-attachments/assets/72f56a0b-389a-46be-98f9-ebed4b7e4eb7" alt="GoPipe Logo" width="40" height="40" style="vertical-align: middle;"/>


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
