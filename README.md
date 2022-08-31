# hashmap

[![Build status](https://github.com/cornelk/hashmap/actions/workflows/go.yaml/badge.svg?branch=main)](https://github.com/cornelk/hashmap/actions)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/cornelk/hashmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/cornelk/hashmap)](https://goreportcard.com/report/github.com/cornelk/hashmap)
[![codecov](https://codecov.io/gh/cornelk/hashmap/branch/main/graph/badge.svg?token=NS5UY28V3A)](https://codecov.io/gh/cornelk/hashmap)

## Overview

A Golang lock-free thread-safe HashMap optimized for fastest read access.

It is not a general-use HashMap and currently has slow write performance for write heavy uses.

The minimal supported Golang version is 1.19 as it makes use of Generics and the new atomic package helpers.

## Usage

For `New()` only Go numeric types are supported. For string keyed maps `NewString` has to be used.

Example uint8 key map uses:

```
m := New[uint8, int]()
m.Set(1, 123)
value, ok := m.Get(1)
```

Example string key map uses:

```
m := NewString[string, int]()
m.Set("amount", 123)
value, ok := m.Get("amount")
```

Using the map to count URL requests:
```
m := NewString[string, *int64]()
var i int64
counter, _ := m.GetOrInsert("api/123", &i)
atomic.AddInt64(counter, 1) // increase counter
...
count := atomic.LoadInt64(counter) // read counter
```

## Benchmarks

Reading from the hash map for numeric key types in a thread-safe way is faster than reading from a standard Golang map
in an unsafe way and four times faster than Golang's `sync.Map`:

```
BenchmarkReadHashMapUint-8                	 1788601	       668.4 ns/op
BenchmarkReadHaxMapUint-8                 	 1691654	       709.6 ns/op
BenchmarkReadGoMapUintUnsafe-8            	 1516452	       784.4 ns/op
BenchmarkReadGoMapUintMutex-8             	   39429	     27978 ns/op
BenchmarkReadGoSyncMapUint-8              	  446930	      2544 ns/op
```

Reading from the map while writes are happening:
```
BenchmarkReadHashMapWithWritesUint-8      	 1418299	       856.4 ns/op
BenchmarkReadHaxMapWithWritesUint-8       	 1262414	       948.5 ns/op
BenchmarkReadGoSyncMapWithWritesUint-8    	  382785	      3240 ns/op
```

Write performance without any concurrent reads:

```
BenchmarkWriteHashMapUint-8               	   54756	     21977 ns/op
BenchmarkWriteGoMapMutexUint-8            	   83907	     14827 ns/op
BenchmarkWriteGoSyncMapUint-8             	   16983	     70305 ns/op
```

The benchmarks were run with Golang 1.19.0 on Linux and AMD64 using `make benchmark`.

## Technical details

* Technical design decisions have been made based on benchmarks that are stored in an external repository:
  [go-benchmark](https://github.com/cornelk/go-benchmark)

* The library uses a sorted linked list and a slice as an index into that list.

* The Get() function contains helper functions that have been inlined manually until the Golang compiler will inline them automatically.

* It optimizes the slice access by circumventing the Golang size check when reading from the slice.
  Once a slice is allocated, the size of it does not change.
  The library limits the index into the slice, therefore the Golang size check is obsolete.
  When the slice reaches a defined fill rate, a bigger slice is allocated and all keys are recalculated and transferred into the new slice.

* For hashing, specialized xxhash implementations are used that match the size of the key type where available

* A specialized String version of the map exists due to a limitation of type switches of parametric types - see https://github.com/golang/go/issues/45380 for more info.
