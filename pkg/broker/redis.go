package broker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mani-s-tiwari/gopipe/pkg/gopipe"
	"github.com/redis/go-redis/v9"
)

var ErrNil = redis.Nil

type RedisBroker struct {
	client           *redis.Client
	ns               string
	processingSuffix string
}

type RedisOptions struct {
	Addr     string
	Password string
	DB       int
	Prefix   string
}

func NewRedisBrokerWithOpts(opts RedisOptions) (*RedisBroker, error) {
	r := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})
	if err := r.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	prefix := opts.Prefix
	if prefix == "" {
		prefix = "gopipe"
	}
	return &RedisBroker{
		client:           r,
		ns:               prefix,
		processingSuffix: ":processing",
	}, nil
}

func NewRedisBroker(addr string) (*RedisBroker, error) {
	return NewRedisBrokerWithOpts(RedisOptions{Addr: addr})
}

func (rb *RedisBroker) key(topic string) string {
	return rb.ns + ":" + topic
}
func (rb *RedisBroker) processingKey(topic string) string {
	return rb.key(topic) + rb.processingSuffix
}

// --- Enqueue ---
func (rb *RedisBroker) Enqueue(ctx context.Context, task *gopipe.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return rb.client.RPush(ctx, rb.key(task.Type), data).Err()
}

// --- Consume ---
func (rb *RedisBroker) Consume(ctx context.Context, topic string, handler func(ctx context.Context, task *gopipe.Task) error) error {
	main := rb.key(topic)
	processing := rb.processingKey(topic)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Atomic pop main → processing
		data, err := rb.client.BRPopLPush(ctx, main, processing, 5*time.Second).Bytes()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			return err
		}

		var task gopipe.Task
		if err := json.Unmarshal(data, &task); err != nil {
			continue
		}
		gopipe.RehydrateTask(&task)

		if err := handler(ctx, &task); err != nil {
			_ = rb.client.RPush(ctx, main, data)          // retry
			_ = rb.client.LRem(ctx, processing, 1, data) // remove stuck
			continue
		}

		_ = rb.client.LRem(ctx, processing, 1, data)
	}
}

func (rb *RedisBroker) Close() error {
	return rb.client.Close()
}
