//go:build !386

package hashmap

import (
	"testing"

	"github.com/cornelk/hashmap/assert"
)

func TestHashingUintptr(t *testing.T) {
	m := New[uintptr, uintptr]()
	assert.Equal(t, uintptr(0x9f29cb17a2a49995), m.hasher(1))
	assert.Equal(t, uintptr(0xeac73e4044e82db0), m.hasher(2))
}

func TestHashingUint64(t *testing.T) {
	m := New[uint64, uint64]()
	assert.Equal(t, uintptr(0x9f29cb17a2a49995), m.hasher(1))
	assert.Equal(t, uintptr(0xeac73e4044e82db0), m.hasher(2))
}

func TestHashingUint32(t *testing.T) {
	m := New[uint32, uint32]()
	assert.Equal(t, uintptr(0xf42f94001fcb5351), m.hasher(1))
	assert.Equal(t, uintptr(0x277af360cedcb29e), m.hasher(2))
}

func TestHashingUint16(t *testing.T) {
	m := New[uint16, uint16]()
	assert.Equal(t, uintptr(0xdd8f621dbf7f57f1), m.hasher(1))
	assert.Equal(t, uintptr(0xfc2f33e9edde6f4a), m.hasher(0x102))
}

func TestHashingUint8(t *testing.T) {
	m := New[uint8, uint8]()
	assert.Equal(t, uintptr(0x8a4127811b21e730), m.hasher(1))
	assert.Equal(t, uintptr(0x4b79b8c95732b0e7), m.hasher(2))
}

func TestHashingString(t *testing.T) {
	m := NewString[string, uint8]()
	assert.Equal(t, uintptr(0x6a1faf26e7da4cb9), m.hasher("properunittesting"))
	assert.Equal(t, uintptr(0x2d4ff7e12135f1f3), m.hasher("longstringlongstringlongstringlongstring"))
}
