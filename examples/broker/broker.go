package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mani-s-tiwari/gopipe/pkg/broker"
	"github.com/mani-s-tiwari/gopipe/pkg/gopipe"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create broker (in-memory)
	mb := broker.NewMemoryBroker()

	// Create pipeline
	p := gopipe.NewPipeline()
	p.RegisterHandler("print", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("Worker handled:", string(task.Payload))
		return nil
	})

	// Start consumer that listens to broker "jobs" topic and pushes into pipeline
	go func() {
		_ = broker.StartConsumer(ctx, mb, "jobs", p)
	}()

	// Enqueue tasks
	for i := 1; i <= 5; i++ {
		task := gopipe.NewTask("print", []byte(fmt.Sprintf("job-%d", i)))
		_ = broker.EnqueueTask(ctx, mb, task)
	}

	// Let it run
	time.Sleep(2 * time.Second)
	p.Stop()
	_ = mb.Close()
}
