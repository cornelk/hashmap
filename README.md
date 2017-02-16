# hashmap [![Build Status](https://travis-ci.org/cornelk/hashmap.svg?branch=master)](https://travis-ci.org/cornelk/hashmap) [![GoDoc](https://godoc.org/github.com/cornelk/hashmap?status.svg)](https://godoc.org/github.com/cornelk/hashmap) [![Go Report Card](https://goreportcard.com/badge/cornelk/hashmap)](https://goreportcard.com/report/github.com/cornelk/hashmap) [![codecov](https://codecov.io/gh/cornelk/hashmap/branch/master/graph/badge.svg)](https://codecov.io/gh/cornelk/hashmap)

## Overview

A Golang thread-safe HashMap optimized for fastest lock-free read access on 64 bit systems.

## Benchmarks

The benchmarks are run with Golang 1.7.4 on MacOS and amd64 architecture.

Reading from the hash map without any concurrent writes, slightly slower than reading from a standard Golang map in a non thread-safe way:

```
BenchmarkReadHashMapUint-8                	  200000	     10237 ns/op
BenchmarkReadGoMapUintUnsafe-8            	  200000	      6214 ns/op
BenchmarkReadGoMapUintMutex-8             	   30000	     60214 ns/op
BenchmarkReadGoSyncMapUint-8              	   50000	     27512 ns/op
```

If the keys of the map are already hashes, no extra hashing needs to be done by the map:

```
```


Reading from the map while writes are happening:
```
BenchmarkReadHashMapWithWritesUint-8      	  100000	     13735 ns/op
BenchmarkReadGoMapWithWritesUintMutex-8   	   10000	    192933 ns/op
BenchmarkReadGoSyncMapWithWritesUint-8    	   30000	     44757 ns/op
```

Pure Write performance without any reads:

```
BenchmarkWriteHashMapUint-8               	    3000	    529583 ns/op
BenchmarkWriteGoMapMutexUint-8            	   10000	    139898 ns/op
BenchmarkWriteGoSyncMapUint-8             	    3000	    433616 ns/op
```

## Technical details

* The API uses [unsafe.Pointer](https://golang.org/pkg/unsafe/#Pointer) instead of the common interface{} for the values for faster speed when reading values:

```
BenchmarkUnsafePointer-8                  	 2000000	       671 ns/op
BenchmarkInterface-8                      	 2000000	       682 ns/op
```

* The library uses a sorted linked list and a slice as an index into that list.

* The Get() function contains helper functions that have been inlined manually until the Golang compiler will inline them automatically.

* It optimizes the slice access by circumventing the Golang size check when reading from the slice. Once a slice is allocated, the size of it does not change.
The library limits the index into the slice, therefor the Golang size check is obsolete. When the slice reaches a defined fill rate, a bigger slice is allocated
and all keys are recalculated and transferred into the new slice.

* The resize operation uses a lock to ensure that only one resize operation is happening. This way, no CPU and memory resources are wasted by multiple goroutines working on the resize.
