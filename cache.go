package hashcache

import (
	"context"
	"errors"
	"log"
	"sync/atomic"

	"github.com/gford1000-go/hasher"
	"github.com/gford1000-go/saferr"
	"github.com/gford1000-go/saferr/types"
)

// Config changes the behaviour of the Cache
type Config struct {
	Log           *log.Logger
	RequestBuffer int
}

type put struct {
	key   string
	value any
}

type get struct {
	key string
}

type result struct {
	value any
	err   error
}

type selector struct {
	putItem *put
	getItem *get
}

// New creates a new Cache instance
func New(ctx context.Context, config *Config) (*Cache, error) {

	c := &Config{
		RequestBuffer: 500,
	}
	if config != nil {
		c.Log = config.Log
		if config.RequestBuffer > c.RequestBuffer {
			c.RequestBuffer = config.RequestBuffer
		}
	}

	sc := &simpleCache{
		m: map[string]any{},
	}

	handler := func(ctx context.Context, s *selector) (*result, error) {
		if s.putItem != nil {
			err := sc.put(s.putItem.key, s.putItem.value)
			return &result{
				err: err,
			}, nil
		}
		if s.getItem != nil {
			v, err := sc.get(s.getItem.key)
			return &result{
				value: v,
				err:   err,
			}, nil
		}

		return nil, errors.New("invalid request received")
	}

	ctx, cancel := context.WithCancel(ctx)

	requestor := saferr.Go(ctx, handler, saferr.WithChanSize(c.RequestBuffer))

	return &Cache{
		cancel: cancel,
		r:      requestor,
	}, nil
}

// Cache provides a concurrency safe in-memory cache, keyed using the value of Hasher.Hash()
type Cache struct {
	invalid atomic.Bool
	cancel  context.CancelFunc
	r       types.Requestor[selector, result]
}

func (c *Cache) keyToString(key any) (string, error) {
	b, err := hasher.Hash(key)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Cache) isInvalid() bool {
	return c.invalid.Load()
}

// ErrCacheInvalidated is returned if the Cache is no longer accepting requests
var ErrCacheInvalidated = errors.New("cache has been invalidated")

// Put adds the value to the key against the specified key
func (c *Cache) Put(key, value any) error {
	if c.isInvalid() {
		return ErrCacheInvalidated
	}

	sk, err := c.keyToString(key)
	if err != nil {
		return err
	}
	s := &selector{
		putItem: &put{
			key:   sk,
			value: value,
		},
	}

	result, err := c.r.Send(context.Background(), s)
	if err != nil {
		return err
	}
	return result.err
}

// Get returns the value for the specified key
func (c *Cache) Get(key any) (any, error) {
	if c.isInvalid() {
		return nil, ErrCacheInvalidated
	}

	sk, err := c.keyToString(key)
	if err != nil {
		return nil, err
	}
	s := &selector{
		getItem: &get{
			key: sk,
		},
	}

	result, err := c.r.Send(context.Background(), s)
	if err != nil {
		return nil, err
	}
	return result.value, result.err
}

// Delete invalidates the cache, preventing further access to the Cache and its contents
func (c *Cache) Delete() {
	c.invalid.Store(true)
	c.cancel()
}
