# hashmap

[![Build status](https://github.com/cornelk/hashmap/actions/workflows/go.yaml/badge.svg?branch=main)](https://github.com/cornelk/hashmap/actions)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/cornelk/hashmap)
[![Go Report Card](https://goreportcard.com/badge/github.com/cornelk/hashmap)](https://goreportcard.com/report/github.com/cornelk/hashmap)
[![codecov](https://codecov.io/gh/cornelk/hashmap/branch/main/graph/badge.svg?token=NS5UY28V3A)](https://codecov.io/gh/cornelk/hashmap)

## Overview

A Golang lock-free thread-safe HashMap optimized for fastest read access.

## Usage

Set a value for a key in the map:

```
m := &HashMap{}
m.Set("amount", 123)
```

Read a value for a key from the map:
```
amount, ok := m.Get("amount")
```

Use the map to count URL requests:
```
var i int64
actual, _ := m.GetOrInsert("api/123", &i)
counter := (actual).(*int64)
atomic.AddInt64(counter, 1) // increase counter
...
count := atomic.LoadInt64(counter) // read counter
```

## Benchmarks

Reading from the hash map in a thread-safe way is nearly as fast as reading from a standard Golang map
in an unsafe way and twice as fast as Go's `sync.Map`:

```
BenchmarkReadHashMapUint-8                	  200000	      6830 ns/op
BenchmarkReadGoMapUintUnsafe-8            	  300000	      4280 ns/op
BenchmarkReadGoMapUintMutex-8             	   30000	     51294 ns/op
BenchmarkReadGoSyncMapUint-8              	  200000	     10351 ns/op
```

If your keys for the map are already hashes, no extra hashing needs to be done by the map:

```
BenchmarkReadHashMapHashedKey-8           	 1000000	      1692 ns/op
```

Reading from the map while writes are happening:
```
BenchmarkReadHashMapWithWritesUint-8      	  200000	      8395 ns/op
BenchmarkReadGoMapWithWritesUintMutex-8   	   10000	    143793 ns/op
BenchmarkReadGoSyncMapWithWritesUint-8    	  100000	     12221 ns/op
```

Write performance without any concurrent reads:

```
BenchmarkWriteHashMapUint-8               	   10000	    210383 ns/op
BenchmarkWriteGoMapMutexUint-8            	   30000	     53331 ns/op
BenchmarkWriteGoSyncMapUint-8             	   10000	    176903 ns/op
```

The benchmarks were run with Golang 1.10.1 on MacOS.

### Benefits over Golang's builtin map

* Faster

* thread-safe access without need of a(n extra) mutex

* [Compare-and-swap](https://en.wikipedia.org/wiki/Compare-and-swap) access for values

* Access functions for keys that are hashes and do not need to be hashed again

### Benefits over [Golang's sync.Map](https://golang.org/pkg/sync/#Map)

* Faster

* Access functions for keys that are hashes and do not need to be hashed again

## Technical details

* Technical design decisions have been made based on benchmarks that are stored in an external repository:
  [go-benchmark](https://github.com/cornelk/go-benchmark)

* The library uses a sorted doubly linked list and a slice as an index into that list.

* The Get() function contains helper functions that have been inlined manually until the Golang compiler will inline them automatically.

* It optimizes the slice access by circumventing the Golang size check when reading from the slice.
  Once a slice is allocated, the size of it does not change.
  The library limits the index into the slice, therefore the Golang size check is obsolete.
  When the slice reaches a defined fill rate, a bigger slice is allocated and all keys are recalculated and transferred into the new slice.
