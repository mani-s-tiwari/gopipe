package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mani-s-tiwari/gopipe/pkg/gopipe"
)

func main() {
	fmt.Println("=== WorkerPool Mode ===")
	runWorkerPoolMode()

	fmt.Println("\n=== Actor Mode ===")
	runActorMode()

	fmt.Println("\n=== Manager-Worker Mode ===")
	runManagerMode()

	fmt.Println("\n=== Gossip Mode ===")
	runGossipMode()
}

// ---------------- WorkerPool ----------------
func runWorkerPoolMode() {
	pipe := gopipe.NewPipelineWithMode(gopipe.ModeWorkerPool)
	pipe.RegisterHandler("echo", echoHandler, gopipe.LoggingMiddleware())

	for i := 1; i <= 3; i++ {
		task := gopipe.NewTask("echo", []byte(fmt.Sprintf(`{"msg":"worker %d"}`, i)))
		pipe.Submit(task)
	}

	time.Sleep(2 * time.Second)
	pipe.Stop()
}

// ---------------- Actor ----------------
func runActorMode() {
	pipe := gopipe.NewPipelineWithMode(gopipe.ModeActor)
	pipe.RegisterHandler("echo", echoHandler, gopipe.LoggingMiddleware())

	for i := 1; i <= 3; i++ {
		task := gopipe.NewTask("echo", []byte(fmt.Sprintf(`{"msg":"actor %d"}`, i)))
		pipe.Submit(task)
	}

	time.Sleep(2 * time.Second)
	pipe.Stop()
}

// ---------------- Manager-Worker ----------------
func runManagerMode() {
	pipe := gopipe.NewPipelineWithMode(gopipe.ModeManager)
	pipe.RegisterHandler("echo", echoHandler, gopipe.LoggingMiddleware())

	for i := 1; i <= 3; i++ {
		task := gopipe.NewTask("echo", []byte(fmt.Sprintf(`{"msg":"manager %d"}`, i)))
		pipe.Submit(task)
	}

	time.Sleep(2 * time.Second)
	pipe.Stop()
}

// ---------------- Gossip ----------------
func runGossipMode() {
	pipeA := gopipe.NewPipelineWithMode(gopipe.ModeWorkerPool)
	pipeB := gopipe.NewPipelineWithMode(gopipe.ModeWorkerPool)

	pipeA.RegisterHandler("echo", echoHandler, gopipe.LoggingMiddleware())
	pipeB.RegisterHandler("echo", echoHandler, gopipe.LoggingMiddleware())

	cluster := gopipe.NewGossipCluster(pipeA, pipeB)

	for i := 1; i <= 6; i++ {
		task := gopipe.NewTask("echo", []byte(fmt.Sprintf(`{"msg":"gossip %d"}`, i)))
		cluster.Submit(task)
	}

	time.Sleep(6 * time.Second)
	cluster.Stop()
}

// ---------------- Handler ----------------
func echoHandler(ctx context.Context, task *gopipe.Task) error {
	fmt.Printf("[handler] %s payload=%s\n", task.ID, string(task.Payload))
	time.Sleep(500 * time.Millisecond)
	return nil
}
