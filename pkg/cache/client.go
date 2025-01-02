package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheClienter interface {
	Close() error
	Get(context.Context, string) (string, error)
	Set(context.Context, string, string, time.Duration) error
	GetMany(context.Context, ...string) (map[string]string, error)
	MakeKey(int32) string
}

type TicketHoldClient struct {
	conn             *redis.Client
	ticketHoldPrefix string
}

func NewTicketHoldClient(conn *redis.Client, ticketHoldPrefix string) *TicketHoldClient {
	return &TicketHoldClient{conn: conn, ticketHoldPrefix: ticketHoldPrefix}
}

func NewTicketHoldClientFromURL(url string, ticketHoldPrefix string) (*TicketHoldClient, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return &TicketHoldClient{
		conn:             redis.NewClient(opts),
		ticketHoldPrefix: ticketHoldPrefix,
	}, nil
}

// Close closes the underlying Redis connection.
func (repo *TicketHoldClient) Close() error {
	return repo.conn.Close()
}

// MakeKey creates a Redis key, i.e. a string, from a ticket's id.
func (repo *TicketHoldClient) MakeKey(id int32) string {
	return fmt.Sprintf("%s%d", repo.ticketHoldPrefix, id)
}

func (repo *TicketHoldClient) ExpireAt(ctx context.Context, key string, expirationTime time.Time) error {
	return repo.conn.ExpireAt(ctx, key, expirationTime).Err()
}

func (repo *TicketHoldClient) Get(ctx context.Context, key string) (string, error) {
	result, err := repo.conn.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return result, ErrNotFound
		}
		return result, err
	}
	return result, nil
}

func (repo *TicketHoldClient) Set(
	ctx context.Context,
	key string,
	value string,
	expiration time.Duration,
) error {
	keySet, err := repo.conn.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return err
	}
	if !keySet {
		// A response of 0 / false indicates that the key already exists.
		return ErrAlreadyHasHold
	}

	return nil
}

func (repo *TicketHoldClient) JoinMGetResults(keys []string, values []interface{}) map[string]string {
	joined := make(map[string]string)
	for idx, value := range values {
		if value != nil {
			key := keys[idx]
			joined[key] = value.(string)
		}
	}
	return joined
}

func (repo *TicketHoldClient) GetMany(ctx context.Context, keys ...string) (map[string]string, error) {
	result, err := repo.conn.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}
	return repo.JoinMGetResults(keys, result), nil
}
