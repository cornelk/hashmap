# hashmap
A Golang thread-safe HashMap optimized for lock-free read access.

Reading from the hash map in a thread-safe way is faster than reading from standard go map in an unsafe way:

```
go test -bench=.

BenchmarkReadHashMap-4            300000              5503 ns/op
BenchmarkReadGoMap-4              200000              6609 ns/op
BenchmarkReadGoMapSafe-4           30000             42261 ns/op
```
