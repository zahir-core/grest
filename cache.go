package grest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	Exp         time.Duration
	Ctx         context.Context
	RedisClient *redis.Client
	IsUseRedis  bool
	inMemCache  map[string]string
}

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

func (c *Cache) Invalidate(prefix string, keys ...string) {
	for _, k := range keys {
		c.Delete(prefix + "." + k)
	}
	go c.DeleteWithPrefix(prefix + "?")
}

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
