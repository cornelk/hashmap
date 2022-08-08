package hashmap

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLog2(t *testing.T) {
	var fixtures = map[uintptr]uintptr{
		0: 0,
		1: 0,
		2: 1,
		3: 2,
		4: 2,
		5: 3,
	}

	for input, result := range fixtures {
		output := log2(input)
		assert.Equal(t, output, result)
	}
}

func TestKeyHash(t *testing.T) {
	type testFixture struct {
		input  interface{}
		output uintptr
	}
	var fixtures = []testFixture{
		{input: "123", output: 4353148100880623749},
		{input: 1, output: 11468921228449061269},
		{input: int8(1), output: 11468921228449061269},
		{input: int16(-1), output: 9642548396912002761},
		{input: int32(math.MaxInt32), output: 6040406647911695984},
		{input: int64(math.MaxInt64), output: 18406436390665352972},
		{input: uint(0), output: 3803688792395291579},
		{input: uint8(123), output: 10501506991603099528},
		{input: uint16(1234), output: 15687506744758839514},
		{input: uint32(12345), output: 17744734807539824643},
		{input: uint64(123456), output: 13032357135521718877},
		{input: uintptr(1234567), output: 2006671164566717660},
	}

	for _, f := range fixtures {
		output := getKeyHash(f.input)
		assert.Equal(t, f.output, output)
	}
}

func TestKeyHashPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Invalid type did not panic")
		}
	}()

	_ = getKeyHash(true)
}
