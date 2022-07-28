package grest

import (
	"math/rand"
	"testing"
	"time"

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
	c := &Cache{}
	err := c.Set(cacheKey, expected)
	if err != nil {
		t.Errorf("Test set cache without redis : Error occurred [%v]", err)
	}
	result := person{}
	err = c.Get(cacheKey, &result)
	if err != nil {
		t.Errorf("Test get cache without redis : Error occurred [%v]", err)
	}
	checkGetCacheTestData(t, "Test get cache without redis", expected, result)
}

func TestSetGetCacheWithRedis(t *testing.T) {
	c := &Cache{}
	c.Configure()
	cacheKey, expected := newCacheTestData()
	err := c.Set(cacheKey, expected)
	if err != nil {
		t.Errorf("Test set cache with redis : Error occurred [%v]", err)
	}
	result := person{}
	err = c.Get(cacheKey, &result)
	if err != nil {
		t.Errorf("Test get cache with redis : Error occurred [%v]", err)
	}
	checkGetCacheTestData(t, "Test get cache with redis", expected, result)
}

func TestDeleteWithKeyCacheWithRedis(t *testing.T) {
	c := &Cache{}
	c.Configure()
	cacheKey1 := uuid.NewString()
	cacheData1 := map[string]bool{"ok": true}
	err := c.Set(cacheKey1, cacheData1)
	if err != nil {
		t.Errorf("Test set cache with redis : Error occurred [%v]", err)
	}
	result1 := map[string]bool{}
	err = c.Get(cacheKey1, &result1)
	if err != nil {
		t.Errorf("Test get cache with redis : Error occurred [%v]", err)
	}
	if res1_ok, ok := result1["ok"]; !ok || !res1_ok {
		t.Errorf("Test get cache with redis : Expected ok [true], got [%v]", ok)
		t.Errorf("Test get cache with redis : Expected res1_ok [true], got [%v]", res1_ok)
	}
	err = c.DeleteWithPrefix(cacheKey1)
	if err != nil {
		t.Errorf("Test delete with prefix cache with redis : Error occurred [%v]", err)
	}
	result1_deleted := map[string]bool{}
	c.Get(cacheKey1, &result1_deleted)
	if res1_ok, ok := result1_deleted["ok"]; ok || res1_ok {
		t.Errorf("Test delete with prefix cache with redis : Expected ok [false], got [%v]", ok)
		t.Errorf("Test delete with prefix cache with redis : Expected res1_ok [false], got [%v]", res1_ok)
	}

	cacheKey2 := "foo.bar.baz?foo=bar"
	cacheData2 := map[string]bool{"ok": true}
	err = c.Set(cacheKey2, cacheData2)
	if err != nil {
		t.Errorf("Test set cache with redis : Error occurred [%v]", err)
	}
	result2 := map[string]bool{}
	err = c.Get(cacheKey2, &result2)
	if err != nil {
		t.Errorf("Test get cache with redis : Error occurred [%v]", err)
	}
	if res2_ok, ok := result2["ok"]; !ok || !res2_ok {
		t.Errorf("Test get cache with redis : Expected ok [true], got [%v]", ok)
		t.Errorf("Test get cache with redis : Expected res2_ok [true], got [%v]", res2_ok)
	}
	err = c.DeleteWithPrefix(cacheKey2)
	if err != nil {
		t.Errorf("Test delete with prefix cache with redis : Error occurred [%v]", err)
	}
	result2_deleted := map[string]bool{}
	c.Get(cacheKey2, &result2_deleted)
	if res2_ok, ok := result2_deleted["ok"]; ok || res2_ok {
		t.Errorf("Test delete with prefix cache with redis : Expected ok [false], got [%v]", ok)
		t.Errorf("Test delete with prefix cache with redis : Expected res2_ok [false], got [%v]", res2_ok)
	}
}

