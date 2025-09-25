package hashcache

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/gford1000-go/hasher"
	"github.com/gford1000-go/saferr"
	"github.com/gford1000-go/saferr/types"
)

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

// New creates a new Cache instance, with Options being available to change behaviour
func New(ctx context.Context, opts ...func(*Options)) (*Cache, error) {

	o := Options{
		BufferSize: 100,
	}
	for _, opt := range opts {
		opt(&o)
	}

	ctx, cancel := context.WithCancel(ctx)

	requestors := make([]types.Requestor[selector, result], 16)

	for i := range len(requestors) {
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

		requestors[i] = saferr.Go(ctx, handler, saferr.WithChanSize(o.BufferSize))
	}

	return &Cache{
		cancel: cancel,
		ht:     o.HashType,
		r:      requestors,
	}, nil
}

// Cache provides a concurrency safe in-memory cache, keyed using the value of Hasher.Hash()
type Cache struct {
	invalid atomic.Bool
	cancel  context.CancelFunc
	ht      hasher.HashType
	r       []types.Requestor[selector, result]
}

func (c *Cache) keyToString(key any) (string, int, error) {
	b, err := hasher.Hash(key, hasher.WithHashType(c.ht))
	if err != nil {
		return "", -1, err
	}
	s := hex.EncodeToString(b)
	first := s[0]

	switch {
	case first >= 'a' && first <= 'f':
		return s, int(first-'a') + 10, nil // a=10, b=11, c=12, d=13, e=14, f=15
	case first >= '0' && first <= '9':
		return s, int(first - '0'), nil // 0=0, 1=1, ..., 9=9
	default:
		return s, -1, fmt.Errorf("invalid character '%c', must be a-e or 0-9", first)
	}
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

	sk, offset, err := c.keyToString(key)
	if err != nil {
		return err
	}
	s := &selector{
		putItem: &put{
			key:   sk,
			value: value,
		},
	}

	result, err := c.r[offset].Send(context.Background(), s)
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

	sk, offset, err := c.keyToString(key)
	if err != nil {
		return nil, err
	}
	s := &selector{
		getItem: &get{
			key: sk,
		},
	}

	result, err := c.r[offset].Send(context.Background(), s)
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
