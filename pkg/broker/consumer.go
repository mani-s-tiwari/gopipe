package broker

import (
	"context"
	"encoding/json"

	"github.com/mani-s-tiwari/gopipe/pkg/gopipe"
)

// StartConsumer wires broker -> pipeline for a topic.
// It unmarshals Task JSON and submits to pipeline.
func StartConsumer(ctx context.Context, b Broker, topic string, pipeline *gopipe.Pipeline) error {
	return b.Consume(ctx, topic, func(ctx context.Context, payload []byte) error {
		var t gopipe.Task
		if err := json.Unmarshal(payload, &t); err != nil {
			return err
		}

		// Restore internals (ctx, cancel, resultChan)
		gopipe.RehydrateTask(&t)

		// Submit pointer because pipeline expects *Task
		return pipeline.Submit(&t)
	})
}

// EnqueueTask serializes gopipe.Task and pushes it to broker
func EnqueueTask(ctx context.Context, b Broker, task *gopipe.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return b.Enqueue(ctx, task.Type, data)
}
