package hashmap

import (
	"strconv"
)

// intSizeBytes is the size in byte of an int or uint value.
const intSizeBytes = strconv.IntSize >> 3

// roundUpPower2 rounds a number to the next power of 2.
func roundUpPower2(i uint64) uint64 {
	i--
	i |= i >> 1
	i |= i >> 2
	i |= i >> 4
	i |= i >> 8
	i |= i >> 16
	i |= i >> 32
	i++
	return i
}

// log2 computes the binary logarithm of x, rounded up to the next integer.
func log2(i uint64) uint64 {
	var n, p uint64
	for p = 1; p < i; p += p {
		n++
	}
	return n
}
