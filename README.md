[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://en.wikipedia.org/wiki/MIT_License)
[![Build Status](https://travis-ci.org/gford1000-go/hashcache.svg?branch=master)](https://travis-ci.org/gford1000-go/hashcache)
[![Documentation](https://img.shields.io/badge/Documentation-GoDoc-green.svg)](https://godoc.org/github.com/gford1000-go/hashcache)


HashCache | Geoff Ford
======================

The hashcache package provides a simple in-memory, non-persistent cache.  

An example of use is available in GoDocs.

The `Cache` supports the use of an arbitrary key, which is then hashed to create
the internal key.  `Cache` internally partitions data into distinct maps using
the first byte of the hash key, to improve concurrent performance.  Each partition
then uses a `chanmgr` to add and retrieve data for a key.


Installing and building the library
===================================

This project requires Go 1.14.2

To use this package in your own code, install it using `go get`:

    go get github.com/gford1000-go/hashcache

Then, you can include it in your project:

	import "github.com/gford1000-go/hashcache"

Alternatively, you can clone it yourself:

    git clone https://github.com/gford1000-go/hashcache.git

Testing and benchmarking
========================

To run all tests, `cd` into the directory and use:

	go test -v

