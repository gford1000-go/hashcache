package hashcache

import (
	"errors"
	"fmt"
)

// cacheKey is the type used in the simpleCache map
type cacheKey [32]byte

// invalidCacheKey indicates a bad key
var invalidCacheKey cacheKey

// cache is the base cache - a simple map to the data
type simpleCache struct {
	m map[cacheKey]interface{}
}

// put will add/overwrite data at the specified key
func (c *simpleCache) put(key cacheKey, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cache.put() panicked: %v", r)
		}
	}()

	c.m[key] = data
	return nil
}

// get will retreive data at the specified key, or return an error
func (c *simpleCache) get(key cacheKey) (data interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cache.put() panicked: %v", r)
		}
	}()

	data, ok := c.m[key]
	if !ok {
		err = errors.New("Key not found")
	}
	return data, err
}
