package redis_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lmtani/rinha-de-backend-2025/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	defaultQueueKey = "payment_queue"
	uuidPrefix      = "uuid:"
	queueTimeout    = 5 * time.Second
)

// RedisQueue implements the PaymentQueue port using Redis lists
type RedisQueue struct {
	client   *redis.Client
	queueKey string
	closed   bool
}

// NewRedisQueue creates a new Redis-backed payment queue
func NewRedisQueue(redisURL string, queueKey string) (*RedisQueue, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	if queueKey == "" {
		queueKey = defaultQueueKey
	}

	return &RedisQueue{
		client:   client,
		queueKey: queueKey,
		closed:   false,
	}, nil
}

// Send adds a payment to the Redis list queue
func (q *RedisQueue) Send(payment domain.Payment) error {
	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), queueTimeout)
	defer cancel()

	// Serialize the payment
	paymentData, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to serialize payment: %w", err)
	}

	// Add to the right of the list (RPUSH)
	if err := q.client.RPush(ctx, q.queueKey, paymentData).Err(); err != nil {
		return fmt.Errorf("failed to push payment to queue: %w", err)
	}

	return nil
}

// Receive returns a channel that delivers payments from the queue
func (q *RedisQueue) Receive() <-chan domain.Payment {
	paymentChan := make(chan domain.Payment)

	// Start a goroutine that polls Redis for new payments
	go func() {
		defer close(paymentChan)

		for !q.closed {
			ctx, cancel := context.WithTimeout(context.Background(), queueTimeout)

			// BLPOP with timeout to get the leftmost (oldest) element with blocking
			result, err := q.client.BLPop(ctx, queueTimeout, q.queueKey).Result()
			cancel()

			if err != nil {
				if err == redis.Nil {
					// No data available, continue polling
					time.Sleep(100 * time.Millisecond)
					continue
				}

				if !q.closed {
					// Only log the error if we're not intentionally closing
					fmt.Printf("Error polling Redis queue: %v\n", err)
				}
				continue
			}

			// BLPOP returns [key, value], we want the value (index 1)
			if len(result) < 2 {
				fmt.Println("Unexpected BLPOP result format")
				continue
			}

			var payment domain.Payment
			if err := json.Unmarshal([]byte(result[1]), &payment); err != nil {
				fmt.Printf("Failed to deserialize payment: %v\n", err)
				continue
			}

			paymentChan <- payment
		}
	}()

	return paymentChan
}

// Close stops the queue and closes resources
func (q *RedisQueue) Close() error {
	if q.closed {
		return nil
	}

	q.closed = true
	return q.client.Close()
}

// RedisStore implements the InMemoryStore port using Redis
type RedisStore struct {
	client *redis.Client
	ttl    time.Duration // TTL for UUID keys
}

// NewRedisStore creates a new Redis-backed UUID store
func NewRedisStore(redisURL string, ttl time.Duration) (*RedisStore, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check connection
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Default TTL to 24 hours if not provided
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	return &RedisStore{
		client: client,
		ttl:    ttl,
	}, nil
}

// Add adds a UUID to the store with TTL
func (s *RedisStore) Add(uuid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if UUID already exists
	if s.Exists(uuid) {
		return fmt.Errorf("UUID %s already exists", uuid)
	}

	// Store the UUID with value 1 and TTL
	key := uuidPrefix + uuid
	if err := s.client.Set(ctx, key, 1, s.ttl).Err(); err != nil {
		return fmt.Errorf("failed to store UUID: %w", err)
	}

	return nil
}

// Exists checks if a UUID exists in the store
func (s *RedisStore) Exists(uuid string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := uuidPrefix + uuid
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		fmt.Printf("Error checking UUID existence: %v\n", err)
		return false
	}

	return exists > 0
}

// Close closes the Redis connection
func (s *RedisStore) Close() error {
	return s.client.Close()
}
