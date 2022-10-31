module github.com/cornelk/hashmap/benchmarks

go 1.19

replace github.com/cornelk/hashmap => ../

require (
	github.com/alphadose/haxmap v1.1.0
	github.com/cornelk/hashmap v1.0.8
	github.com/puzpuzpuz/xsync v1.5.2
	github.com/zhangyunhao116/skipmap v0.10.1
)

require github.com/zhangyunhao116/fastrand v0.3.0 // indirect
