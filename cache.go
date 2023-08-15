package grest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cache is a cache utility that manages caching using redis or in-memory data.
type Cache struct {
	Exp         time.Duration
	Ctx         context.Context
	RedisClient *redis.Client
	IsUseRedis  bool
	inMemCache  map[string]string
}

// Configure initializes the Cache by setting up the Redis client and context.
func (c *Cache) Configure() error {
	if c.RedisClient == nil {
		c.RedisClient = redis.NewClient(&redis.Options{})
	}
	if c.Ctx == nil {
		c.Ctx = context.Background()
	}
	err := c.RedisClient.Ping(c.Ctx).Err()
	if err != nil {
		c.IsUseRedis = false
		return NewError(http.StatusInternalServerError, err.Error())
	}
	c.IsUseRedis = true
	return nil
}

// Get retrieves a cached value associated with a key and stores the result in the value pointed to by val.
func (c *Cache) Get(key string, val any) error {
	if c.IsUseRedis {
		value, err := c.RedisClient.Get(c.Ctx, key).Result()
		if err != nil {
			return NewError(http.StatusInternalServerError, err.Error())
		}
		err = json.Unmarshal([]byte(value), val)
		if err != nil {
			return NewError(http.StatusInternalServerError, err.Error())
		}
	} else {
		value, ok := c.inMemCache[key]
		if !ok {
			return NewError(http.StatusInternalServerError, "Cache with key "+key+" is not found")
		}
		err := json.Unmarshal([]byte(value), val)
		if err != nil {
			return NewError(http.StatusInternalServerError, err.Error())
		}
	}
	return nil
}

// Set stores a value in the cache associated with a key and an optional expiration time.
func (c *Cache) Set(key string, val any, e ...time.Duration) error {
	expiration := c.Exp
	if len(e) > 0 {
		expiration = e[0]
	}
	value, err := json.Marshal(val)
	if err != nil {
		return NewError(http.StatusInternalServerError, err.Error())
	}
	if c.IsUseRedis {
		err = c.RedisClient.Set(c.Ctx, key, string(value), expiration).Err()
		if err != nil {
			return NewError(http.StatusInternalServerError, err.Error())
		}
	} else {
		if c.inMemCache != nil {
			c.inMemCache[key] = string(value)
		} else {
			c.inMemCache = map[string]string{key: string(value)}
		}
	}
	return nil
}

// Delete removes a cached value associated with a key.
func (c *Cache) Delete(key string) error {
	if c.IsUseRedis {
		err := c.RedisClient.Del(c.Ctx, key).Err()
		if err != nil {
			return NewError(http.StatusInternalServerError, err.Error())
		}
	} else {
		_, ok := c.inMemCache[key]
		if ok {
			delete(c.inMemCache, key)
		}
	}
	return nil
}

// DeleteWithPrefix removes all cached values with keys matching the specified prefix.
func (c *Cache) DeleteWithPrefix(prefix string) error {
	if c.IsUseRedis {
		var cursor uint64
		for {
			var keys []string
			var err error
			keys, cursor, err = c.RedisClient.Scan(c.Ctx, cursor, prefix+"*", 0).Result()
			if err != nil {
				return NewError(http.StatusInternalServerError, err.Error())
			}
			for _, k := range keys {
				c.RedisClient.Del(c.Ctx, k)
			}
			if cursor == 0 {
				break
			}
		}
	} else {
		for k := range c.inMemCache {
			if strings.HasPrefix(k, prefix+"") {
				delete(c.inMemCache, k)
			}
		}
	}
	return nil
}

// Invalidate removes cached values with keys having the specified prefix and additional keys.
func (c *Cache) Invalidate(prefix string, keys ...string) {
	for _, k := range keys {
		c.Delete(prefix + "." + k)
	}
	go c.DeleteWithPrefix(prefix + "?")
}

// Clear removes all cached values from the cache.
func (c *Cache) Clear() error {
	if c.IsUseRedis {
		err := c.RedisClient.FlushDB(c.Ctx).Err()
		if err != nil {
			return NewError(http.StatusInternalServerError, err.Error())
		}
	} else {
		c.inMemCache = map[string]string{}
	}
	return nil
}