func TestDeleteWithPrefixCacheWithRedis(t *testing.T) {
	c := &Cache{}
	c.Configure()
	prefix1 := "foo.bar.baz."
	cacheKey1 := prefix1 + uuid.NewString()
	cacheData1 := map[string]bool{"ok": true}
	err := c.Set(cacheKey1, cacheData1)
	if err != nil {
		t.Errorf("Test set cache with redis : Error occurred [%v]", err)
	}
	result1 := map[string]bool{}
	err = c.Get(cacheKey1, &result1)
	if err != nil {
		t.Errorf("Test get cache with redis : Error occurred [%v]", err)
	}
	if res1_ok, ok := result1["ok"]; !ok || !res1_ok {
		t.Errorf("Test get cache with redis : Expected ok [true], got [%v]", ok)
		t.Errorf("Test get cache with redis : Expected res1_ok [true], got [%v]", res1_ok)
	}
	err = c.DeleteWithPrefix(prefix1)
	if err != nil {
		t.Errorf("Test delete with prefix cache with redis : Error occurred [%v]", err)
	}
	result1_deleted := map[string]bool{}
	c.Get(cacheKey1, &result1_deleted)
	if res1_ok, ok := result1_deleted["ok"]; ok || res1_ok {
		t.Errorf("Test delete with prefix cache with redis : Expected ok [false], got [%v]", ok)
		t.Errorf("Test delete with prefix cache with redis : Expected res1_ok [false], got [%v]", res1_ok)
	}

	prefix2 := "foo.bar.baz?"
	cacheKey2 := prefix2 + "foo=bar"
	cacheData2 := map[string]bool{"ok": true}
	err = c.Set(cacheKey2, cacheData2)
	if err != nil {
		t.Errorf("Test set cache with redis : Error occurred [%v]", err)
	}
	result2 := map[string]bool{}
	err = c.Get(cacheKey2, &result2)
	if err != nil {
		t.Errorf("Test get cache with redis : Error occurred [%v]", err)
	}
	if res2_ok, ok := result2["ok"]; !ok || !res2_ok {
		t.Errorf("Test get cache with redis : Expected ok [true], got [%v]", ok)
		t.Errorf("Test get cache with redis : Expected res2_ok [true], got [%v]", res2_ok)
	}
	err = c.DeleteWithPrefix(prefix2)
	if err != nil {
		t.Errorf("Test delete with prefix cache with redis : Error occurred [%v]", err)
	}
	result2_deleted := map[string]bool{}
	c.Get(cacheKey2, &result2_deleted)
	if res2_ok, ok := result2_deleted["ok"]; ok || res2_ok {
		t.Errorf("Test delete with prefix cache with redis : Expected ok [false], got [%v]", ok)
		t.Errorf("Test delete with prefix cache with redis : Expected res2_ok [false], got [%v]", res2_ok)
	}
}

func BenchmarkSetGetCacheWithoutRedis(b *testing.B) {
	c := &Cache{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cacheKey, expected := newCacheTestData()
		err := c.Set(cacheKey, expected)
		if err == nil {
			result := person{}
			err = c.Get(cacheKey, &result)
		}
	}
}

func BenchmarkSetGetCacheWithRedis(b *testing.B) {
	c := &Cache{}
	c.Configure()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cacheKey, expected := newCacheTestData()
		err := c.Set(cacheKey, expected)
		if err == nil {
			result := person{}
			err = c.Get(cacheKey, &result)
		}
	}
}

func BenchmarkSetDeleteWithKeyCacheWithRedis(b *testing.B) {
	c := &Cache{}
	c.Configure()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cacheKey := uuid.NewString()
		c.Set(cacheKey, map[string]bool{"ok": true})
		c.Delete(cacheKey)
	}
}

func BenchmarkSetDeleteWithPrefixCacheWithRedis(b *testing.B) {
	prefix := "foo.bar.baz."
	c := &Cache{}
	c.Configure()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cacheKey := prefix + uuid.NewString()
		c.Set(cacheKey, map[string]bool{"ok": true})
		c.DeleteWithPrefix(prefix)
	}
}
