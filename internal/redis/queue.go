package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis/v8"
)

const (
	JobQueueKey = "orbit:job_queue"
)

type Queue struct {
	client *redis.Client
}

func NewQueue(redisURL string) (*Queue, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Queue{client: client}, nil
}

func (q *Queue) EnqueueJob(ctx context.Context, jobID string) error {
	return q.client.RPush(ctx, JobQueueKey, jobID).Err()
}

func (q *Queue) DequeueJob(ctx context.Context) (string, error) {
	result, err := q.client.BLPop(ctx, 0, JobQueueKey).Result()
	if err != nil {
		return "", err
	}
	if len(result) < 2 {
		return "", fmt.Errorf("invalid redis response")
	}
	return result[1], nil
}

func (q *Queue) QueueLength(ctx context.Context) (int64, error) {
	return q.client.LLen(ctx, JobQueueKey).Result()
}

func (q *Queue) PublishJobUpdate(ctx context.Context, jobID string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return q.client.Publish(ctx, fmt.Sprintf("orbit:job:%s", jobID), payload).Err()
}

func (q *Queue) Close() error {
	return q.client.Close()
}
