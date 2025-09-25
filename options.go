package hashcache

import "github.com/gford1000-go/hasher"

// Options provides values that can change the HashCache behaviour
type Options struct {
	BufferSize int
	HashType   hasher.HashType
}

// WithBufferSize sets the size of the buffer to each of the internal caches
func WithBufferSize(size int) func(*Options) {
	return func(o *Options) {
		if size > o.BufferSize {
			o.BufferSize = size
		}
	}
}

// WithHashType sets the HashType to be used when hashing keys
func WithHashType(ht hasher.HashType) func(*Options) {
	return func(o *Options) {
		o.HashType = ht
	}
}
