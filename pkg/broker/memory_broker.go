package broker

import (
	"context"
)

type MemoryBroker struct {
	queues map[string]chan []byte
}

func NewMemoryBroker() *MemoryBroker {
	return &MemoryBroker{queues: make(map[string]chan []byte)}
}

func (mb *MemoryBroker) getQueue(topic string) chan []byte {
	if q, ok := mb.queues[topic]; ok {
		return q
	}
	q := make(chan []byte, 100)
	mb.queues[topic] = q
	return q
}

// --- Enqueue ---
// Matches the Broker interface: takes raw []byte payload
func (mb *MemoryBroker) Enqueue(ctx context.Context, topic string, payload []byte) error {
	q := mb.getQueue(topic)
	select {
	case q <- payload:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// --- Consume ---
// Calls the provided handler with raw []byte payload
func (mb *MemoryBroker) Consume(ctx context.Context, topic string, handler func(ctx context.Context, payload []byte) error) error {
	q := mb.getQueue(topic)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data := <-q:
			if err := handler(ctx, data); err != nil {
				continue // skip errors, don't kill loop
			}
		}
	}
}

// --- Close ---
func (mb *MemoryBroker) Close() error {
	return nil
}
