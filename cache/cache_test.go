package cache

import (
	"testing"

	"github.com/go-redis/redis/v8"
)

func TestSetGetCacheWithRedis(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	cache_key := "cache_key"
	expectedName := "fulan"
	expectedAge := 23
	var err error

	Configure(Config{RedisOptions: &redis.Options{}})

	err = Set(cache_key, person{Name: expectedName, Age: expectedAge})
	if err != nil {
		t.Errorf("Error occurred on SetCache [%v]", err)
	}

	p := person{}
	err = Get(cache_key, &p)
	if err != nil {
		t.Errorf("Error occurred on GetCache [%v]", err)
	}
	if p.Name != expectedName {
		t.Errorf("Expected Name [%v], got [%v]", expectedName, p.Name)
	}
	if p.Age != expectedAge {
		t.Errorf("Expected Age [%v], got [%v]", expectedAge, p.Age)
	}
}

func TestSetGetCacheWithoutRedis(t *testing.T) {
	type person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	cache_key := "cache_key"
	expectedName := "fulan"
	expectedAge := 23
	var err error

	err = Set(cache_key, person{Name: expectedName, Age: expectedAge})
	if err != nil {
		t.Errorf("Error occurred on SetCache [%v]", err)
	}

	p := person{}
	err = Get(cache_key, &p)
	if err != nil {
		t.Errorf("Error occurred on GetCache [%v]", err)
	}
	if p.Name != expectedName {
		t.Errorf("Expected Name [%v], got [%v]", expectedName, p.Name)
	}
	if p.Age != expectedAge {
		t.Errorf("Expected Age [%v], got [%v]", expectedAge, p.Age)
	}
}
