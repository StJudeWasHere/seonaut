package cache

import (
	"errors"
	"reflect"
	"sync"
)

type MemCache struct {
	cache map[string]interface{}
	lock  *sync.RWMutex
}

func NewMemCache() *MemCache {
	return &MemCache{
		cache: make(map[string]interface{}),
		lock:  &sync.RWMutex{},
	}
}

// Set stores a value in the cache map.
func (c *MemCache) Set(key string, v interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("non-pointer interface passed")
	}

	c.cache[key] = v
	return nil
}

// Gets a value stored in the cache map.
func (c *MemCache) Get(key string, v interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, ok := c.cache[key]
	if !ok {
		return errors.New("key does not exist")
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("non-pointer interface passed")
	}

	if reflect.TypeOf(val) != rv.Type() || reflect.TypeOf(val).Elem() != rv.Type().Elem() {
		return errors.New("mismatched types")
	}

	reflect.ValueOf(v).Elem().Set(reflect.ValueOf(val).Elem())

	return nil
}

// Delete removes a value from the cache map.
func (c *MemCache) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.cache[key]; ok {
		delete(c.cache, key)
		return nil
	}

	return errors.New("key does not exist")
}
