package cache

import (
	"math/rand"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type friend struct {
	FirstName string `json:"name.first,omitempty"`
	LastName  string `json:"name.last,omitempty"`
}

type person struct {
	ID        string    `json:"id,omitempty"`
	FirstName string    `json:"name.first,omitempty"`
	LastName  string    `json:"name.last,omitempty"`
	Age       int       `json:"age,omitempty"`
	Friends   []friend  `json:"friends"`
	CreatedAt time.Time `json:"created.time,omitempty"`
	UpdatedAt time.Time `json:"updated.time,omitempty"`
}

func newCacheTestData() (cacheKey string, expected person) {
	e := person{}
	e.ID = uuid.NewString()
	e.FirstName = "Phay"
	e.LastName = "Joe"
	e.Age = rand.Intn(100)
	e.Friends = append(e.Friends, friend{FirstName: "John", LastName: "Thor"})
	e.Friends = append(e.Friends, friend{FirstName: "Ryan", LastName: "Tho"})
	e.CreatedAt, _ = time.Parse(time.RFC3339, "2021-08-30T10:12:09Z")
	e.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	cacheKey = uuid.NewString()
	expected = e
	return
}

func checkGetCacheTestData(t *testing.T, title string, expected, result person) {
	if result.ID != expected.ID {
		t.Errorf("%v : Expected ID [%v], got [%v]", title, expected.ID, result.ID)
	}
	if result.FirstName != expected.FirstName {
		t.Errorf("%v : Expected FirstName [%v], got [%v]", title, expected.FirstName, result.FirstName)
	}
	if result.LastName != expected.LastName {
		t.Errorf("%v : Expected LastName [%v], got [%v]", title, expected.LastName, result.LastName)
	}
	if result.Age != expected.Age {
		t.Errorf("%v : Expected Age [%v], got [%v]", title, expected.Age, result.Age)
	}
	if len(result.Friends) != len(expected.Friends) {
		t.Errorf("%v : Expected Friends count [%v], got [%v]", title, len(expected.Friends), len(result.Friends))
	} else {
		if result.Friends[0].FirstName != expected.Friends[0].FirstName {
			t.Errorf("%v : Expected Friends[0].FirstName [%v], got [%v]", title, expected.Friends[0].FirstName, result.Friends[0].FirstName)
		}
		if result.Friends[0].LastName != expected.Friends[0].LastName {
			t.Errorf("%v : Expected Friends[0].LastName [%v], got [%v]", title, expected.Friends[0].LastName, result.Friends[0].LastName)
		}
		if result.Friends[1].FirstName != expected.Friends[1].FirstName {
			t.Errorf("%v : Expected Friends[1].FirstName [%v], got [%v]", title, expected.Friends[1].FirstName, result.Friends[1].FirstName)
		}
		if result.Friends[1].LastName != expected.Friends[1].LastName {
			t.Errorf("%v : Expected Friends[1].LastName [%v], got [%v]", title, expected.Friends[1].LastName, result.Friends[1].LastName)
		}
	}
	if result.CreatedAt != expected.CreatedAt {
		t.Errorf("%v : Expected CreatedAt [%v], got [%v]", title, expected.CreatedAt, result.CreatedAt)
	}
	if result.UpdatedAt != expected.UpdatedAt {
		t.Errorf("%v : Expected UpdatedAt [%v], got [%v]", title, expected.UpdatedAt, result.UpdatedAt)
	}
}

func TestSetGetCacheWithoutRedis(t *testing.T) {
	cacheKey, expected := newCacheTestData()
	err := Set(cacheKey, expected)
	if err != nil {
		t.Errorf("Test set cache without redis : Error occurred [%v]", err)
	}
	result := person{}
	err = Get(cacheKey, &result)
	if err != nil {
		t.Errorf("Test get cache without redis : Error occurred [%v]", err)
	}
	checkGetCacheTestData(t, "Test get cache without redis", expected, result)
}

func TestSetGetCacheWithRedis(t *testing.T) {
	Configure(Config{RedisOptions: &redis.Options{}})
	cacheKey, expected := newCacheTestData()
	err := Set(cacheKey, expected)
	if err != nil {
		t.Errorf("Test set cache with redis : Error occurred [%v]", err)
	}
	result := person{}
	err = Get(cacheKey, &result)
	if err != nil {
		t.Errorf("Test get cache with redis : Error occurred [%v]", err)
	}
	checkGetCacheTestData(t, "Test get cache with redis", expected, result)
}

func BenchmarkSetGetCacheWithoutRedis(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cacheKey, expected := newCacheTestData()
		err := Set(cacheKey, expected)
		if err == nil {
			result := person{}
			err = Get(cacheKey, &result)
		}
	}
}

func BenchmarkSetGetCacheWithRedis(b *testing.B) {
	Configure(Config{RedisOptions: &redis.Options{}})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cacheKey, expected := newCacheTestData()
		err := Set(cacheKey, expected)
		if err == nil {
			result := person{}
			err = Get(cacheKey, &result)
		}
	}
}
