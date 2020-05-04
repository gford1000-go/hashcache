package hashcache

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gford1000-go/hasher"

	"github.com/gford1000-go/chanmgr"
)

// cachePut combines Cache.Put() args so that they can be passed via chanmgr.InOut
type cachePut struct {
	key  hasher.DataHash
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

		// These are the internal caches, with no synchronisation
		cache := &simpleCache{m: make(map[hasher.DataHash]interface{})}

		// chanmgr provides synchronisation via channels
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
					key, ok := i.(hasher.DataHash)
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

// Put adds or overwrites the value at the specified key
func (c *Cache) Put(key, value interface{}) error {
	h, err := hasher.Hash(key)
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
	h, err := hasher.Hash(key)
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
