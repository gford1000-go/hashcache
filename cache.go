package hashcache

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/gford1000-go/hasher"
	"github.com/gford1000-go/lru"
	"github.com/gford1000-go/saferr"
	"github.com/gford1000-go/saferr/types"
)

type put[T any] struct {
	key   string
	value T
}

type get struct {
	key string
}

type result[T any] struct {
	value T
	err   error
}

type selector[T any] struct {
	putItem *put[T]
	getItem *get
}

// New creates a new Cache instance, with Options being available to change behaviour
func New[T any](ctx context.Context, opts ...func(*Options)) (*Cache[T], error) {

	o := Options{
		BufferSize: 100,
		MaxEntries: 0,
	}
	for _, opt := range opts {
		opt(&o)
	}

	ctxGo, cancel := context.WithCancel(context.Background())

	caches := make([]*lru.BasicCache, 16)
	var err error
	for i := range len(caches) {
		caches[i], err = lru.NewBasicCache(ctxGo, o.MaxEntries, 0)
		if err != nil {
			for j := range i {
				caches[j].Close()
			}
			cancel()
			return nil, err
		}
	}

	requestors := make([]types.Requestor[selector[T], result[T]], 16)

	for i := range len(requestors) {

		postGo := func(err error) {
			caches[i].Close()
		}

		handler := func(ctx context.Context, s *selector[T]) (*result[T], error) {
			if s.putItem != nil {
				err := caches[i].Put(ctx, s.putItem.key, s.putItem.value)
				return &result[T]{
					err: err,
				}, nil
			}
			if s.getItem != nil {
				v, _, err := caches[i].Get(ctx, s.getItem.key)
				return &result[T]{
					value: v.(T),
					err:   err,
				}, nil
			}

			return nil, errors.New("invalid request received")
		}

		requestors[i] = saferr.Go(ctxGo, handler,
			saferr.WithChanSize(o.BufferSize),
			saferr.WithGoPostEnd(postGo))
	}

	c := &Cache[T]{
		cancel: cancel,
		ht:     o.HashType,
		r:      requestors,
	}

	// Listen for external context to be completed, and tidy up cache gracefully;
	// When Delete() is called directly, then ctxGo will be completed, so goroutine will exit
	go func() {
		select {
		case <-ctx.Done():
			c.Delete()
		case <-ctxGo.Done():
			return
		}
	}()

	return c, nil
}

// Cache provides a concurrency safe in-memory cache, keyed using the value of Hasher.Hash()
type Cache[T any] struct {
	invalid atomic.Bool
	cancel  context.CancelFunc
	ht      hasher.HashType
	r       []types.Requestor[selector[T], result[T]]
}

func (c *Cache[T]) keyToString(key any) (string, int, error) {
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

func (c *Cache[T]) isInvalid() bool {
	return c.invalid.Load()
}

// ErrCacheInvalidated is returned if the Cache is no longer accepting requests
var ErrCacheInvalidated = errors.New("cache has been invalidated")

// Put adds the value to the key against the specified key
func (c *Cache[T]) Put(key any, value T) error {
	if c.isInvalid() {
		return ErrCacheInvalidated
	}

	sk, offset, err := c.keyToString(key)
	if err != nil {
		return err
	}
	s := &selector[T]{
		putItem: &put[T]{
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
func (c *Cache[T]) Get(key any) (T, error) {
	var nilT T
	if c.isInvalid() {
		return nilT, ErrCacheInvalidated
	}

	sk, offset, err := c.keyToString(key)
	if err != nil {
		return nilT, err
	}
	s := &selector[T]{
		getItem: &get{
			key: sk,
		},
	}

	result, err := c.r[offset].Send(context.Background(), s)
	if err != nil {
		return nilT, err
	}
	return result.value, result.err
}

// Delete invalidates the cache, preventing further access to the Cache and its contents
func (c *Cache[T]) Delete() {
	c.invalid.Store(true)
	c.cancel()
}
