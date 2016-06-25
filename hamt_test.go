package hamt

import (
	"os"
	"testing"
)

type entry struct {
	hashPath []uint8
	hash     uint64
	key      []byte
	val      interface{}
}

var TEST_SET_1 = []entry{
	{hash: path2hash([]uint8{1, 2, 3}), key: []byte("foo"), val: 1},
	{hash: path2hash([]uint8{1, 2, 1}), key: []byte("foo"), val: 1},
}

func printHash(hash uint64) {
	top4 := ^uint64(1<<60 - 1)
}

func addToHashPath(hashPath uint64, depth int, val uint8) uint64 {
	return hashPath & val << (depth * NBITS)
}

func path2hash(path []uint8) uint64 {
	var hashPath uint64
	for depth, val := range path {
		hashPath = hashPath & val << (depth * NBITS)
	}
}

func TestMain(m *testing.M) {
	// SETUP

	// RUN
	xit := m.Run()

	// TEARDOW

	os.Exit(0)
}

func TestEmptyPutOnce(t *testing.T) {
	key := []byte("foo")

	h := EMPTY.put(key, 1)

	val, ok := h.get(key)

	if !ok {
		t.Fail("failed to retrieve \"foo\"")
	}

	if val != 1 {
		t.Fail("failed to rerieve the correct val for \"foo\"")
	}
}
