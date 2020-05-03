package hashcache

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gford1000-go/chanmgr"
)

// cachePut combines Cache.Put() args so that they can be passed via chanmgr.InOut
type cachePut struct {
	key  cacheKey
	data interface{}
}

// cacheInstance holds details of the underlying cache for a given key in Cache
type cacheInstance struct {
	exit  chanmgr.ExitChannel
	chans []*chanmgr.InOut
}

// Cache is a key/value in-memory, non-persistent store of data
// Cache will hash the provided key and use this value as the key to the data,
// which allows any type to be used as the key.
// The Cache uses 256 inner caches, mapped to the first byte of the hash value,
// to minimise contention.
type Cache struct {
	cache map[byte]*cacheInstance
}

// Config changes the behaviour of the Cache
type Config struct {
	Log           *log.Logger
	RequestBuffer int
}

// defaultConfig used if Config is not specified in New()
var defaultConfig *Config = &Config{
	Log:           log.New(ioutil.Discard, "", 0),
	RequestBuffer: 10,
}

// New creates a new Cache instance
func New(config *Config) (*Cache, error) {

	c := defaultConfig
	if config != nil {
		if config.Log != nil {
			c.Log = config.Log
		}
		if config.RequestBuffer > 0 {
			c.RequestBuffer = config.RequestBuffer
		}
	}

	// Prep the map of byte->cache
	m := make(map[byte]*cacheInstance)

	for i := 0; i < 256; i++ {

		cache := &simpleCache{m: make(map[cacheKey]interface{})}

		chans := []*chanmgr.InOut{
			&chanmgr.InOut{
				Processor: func(i interface{}) (interface{}, error) {
					p, ok := i.(*cachePut)
					if !ok {
						return nil, errors.New("Internal error")
					}
					return nil, cache.put(p.key, p.data)
				},
				WantResponse: chanmgr.WantResponse,
			},
			&chanmgr.InOut{
				Processor: func(i interface{}) (interface{}, error) {
					key, ok := i.(cacheKey)
					if !ok {
						return nil, errors.New("Internal error")
					}
					return cache.get(key)
				},
				WantResponse: chanmgr.WantResponse,
			},
		}

		mgrConfig := &chanmgr.Config{
			Log:           c.Log,
			RequestBuffer: c.RequestBuffer,
		}

		exit, err := chanmgr.New(chans, nil, mgrConfig)
		if err != nil {
			return nil, fmt.Errorf("Failed to create cache[%d]: %v", i, err)
		}
		m[byte(i)] = &cacheInstance{
			exit:  exit,
			chans: chans,
		}
	}

	return &Cache{
		cache: m,
	}, nil
}

func (c *Cache) keyToBytes(key interface{}) (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("keyToBytes() panicked: %v", r)
		}
	}()

	var stream bytes.Buffer
	enc := gob.NewEncoder(&stream)
	e := enc.Encode(key)
	if e != nil {
		return nil, fmt.Errorf("Failed to covert %v to []byte: %v", key, e)
	}
	return stream.Bytes(), nil
}

// getCacheKey converts the key to a []byte and then hashes with sha256
func (c *Cache) getCacheKey(key interface{}) (cacheKey, error) {
	b, err := c.keyToBytes(key)
	if err != nil {
		return invalidCacheKey, fmt.Errorf("Invalid key: %v", err)
	}
	return sha256.Sum256(b), nil
}

// Put adds or overwrites the value at the specified key
func (c *Cache) Put(key, value interface{}) error {
	h, err := c.getCacheKey(key)
	if err != nil {
		return err
	}

	_, err = c.cache[h[0]].chans[0].SendRecv(&cachePut{
		key:  h,
		data: value,
	})

	if err != nil {
		return fmt.Errorf("Manager.Put() error: %v", err)
	}

	return nil
}

// Get returns the value at the specified key (error if not found)
func (c *Cache) Get(key interface{}) (interface{}, error) {
	h, err := c.getCacheKey(key)
	if err != nil {
		return nil, err
	}

	resp, err := c.cache[h[0]].chans[1].SendRecv(h)

	if err != nil {
		return nil, fmt.Errorf("Manager.Get() error: %v", err)
	}

	return resp, nil
}

// Delete terminates the Cache, which should not be used afterwards
func (c *Cache) Delete() {
	for _, instance := range c.cache {
		instance.exit <- chanmgr.Exit
	}
	c.cache = nil
}
