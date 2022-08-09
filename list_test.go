package hashmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListNew(t *testing.T) {
	l := NewList[uintptr, uintptr]()
	node := l.First()
	assert.Nil(t, node)

	node = l.head.Next()
	assert.Nil(t, node)
}
