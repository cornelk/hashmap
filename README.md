# hashmap [![Build Status](https://travis-ci.org/cornelk/hashmap.svg?branch=master)](https://travis-ci.org/cornelk/hashmap) [![GoDoc](https://godoc.org/github.com/cornelk/hashmap?status.svg)](https://godoc.org/github.com/cornelk/hashmap) [![Go Report Card](https://goreportcard.com/badge/cornelk/hashmap)](https://goreportcard.com/report/github.com/cornelk/hashmap) [![codecov](https://codecov.io/gh/cornelk/hashmap/branch/master/graph/badge.svg)](https://codecov.io/gh/cornelk/hashmap)

## Overview

A Golang thread-safe HashMap optimized for fastest lock-free read access on 64 bit systems.

## Benchmarks

The benchmarks are run with Golang 1.7.4 on MacOS and amd64 architecture.
Reading from the hash map in a thread-safe way is faster than reading from standard Golang map in an unsafe way:

```
BenchmarkReadHashMap-8       	 1000000	      1637 ns/op
BenchmarkReadGoMapUnsafe-8   	 1000000	      2044 ns/op
BenchmarkReadGoMap-8         	   50000	     28846 ns/op
```

Writing to the hash map vs a Mutex protected standard Golang map:

```
BenchmarkWriteHashMap-8      	  100000	     23718 ns/op
BenchmarkWriteGoMap-8        	   10000	    102594 ns/op
```

The API uses [unsafe.Pointer](https://golang.org/pkg/unsafe/#Pointer) instead of the common interface{} for the values for faster speed when reading values:

```
BenchmarkUnsafePointer-8     	20000000	       107 ns/op
BenchmarkInterface-8         	10000000	       128 ns/op
```

Hashing algorithm benchmark:

```
BenchmarkComparisonMD5-8           	 5000000	       257 ns/op	  31.06 MB/s
BenchmarkComparisonSHA1-8          	 5000000	       313 ns/op	  25.51 MB/s
BenchmarkComparisonSHA256-8        	 3000000	       562 ns/op	  14.22 MB/s
BenchmarkComparisonSHA3B224-8      	 1000000	      1359 ns/op	   5.88 MB/s
BenchmarkComparisonSHA3B256-8      	 1000000	      1579 ns/op	   5.06 MB/s
BenchmarkComparisonRIPEMD160-8     	 1000000	      1106 ns/op	   7.23 MB/s
BenchmarkComparisonBlake2B-8       	 2000000	       717 ns/op	  11.14 MB/s
BenchmarkComparisonBlake2BSimd-8   	 2000000	       628 ns/op	  12.73 MB/s
BenchmarkComparisonMurmur3-8       	10000000	       139 ns/op	  57.47 MB/s
BenchmarkComparisonSipHash-8       	10000000	       127 ns/op	  62.95 MB/s
```

## Technical details

The library uses a sorted linked list and a slice as an index into the list.

The Get() function contains helper functions that have been inlined manually until the Golang compiler will inline them automatically.

It optimizes the slice access by circumventing the Golang size check when reading from the slice. Once a slice is allocated, the size of it does not change.
The library limits the index into the slice, therefor the Golang size check is obsolete. When the slice reaches a defined fill rate, a bigger slice is allocated
and all keys are recalculated and transferred into the new slice.

The resize operation uses a lock to ensure that only one resize operation is happening. This way, no CPU and memory resources are wasted by multiple goroutines working on the resize.
