package cache

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheClienter interface {
	Close() error
	HashExpireAt(context.Context, string, time.Time) error
	HashGet(context.Context, string) (string, error)
	HashSet(context.Context, string, string) error
	HashMultiGet(context.Context, ...string) (map[string]string, error)
	MakeField(int32) string
}

type TicketHoldClient struct {
	conn    *redis.Client
	hashKey string
}

func NewTicketHoldClient(conn *redis.Client, hashKey string) *TicketHoldClient {
	return &TicketHoldClient{conn: conn, hashKey: hashKey}
}

func NewTicketHoldClientFromURL(url, hashKey string) (*TicketHoldClient, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return &TicketHoldClient{conn: redis.NewClient(opts), hashKey: hashKey}, nil
}

// Close closes the underlying Redis connection.
func (repo *TicketHoldClient) Close() error {
	return repo.conn.Close()
}

// MakeField creates a Redis map field, i.e. a string, from a ticket's id.
func (repo *TicketHoldClient) MakeField(id int32) string {
	return strconv.FormatInt(int64(id), 10)
}

func (repo *TicketHoldClient) HashExpireAt(ctx context.Context, field string, expirationTime time.Time) error {
	return repo.conn.HExpireAt(ctx, repo.hashKey, expirationTime, field).Err()
}

func (repo *TicketHoldClient) HashGet(ctx context.Context, field string) (string, error) {
	result, err := repo.conn.HGet(ctx, repo.hashKey, field).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return result, ErrNotFound
		}
		return result, err
	}
	return result, nil
}

func (repo *TicketHoldClient) HashSet(ctx context.Context, field string, value string) error {
	fieldSet, err := repo.conn.HSetNX(ctx, repo.hashKey, field, value).Result()
	if err != nil {
		return err
	}
	if !fieldSet {
		// A response of 0 / false indicates that the field already exists.
		return ErrAlreadyHasHold
	}

	return nil
}

func (repo *TicketHoldClient) JoinHMGetResults(fields []string, values []interface{}) map[string]string {
	joined := make(map[string]string)
	for idx, value := range values {
		if value != nil {
			field := fields[idx]
			joined[field] = value.(string)
		}
	}
	return joined
}

func (repo *TicketHoldClient) HashMultiGet(ctx context.Context, fields ...string) (map[string]string, error) {
	result, err := repo.conn.HMGet(ctx, repo.hashKey, fields...).Result()
	if err != nil {
		return nil, err
	}
	return repo.JoinHMGetResults(fields, result), nil
}
