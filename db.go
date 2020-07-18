package beepboop

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
)

// DB ...
type DB struct {
	client        *redis.Client
	CacheDuration time.Duration
}

type dbContextKeyType struct{}

var dbContextKey = &dbContextKeyType{}

// NewDB returns a new DB
func NewDB(addr, password string, db int) (*DB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	err := client.Ping().Err()
	if err != nil {
		client.Close()
		return nil, err
	}

	return &DB{
		client:        client,
		CacheDuration: time.Hour,
	}, nil
}

// DBFromContext returns the DB from the given Context (if exists)
func DBFromContext(ctx context.Context) *DB {
	db := ctx.Value(dbContextKey)
	if db != nil {
		return db.(*DB)
	}
	return nil
}

// ToContext adds the DB to the given Context
func (db *DB) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, dbContextKey, db)
}

// CacheValue caches a value
func (db *DB) CacheValue(key string, value interface{}, rewriteExisting bool) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if rewriteExisting {
		return db.client.Set("beepboop-cache:"+key, data, db.CacheDuration).Err()
	}
	return db.client.SetNX("beepboop-cache:"+key, data, db.CacheDuration).Err()
}

// UncacheValue removes a cached value
func (db *DB) UncacheValue(key string) error {
	return db.client.Del("beepboop-cache:" + key).Err()
}

// GetCachedValue tries to unmarshal a cached value
func (db *DB) GetCachedValue(key string, value interface{}) error {
	data, err := db.client.Get("beepboop-cache:" + key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), value)
}

// IsWithinRateLimit returns whether a request is withing rate limit per minute
func (db *DB) IsWithinRateLimit(reqType, ip string, rate int) (bool, error) {
	key := fmt.Sprintf("beepboop-rate:%s:%s", reqType, ip)
	pipe := db.client.TxPipeline()
	incr := pipe.Incr(key)
	pipe.Expire(key, time.Minute)
	_, err := pipe.Exec()
	if err != nil {
		return false, err
	}

	return int(incr.Val()) <= rate, nil
}
