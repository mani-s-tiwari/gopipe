package broker

import "context"

// Broker is the minimal contract for pluggable queue backends.
type Broker interface {
	// Enqueue pushes a payload (serialized task) to a topic/queue.
	Enqueue(ctx context.Context, topic string, payload []byte) error

	// Consume starts consuming messages from topic. Handler is called for each message.
	// Implementations should return when ctx is canceled.
	Consume(ctx context.Context, topic string, handler func(ctx context.Context, payload []byte) error) error

	// Close frees resources.
	Close() error
}
