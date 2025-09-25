package hashcache

import (
	"errors"
	"fmt"
)

func newSimpleCache[T any]() *simpleCache[T] {
	return &simpleCache[T]{
		m: map[string]T{},
	}
}

// cache is the base cache - a simple map to the data
type simpleCache[T any] struct {
	m map[string]T
}

// put will add/overwrite data at the specified key
func (c *simpleCache[T]) put(key string, data T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cache.put() panicked: %v", r)
		}
	}()

	c.m[key] = data
	return nil
}

// get will retreive data at the specified key, or return an error
func (c *simpleCache[T]) get(key string) (data T, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cache.get() panicked: %v", r)
		}
	}()

	data, ok := c.m[key]
	if !ok {
		err = errors.New("key not found")
	}
	return data, err
}
