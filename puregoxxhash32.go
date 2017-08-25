// Pure Go implementation from vova616
// https://github.com/vova616/xxhash

package hashmap

import (
	"errors"
	"hash"
)

const (
	xxhPRIME32_1 = 2654435761
	xxhPRIME32_2 = 2246822519
	xxhPRIME32_3 = 3266489917
	xxhPRIME32_4 = 668265263
	xxhPRIME32_5 = 374761393
)

type xxHash struct {
	seed, v1, v2, v3, v4 uint32
	total_len            uint64
	memory               [16]byte
	memsize              int
}

// Pure Go implementation of New32()
func XXHash_GoNew32() hash.Hash32 {
	return XXHash_GoNew32Seed(0)
}

// Pure Go implementation of GoNew32Seed()
func XXHash_GoNew32Seed(seed uint32) hash.Hash32 {
	return &xxHash{
		seed: seed,
		v1:   seed + xxhPRIME32_1 + xxhPRIME32_2,
		v2:   seed + xxhPRIME32_2,
		v3:   seed,
		v4:   seed - xxhPRIME32_1,
	}
}

func (self *xxHash) BlockSize() int {
	return 1
}

// Size returns the number of bytes Sum will return.
func (self *xxHash) Size() int {
	return 4
}

func (self *xxHash) feed(in []byte) uint32 {
	p := 0
	bEnd := len(in)

	self.total_len += uint64(bEnd)

	// fill in tmp buffer
	if self.memsize+bEnd < 16 {
		copy(self.memory[self.memsize:], in)
		self.memsize += bEnd
		return 0
	}

	if self.memsize > 0 {
		copy(self.memory[self.memsize:], in[:16-self.memsize])
		var p32 uint32 = uint32(self.memory[0]) |
			(uint32(self.memory[1]) << 8) |
			(uint32(self.memory[2]) << 16) |
			(uint32(self.memory[3]) << 24)

		self.v1 += p32 * xxhPRIME32_2
		self.v1 = ((self.v1 << 13) | (self.v1 >> (32 - 13))) * xxhPRIME32_1

		p32 = uint32(self.memory[4]) |
			(uint32(self.memory[5]) << 8) |
			(uint32(self.memory[6]) << 16) |
			(uint32(self.memory[7]) << 24)

		self.v2 += p32 * xxhPRIME32_2
		self.v2 = ((self.v2 << 13) | (self.v2 >> (32 - 13))) * xxhPRIME32_1

		p32 = uint32(self.memory[8]) |
			(uint32(self.memory[9]) << 8) |
			(uint32(self.memory[10]) << 16) |
			(uint32(self.memory[11]) << 24)

		self.v3 += p32 * xxhPRIME32_2
		self.v3 = ((self.v3 << 13) | (self.v3 >> (32 - 13))) * xxhPRIME32_1

		p32 = uint32(self.memory[12]) |
			(uint32(self.memory[13]) << 8) |
			(uint32(self.memory[14]) << 16) |
			(uint32(self.memory[15]) << 24)

		self.v4 += p32 * xxhPRIME32_2
		self.v4 = ((self.v4 << 13) | (self.v4 >> (32 - 13))) * xxhPRIME32_1

		s := 16 - self.memsize
		p += s
		self.memsize = 0
	}

	limit := bEnd - 16
	v1, v2, v3, v4 := self.v1, self.v2, self.v3, self.v4

	for ; p <= limit; p += 16 {
		var p32 uint32 = uint32(in[p]) |
			(uint32(in[p+1]) << 8) |
			(uint32(in[p+2]) << 16) |
			(uint32(in[p+3]) << 24)

		v1 += p32 * xxhPRIME32_2
		v1 = ((v1 << 13) | (v1 >> (32 - 13))) * xxhPRIME32_1

		p32 = uint32(in[p+4]) |
			(uint32(in[p+5]) << 8) |
			(uint32(in[p+6]) << 16) |
			(uint32(in[p+7]) << 24)

		v2 += p32 * xxhPRIME32_2
		v2 = ((v2 << 13) | (v2 >> (32 - 13))) * xxhPRIME32_1

		p32 = uint32(in[p+8]) |
			(uint32(in[p+9]) << 8) |
			(uint32(in[p+10]) << 16) |
			(uint32(in[p+11]) << 24)

		v3 += p32 * xxhPRIME32_2
		v3 = ((v3 << 13) | (v3 >> (32 - 13))) * xxhPRIME32_1

		p32 = uint32(in[p+12]) |
			(uint32(in[p+13]) << 8) |
			(uint32(in[p+14]) << 16) |
			(uint32(in[p+15]) << 24)

		v4 += p32 * xxhPRIME32_2
		v4 = ((v4 << 13) | (v4 >> (32 - 13))) * xxhPRIME32_1
	}

	limit = bEnd - p

	if limit > 0 {
		copy(self.memory[:], in[p:bEnd])
		self.memsize = limit
	}

	self.v1 = v1
	self.v2 = v2
	self.v3 = v3
	self.v4 = v4

	return 0
}

