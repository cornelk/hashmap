# hashmap [![Build Status](https://travis-ci.org/cornelk/hashmap.svg?branch=master)](https://travis-ci.org/cornelk/hashmap) [![GoDoc](https://godoc.org/github.com/cornelk/hashmap?status.svg)](https://godoc.org/github.com/cornelk/hashmap) [![Go Report Card](https://goreportcard.com/badge/cornelk/hashmap)](https://goreportcard.com/report/github.com/cornelk/hashmap) [![codecov](https://codecov.io/gh/cornelk/hashmap/branch/master/graph/badge.svg)](https://codecov.io/gh/cornelk/hashmap)

## Overview

A Golang lock-free thread-safe HashMap optimized for fastest read access.

## Usage

Set a value for a key in the map:

```
m := &HashMap{}
i := 123
m.Set("amount", unsafe.Pointer(&i))
```

Read a value for a key from the map:
```
val, ok := m.Get("amount")
if ok {
    amount := *(*int)(val)
}
```

Use the map to count URL requests:
```
var i int64
actual, _ := m.GetOrInsert("api/123", unsafe.Pointer(&i))
counter := (*int64)(actual)
atomic.AddInt64(counter, 1) // increase counter
...
count := atomic.LoadInt64(counter) // read counter
```

## Benchmarks

Reading from the hash map in a thread-safe way is nearly as fast as reading from a standard Golang map
in an unsafe way and twice as fast as Go's `sync.Map`:

```
BenchmarkReadHashMapUint-8                	  200000	      6620 ns/op
BenchmarkReadGoMapUintUnsafe-8            	  300000	      5108 ns/op
BenchmarkReadGoMapUintMutex-8             	   30000	     40743 ns/op
BenchmarkReadGoSyncMapUint-8              	  100000	     13817 ns/op
```

If the keys of the map are already hashes, no extra hashing needs to be done by the map:

```
BenchmarkReadHashMapHashedKey-8           	 1000000	      1491 ns/op
```

Reading from the map while writes are happening:
```
BenchmarkReadHashMapWithWritesUint-8      	  200000	      8498 ns/op
BenchmarkReadGoMapWithWritesUintMutex-8   	   10000	    128451 ns/op
BenchmarkReadGoSyncMapWithWritesUint-8    	  100000	     14572 ns/op
```

Write performance without any concurrent reads:

```
BenchmarkWriteHashMapUint-8               	   10000	    180306 ns/op
BenchmarkWriteGoMapMutexUint-8            	   30000	     51930 ns/op
BenchmarkWriteGoSyncMapUint-8             	   10000	    194480 ns/op
```

The benchmarks were run with Golang 1.10.1 on MacOS.

## Technical details

* Technical design decisions have been made based on benchmarks that are stored in an external repository:
  [go-benchmark](https://github.com/cornelk/go-benchmark)

* [unsafe.Pointer](https://golang.org/pkg/unsafe/#Pointer) is used instead of interface{} for the values
  as a speed optimization.

* The library uses a sorted doubly linked list and a slice as an index into that list.

* The Get() function contains helper functions that have been inlined manually until the Golang compiler will inline them automatically.

* It optimizes the slice access by circumventing the Golang size check when reading from the slice. Once a slice is allocated, the size of it does not change.
  The library limits the index into the slice, therefor the Golang size check is obsolete. When the slice reaches a defined fill rate, a bigger slice is allocated
and all keys are recalculated and transferred into the new slice.
