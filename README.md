# hashmap

[![Build status](https://github.com/cornelk/hashmap/actions/workflows/go.yaml/badge.svg?branch=main)](https://github.com/cornelk/hashmap/actions)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/cornelk/hashmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/cornelk/hashmap)](https://goreportcard.com/report/github.com/cornelk/hashmap)
[![codecov](https://codecov.io/gh/cornelk/hashmap/branch/main/graph/badge.svg?token=NS5UY28V3A)](https://codecov.io/gh/cornelk/hashmap)

## Overview

A Golang lock-free thread-safe HashMap optimized for fastest read access.

It requires Golang 1.19+ as it makes use of Generics and the new atomic package helpers. 

***Warning: This library is experimental, there are currently no known bugs but more battle-testing is needed.***

## Usage

Only Go comparable types are supported as keys.

Set a value for a key in the map:

```
m := New[string, int]()
m.Set("amount", 123)
```

Read a value for a key from the map:
```
value, ok := m.Get("amount")
```

Use the map to count URL requests:
```
m := New[string, *int64]()
var i int64
counter, _ := m.GetOrInsert("api/123", &i)
atomic.AddInt64(counter, 1) // increase counter
...
count := atomic.LoadInt64(counter) // read counter
```

## Benchmarks

Reading from the hash map in a thread-safe way is nearly as fast as reading from a standard Golang map
in an unsafe way and twice as fast as Go's `sync.Map`:

```
BenchmarkReadHashMapUint-8                       2047108               547.2 ns/op
BenchmarkReadHaxMapUint-8                        2295067               532.3 ns/op (not safe to use, can lose writes during concurrent deletes)
BenchmarkReadGoMapUintUnsafe-8                   3303577               360.6 ns/op
BenchmarkReadGoMapUintMutex-8                      83017             14266 ns/op
BenchmarkReadGoSyncMapUint-8                      943773              1208 ns/op
```

Reading from the map while writes are happening:
```
BenchmarkReadHashMapWithWritesUint-8             1669750               718.1 ns/op
BenchmarkReadGoMapWithWritesUintMutex-8            24919             54270 ns/op
BenchmarkReadGoSyncMapWithWritesUint-8            823063              1439 ns/op
```

Write performance without any concurrent reads:

```
BenchmarkWriteHashMapUint-8                        30104             49176 ns/op
BenchmarkWriteGoMapMutexUint-8                    177536              6571 ns/op
BenchmarkWriteGoSyncMapUint-8                      19857             61428 ns/op
```

The benchmarks were run with Golang 1.19.0 on Linux using `make benchmark`.

### Benefits over Golang's builtin map

* Faster

* thread-safe access without need of a(n extra) mutex

### Benefits over [Golang's sync.Map](https://golang.org/pkg/sync/#Map)

* Faster

## Technical details

* Technical design decisions have been made based on benchmarks that are stored in an external repository:
  [go-benchmark](https://github.com/cornelk/go-benchmark)

* The library uses a sorted linked list and a slice as an index into that list.

* The Get() function contains helper functions that have been inlined manually until the Golang compiler will inline them automatically.

* It optimizes the slice access by circumventing the Golang size check when reading from the slice.
  Once a slice is allocated, the size of it does not change.
  The library limits the index into the slice, therefore the Golang size check is obsolete.
  When the slice reaches a defined fill rate, a bigger slice is allocated and all keys are recalculated and transferred into the new slice.
