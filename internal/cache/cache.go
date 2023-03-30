package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type Config struct {
	Server string `mapstructure:"server"`
	Port   int    `mapstructure:"port"`
}

type Cache struct {
	cache *cache.Cache
	ctx   context.Context
}

func New(config *Config) *Cache {
	ring := redis.NewRing(&redis.RingOptions{
		Addrs: map[string]string{
			config.Server: fmt.Sprintf(":%d", config.Port),
		},
	})

	cache := cache.New(&cache.Options{
		Redis:      ring,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	return &Cache{
		cache: cache,
		ctx:   context.TODO(),
	}
}

// Set stores a value in the Redis cache.
func (c *Cache) Set(key string, v interface{}) error {
	return c.cache.Set(&cache.Item{
		Ctx:   c.ctx,
		Key:   key,
		Value: v,
		TTL:   time.Hour * 24 * 30,
	})
}

// Get reads a value from the Redis cache and saves it into the provided interface.
func (c *Cache) Get(key string, v interface{}) error {
	return c.cache.Get(c.ctx, key, &v)
}

// Delete removes a value from the Redis cache.
func (c *Cache) Delete(key string) error {
	return c.cache.Delete(c.ctx, key)
}
