module github.com/cornelk/hashmap/benchmarks

go 1.19

replace github.com/cornelk/hashmap => ../

require (
	github.com/alphadose/haxmap v0.3.2-0.20220906124448-c6ee4c0e1f5d
	github.com/cornelk/hashmap v1.0.8
	github.com/zhangyunhao116/skipmap v0.9.1
)

require github.com/zhangyunhao116/fastrand v0.2.1 // indirect
