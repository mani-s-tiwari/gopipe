package main

import (
	"context"
	"fmt"
	"time"

	"github.com/mani-s-tiwari/gopipe/pkg/gopipe"
)

func main() {
	fmt.Println("===== Example 1: Connect =====")
	exampleConnect()

	fmt.Println("\n===== Example 2: ConnectTo =====")
	exampleConnectTo()

	fmt.Println("\n===== Example 3: ConnectIf =====")
	exampleConnectIf()

	fmt.Println("\n===== Example 4: ConnectTransform =====")
	exampleConnectTransform()

	fmt.Println("\n===== Example 5: MultiConnect =====")
	exampleMultiConnect()
}

// ---------------- Example 1 ----------------
// Connect just links pipelines, you forward manually inside handler
func exampleConnect() {
	p1 := gopipe.NewPipeline()
	p2 := gopipe.NewPipeline()
	p1.Connect(p2)

	p1.RegisterHandler("step1", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P1] Step1 handling:", string(task.Payload))
		// manually forward to p2
		next := gopipe.NewTask("step2", []byte("from step1"))
		return p2.Submit(next)
	})

	p2.RegisterHandler("step2", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P2] Step2 received:", string(task.Payload))
		return nil
	})

	p1.Submit(gopipe.NewTask("step1", []byte("hello connect")))
	time.Sleep(1 * time.Second)
}

// ---------------- Example 2 ----------------
// ConnectTo automatically forwards after step1 → step2
func exampleConnectTo() {
	p1 := gopipe.NewPipeline()
	p2 := gopipe.NewPipeline()

	p1.RegisterHandler("resize", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P1] Resizing:", string(task.Payload))
		return nil
	})
	p2.RegisterHandler("upload", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P2] Uploading:", string(task.Payload))
		return nil
	})

	p1.ConnectTo(p2, "resize", "upload")

	p1.Submit(gopipe.NewTask("resize", []byte("photo.jpg")))
	time.Sleep(1 * time.Second)
}

// ---------------- Example 3 ----------------
// ConnectIf forwards only when condition is true
func exampleConnectIf() {
	p1 := gopipe.NewPipeline()
	p2 := gopipe.NewPipeline()

	p1.RegisterHandler("validate", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P1] Validating:", string(task.Payload))
		return nil
	})
	p2.RegisterHandler("process", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P2] Processing:", string(task.Payload))
		return nil
	})

	// only forward if payload contains "ok"
	p1.ConnectIf(p2, "validate", "process", func(t *gopipe.Task) bool {
		return string(t.Payload) == "ok"
	})

	p1.Submit(gopipe.NewTask("validate", []byte("ok")))    // ✅ forwarded
	p1.Submit(gopipe.NewTask("validate", []byte("reject"))) // 🚫 not forwarded
	time.Sleep(1 * time.Second)
}

// ---------------- Example 4 ----------------
// ConnectTransform changes payload/metadata before forwarding
func exampleConnectTransform() {
	p1 := gopipe.NewPipeline()
	p2 := gopipe.NewPipeline()

	p1.RegisterHandler("parse", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P1] Parsing:", string(task.Payload))
		return nil
	})
	p2.RegisterHandler("store", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P2] Storing:", string(task.Payload))
		return nil
	})

	p1.ConnectTransform(p2, "parse", "store", func(t *gopipe.Task) *gopipe.Task {
		return gopipe.NewTask("store", []byte("transformed:"+string(t.Payload)))
	})

	p1.Submit(gopipe.NewTask("parse", []byte("raw data")))
	time.Sleep(1 * time.Second)
}

// ---------------- Example 5 ----------------
// MultiConnect fans out to multiple pipelines
func exampleMultiConnect() {
	p1 := gopipe.NewPipeline()
	p2a := gopipe.NewPipeline()
	p2b := gopipe.NewPipeline()

	p1.RegisterHandler("event", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P1] Event:", string(task.Payload))
		return nil
	})

	p2a.RegisterHandler("log", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P2A] Logging:", string(task.Payload))
		return nil
	})
	p2b.RegisterHandler("alert", func(ctx context.Context, task *gopipe.Task) error {
		fmt.Println("[P2B] Alerting:", string(task.Payload))
		return nil
	})

	// fan-out to both pipelines
	p1.MultiConnect("event", []struct {
		Other  *gopipe.Pipeline
		ToTask string
	}{
		{Other: p2a, ToTask: "log"},
		{Other: p2b, ToTask: "alert"},
	})

	p1.Submit(gopipe.NewTask("event", []byte("system down")))
	time.Sleep(1 * time.Second)
}
