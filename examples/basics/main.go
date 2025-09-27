package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mani-s-tiwari/gopipe/pkg/gopipe"
)

func main() {
	// Create two pipelines
	p1 := gopipe.NewPipeline()
	p2 := gopipe.NewPipeline()

	// Register handler in p1
	p1.RegisterHandler("step1", step1Handler, gopipe.LoggingMiddleware())

	// Register handler in p2
	p2.RegisterHandler("step2", step2Handler, gopipe.LoggingMiddleware())

	// Connect step1 → step2 automatically
	p1.ConnectTo(p2, "step1", "step2")

	// Submit a task into p1
	task := gopipe.NewTask("step1", []byte("hello world"))
	p1.Submit(task)

	// Let it run
	time.Sleep(3 * time.Second)

	// Shutdown
	p1.Stop()
	p2.Stop()
}

func step1Handler(ctx context.Context, task *gopipe.Task) error {
	fmt.Println("➡️ Step1 processed:", string(task.Payload))
	return nil // no need to forward manually anymore
}

func step2Handler(ctx context.Context, task *gopipe.Task) error {
	fmt.Println("✅ Step2 received:", string(task.Payload))
	return nil
}
