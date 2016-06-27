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

//func printHash(hash uint64) {
//	top4 := ^uint64(1<<60 - 1)
//}

//func addToHashPath(hashPath uint64, depth int, val uint8) uint64 {
//	return hashPath & uint64(val << (depth * NBITS))
//}

func path2hash(path []uint8) uint64 {
	var hashPath uint64
	for depth, val := range path {
		hashPath &= uint64(val << (uint(depth) * NBITS))
	}
	return hashPath
}

func TestMain(m *testing.M) {
	// SETUP

	// RUN
	xit := m.Run()

	// TEARDOW

	os.Exit(xit)
}

func TestEmptyPutOnce(t *testing.T) {
	key := []byte("foo")

	h, _ := EMPTY.Put(key, 1)

	val, ok := h.Get(key)

	if !ok {
		t.Fatal("failed to retrieve \"foo\"")
	}

	if val != 1 {
		t.Fatal("failed to rerieve the correct val for \"foo\"")
	}
}

func TestEmptyPutThrice(t *testing.T) {
	var keys = [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}
	var vals = []int{1, 2, 3}

	var h = EMPTY

	for i := range keys {
		t.Logf(" for i=%d calling h.Put(\"%s\", %d)\n", i, keys[i], vals[i])
		h, _ = h.Put(keys[i], vals[i])
	}

	t.LogF("h=\n%s", h.String())
	t.Logf("h.root=\n%s", h.root.String())

	for i := range vals {
		t.Logf("for i=%d calling h.Get(\"%s\")", i, keys[i])
		var val, found = h.Get(keys[i])
		t.Logf("val = %v, found = %v\n", val, found)
		if !found {
			t.Fatalf("failed to get key \"%s\" from h", keys[i])
		}
		if val != vals[i] {
			t.Fatalf("failed to get val for \"%s\" val,%d != vals[%d],%d from h", keys[i], val, i, vals[i])
		}
	}
}
