package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	isUseRedis bool
	rdb        *redis.Client
	exp        time.Duration
	cache      map[string]string
	bgCtx      = context.Background()
)

type Config struct {
	RedisOptions      *redis.Options
	DefaultExpiration time.Duration
}

func Configure(c Config) {
	exp = c.DefaultExpiration
	rdb = redis.NewClient(c.RedisOptions)
	err := rdb.Ping(bgCtx).Err()
	if err == nil {
		isUseRedis = true
	} else {
		fmt.Println("Failed to connect to redis. The cache will be use in-memory local storage")
	}
}

func Get(key string, val interface{}) error {
	if isUseRedis {
		value, err := rdb.Get(bgCtx, key).Result()
		if err != nil {
			return err
		}
		return json.Unmarshal([]byte(value), val)
	}
	value, ok := cache[key]
	if !ok {
		return errors.New("Cache with key " + key + " is not found")
	}
	return json.Unmarshal([]byte(value), val)
}

func Set(key string, val interface{}, e ...time.Duration) error {
	expiration := exp
	if len(e) > 0 {
		expiration = e[0]
	}
	value, err := json.Marshal(val)
	if err != nil {
		return err
	}
	if isUseRedis {
		return rdb.Set(bgCtx, key, string(value), expiration).Err()
	} else {
		if cache != nil {
			cache[key] = string(value)
		} else {
			cache = map[string]string{key: string(value)}
		}
	}
	return nil
}

func Delete(key string) error {
	if isUseRedis {
		err := rdb.Del(bgCtx, key).Err()
		if err != nil {
			return err
		}
	} else {
		_, ok := cache[key]
		if ok {
			delete(cache, key)
		}
	}
	return nil
}

func DeleteWithPrefix(prefix string) error {
	if isUseRedis {
		var cursor uint64
		for {
			var keys []string
			var err error
			keys, cursor, err = rdb.Scan(bgCtx, cursor, prefix+":*", 0).Result()
			if err != nil {
				return err
			}
			for _, k := range keys {
				rdb.Del(bgCtx, k)
			}
			if cursor == 0 {
				break
			}
		}
	} else {
		for k := range cache {
			if strings.HasPrefix(k, prefix+"") {
				delete(cache, k)
			}
		}
	}
	return nil
}

func Clear() error {
	if isUseRedis {
		return rdb.FlushDB(bgCtx).Err()
	} else {
		cache = map[string]string{}
	}
	return nil
}
