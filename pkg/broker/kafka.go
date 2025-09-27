package broker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mani-s-tiwari/gopipe/pkg/gopipe"
	"github.com/segmentio/kafka-go"
)

type KafkaBroker struct {
	brokers []string
	writer  *kafka.Writer
	readers []*kafka.Reader
}

func NewKafkaBroker(brokers []string) *KafkaBroker {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		BatchTimeout: 200 * time.Millisecond,
	})
	return &KafkaBroker{
		brokers: brokers,
		writer:  w,
	}
}

// --- Enqueue ---
func (kb *KafkaBroker) Enqueue(ctx context.Context, task *gopipe.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return kb.writer.WriteMessages(ctx, kafka.Message{
		Topic: task.Type,
		Value: data,
		Time:  time.Now(),
	})
}

// --- Consume ---
func (kb *KafkaBroker) Consume(ctx context.Context, topic string, handler func(ctx context.Context, task *gopipe.Task) error) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: kb.brokers,
		Topic:   topic,
		GroupID: "gopipe-group",
	})
	kb.readers = append(kb.readers, r)

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			return err
		}

		var task gopipe.Task
		if err := json.Unmarshal(m.Value, &task); err != nil {
			continue
		}
		gopipe.RehydrateTask(&task)

		if err := handler(ctx, &task); err != nil {
			continue
		}
	}
}

func (kb *KafkaBroker) Close() error {
	_ = kb.writer.Close()
	for _, r := range kb.readers {
		_ = r.Close()
	}
	return nil
}
