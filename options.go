package hashcache

import "github.com/gford1000-go/hasher"

// Options provides values that can change the HashCache behaviour
type Options struct {
	BufferSize int
	HashType   hasher.HashType
	MaxEntries int
}

// WithBufferSize sets the size of the buffer to each of the internal caches
// Default: 100
func WithBufferSize(size int) func(*Options) {
	return func(o *Options) {
		if size > o.BufferSize {
			o.BufferSize = size
		}
	}
}

// WithHashType sets the HashType to be used when hashing keys
// Default: sha256
func WithHashType(ht hasher.HashType) func(*Options) {
	return func(o *Options) {
		o.HashType = ht
	}
}

// WithMaxEntries sets the max size the internal caches can each take - i.e. overall max
// number of cached objects can be 16 times this value, if keys uniformly hash.
// Default: unbounded
func WithMaxEntries(size int) func(*Options) {
	return func(o *Options) {
		if size > o.MaxEntries {
			o.MaxEntries = size
		}
	}
}
