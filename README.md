[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://en.wikipedia.org/wiki/MIT_License)
[![Build Status](https://travis-ci.org/gford1000-go/hashcache.svg?branch=master)](https://travis-ci.org/gford1000-go/hashcache)
[![Documentation](https://img.shields.io/badge/Documentation-GoDoc-green.svg)](https://godoc.org/github.com/gford1000-go/hashcache)

HashCache | In-memory caching of objects, with arbitrary keys
=============================================================

The hashcache package provides a simple in-memory, non-persistent and concurrency-safe cache, where the key does not have to be comparable.

An example of use is shown below:

```go
func main() {
    // Create the cache, using sha512 hashing for the key
    cache, _ := hashcache.New[string](context.Background(), WithHashType(hasher.Sha512))
    defer cache.Delete()

    // Store some data
    cache.Put("MyKey", "Hello World")

    // Retrieve data
    resp, _ := cache.Get("MyKey")

    fmt.Println(resp) // Hello World
}
```

The `Cache` supports the use of an arbitrary key, which is then hashed to create
the internal key.  `Cache` is partitioned internally into distinct maps using
the first byte of the hash key, to improve concurrent performance.  

Installing and building the library
===================================

This project requires Go 1.24+

To use this package in your own code, install it using `go get`:

`go get github.com/gford1000-go/hashcache`

Alternatively, you can clone it yourself:

`git clone https://github.com/gford1000-go/hashcache.git`

Testing and benchmarking
========================

To run all tests, `cd` into the directory and use:

`go test -v`
