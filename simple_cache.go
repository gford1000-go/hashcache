package hashcache

import (
	"errors"
	"fmt"
)

// cache is the base cache - a simple map to the data
type simpleCache struct {
	m map[string]any
}

// put will add/overwrite data at the specified key
func (c *simpleCache) put(key string, data any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cache.put() panicked: %v", r)
		}
	}()

	c.m[key] = data
	return nil
}

// get will retreive data at the specified key, or return an error
func (c *simpleCache) get(key string) (data any, err error) {
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