func (self *xxHash) Sum32() uint32 {
	p := 0
	bEnd := self.memsize
	h32 := uint32(0)

	if self.total_len >= 16 {
		h32 = ((self.v1 << 1) | (self.v1 >> (32 - 1))) +
			((self.v2 << 7) | (self.v2 >> (32 - 7))) +
			((self.v3 << 12) | (self.v3 >> (32 - 12))) +
			((self.v4 << 18) | (self.v4 >> (32 - 18)))
	} else {
		h32 = self.seed + xxhPRIME32_5
	}

	h32 += uint32(self.total_len)

	for p <= bEnd-4 {
		var p32 uint32 = uint32(self.memory[p]) |
			(uint32(self.memory[p+1]) << 8) |
			(uint32(self.memory[p+2]) << 16) |
			(uint32(self.memory[p+3]) << 24)
		h32 += p32 * xxhPRIME32_3
		h32 = ((h32 << 17) | (h32 >> (32 - 17))) * xxhPRIME32_4
		p += 4
	}

	for p < bEnd {
		h32 += uint32(self.memory[p]) * xxhPRIME32_5
		h32 = ((h32 << 11) | (h32 >> (32 - 11))) * xxhPRIME32_1
		p++
	}

	h32 ^= h32 >> 15
	h32 *= xxhPRIME32_2
	h32 ^= h32 >> 13
	h32 *= xxhPRIME32_3
	h32 ^= h32 >> 16

	return h32
}

func (self *xxHash) Sum(in []byte) []byte {
	h := self.Sum32()
	in = append(in, byte(h>>24))
	in = append(in, byte(h>>16))
	in = append(in, byte(h>>8))
	in = append(in, byte(h))
	return in
}

func (self *xxHash) Reset() {
	seed := self.seed
	self.v1 = seed + xxhPRIME32_1 + xxhPRIME32_2
	self.v2 = seed + xxhPRIME32_2
	self.v3 = seed
	self.v4 = seed - xxhPRIME32_1
	self.total_len = 0
	self.memsize = 0
}

// Write adds more data to the running hash.
// Length of data MUST BE less than 1 Gigabytes.
func (self *xxHash) Write(data []byte) (nn int, err error) {
	l := len(data)
	if l > 1<<30 {
		return 0, errors.New("Cannot add more than 1 Gigabytes at once.")
	}
	self.feed(data)
	return len(data), nil
}

// Pure Go implementation of Checksum32()
func XXHash_GoChecksum32(data []byte) uint32 {
	return XXHash_GoChecksum32Seed(data, 0)
}

// Pure Go implementation of Checksum32Seed()
func XXHash_GoChecksum32Seed(data []byte, seed uint32) uint32 {
	p := 0
	bEnd := len(data)
	h32 := uint32(0)

	if bEnd >= 16 {
		limit := bEnd - 16

		v1 := seed + xxhPRIME32_1 + xxhPRIME32_2
		v2 := seed + xxhPRIME32_2
		v3 := seed + 0
		v4 := seed - xxhPRIME32_1
		for {
			var p32 uint32 = uint32(data[p]) |
				(uint32(data[p+1]) << 8) |
				(uint32(data[p+2]) << 16) |
				(uint32(data[p+3]) << 24)

			v1 += p32 * xxhPRIME32_2
			v1 = ((v1 << 13) | (v1 >> (32 - 13))) * xxhPRIME32_1

			p32 = uint32(data[p+4]) |
				(uint32(data[p+5]) << 8) |
				(uint32(data[p+6]) << 16) |
				(uint32(data[p+7]) << 24)

			v2 += p32 * xxhPRIME32_2
			v2 = ((v2 << 13) | (v2 >> (32 - 13))) * xxhPRIME32_1

			p32 = uint32(data[p+8]) |
				(uint32(data[p+9]) << 8) |
				(uint32(data[p+10]) << 16) |
				(uint32(data[p+11]) << 24)

			v3 += p32 * xxhPRIME32_2
			v3 = ((v3 << 13) | (v3 >> (32 - 13))) * xxhPRIME32_1

			p32 = uint32(data[p+12]) |
				(uint32(data[p+13]) << 8) |
				(uint32(data[p+14]) << 16) |
				(uint32(data[p+15]) << 24)

			v4 += p32 * xxhPRIME32_2
			v4 = ((v4 << 13) | (v4 >> (32 - 13))) * xxhPRIME32_1

			p += 16

			if p > limit {
				break
			}
		}
		h32 = ((v1 << 1) | (v1 >> (32 - 1))) +
			((v2 << 7) | (v2 >> (32 - 7))) +
			((v3 << 12) | (v3 >> (32 - 12))) +
			((v4 << 18) | (v4 >> (32 - 18)))
	} else {
		h32 = seed + xxhPRIME32_5
	}

	h32 += uint32(bEnd)

	for p <= bEnd-4 {
		var p32 uint32 = uint32(data[p]) |
			(uint32(data[p+1]) << 8) |
			(uint32(data[p+2]) << 16) |
			(uint32(data[p+3]) << 24)
		h32 += p32 * xxhPRIME32_3
		h32 = ((h32 << 17) | (h32 >> (32 - 17))) * xxhPRIME32_4
		p += 4
	}

	for p < bEnd {
		h32 += uint32(data[p]) * xxhPRIME32_5
		h32 = ((h32 << 11) | (h32 >> (32 - 11))) * xxhPRIME32_1
		p++
	}

	h32 ^= h32 >> 15
	h32 *= xxhPRIME32_2
	h32 ^= h32 >> 13
	h32 *= xxhPRIME32_3
	h32 ^= h32 >> 16

	return h32
}
