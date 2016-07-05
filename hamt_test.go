package hamt

import (
	"os"
	"testing"

	"github.com/lleo/util"
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

var midNumEnts []keyVal

func TestMain(m *testing.M) {
	// SETUP

	midNumEnts = make([]keyVal, 0, 32) //binary growth
	var s = util.Str("")
	//nEnts := 10000 //ten thousand
	nEnts := 1000
	for i := 0; i < nEnts; i++ {
		s = s.Inc(1) //get off "" first
		var key = []byte(s)
		var val = i + 1
		midNumEnts = append(midNumEnts, keyVal{key, val})
	}

	// RUN
	xit := m.Run()

	// TEARDOW

	os.Exit(xit)
}

func TestEmptyPutOnceGetOnce(t *testing.T) {
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

func TestEmptyPutThriceGetThrice(t *testing.T) {
	var keys = [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}
	var vals = []int{1, 2, 3}

	var h *Hamt = &EMPTY

	for i := range keys {
		t.Logf("for i=%d calling h.Put(\"%s\", %d)\n", i, keys[i], vals[i])
		h, _ = h.Put(keys[i], vals[i])
		t.Logf("after i=%d calling h.Put(\"%s\", %d) h=\n%s", i, keys[i], vals[i], h)
	}

	t.Logf("h=\n%s", h.String())

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

// "d":4 && "aa":27 collide at depth 0 & 1
func TestPutGetTwoTableDeepCollision(t *testing.T) {
	var h = &EMPTY

	h, _ = h.Put([]byte("d"), 4)
	h, _ = h.Put([]byte("aa"), 27)
	t.Logf("h.root = %s", h.root.LongString(""))

	var val interface{}
	var found bool
	val, found = h.Get([]byte("d"))
	if !found {
		t.Error("failed to find val for key=\"d\"")
	}
	if val != 4 {
		t.Error("h.Get(\"d\") failed to retrieve val = 4")
	}

	val, found = h.Get([]byte("aa"))
	if !found {
		t.Error("failed to find val for key=\"aa\"")
	}
	if val != 27 {
		t.Error("h.Get(\"d\") failed to retrieve val = 27")
	}

	return
}

// Where Many == 64
func TestEmptyPutManyGetMany(t *testing.T) {
	var h = &EMPTY

	for i := 0; i < 64; i++ {
		var key = midNumEnts[i].key
		var val = midNumEnts[i].val
		h, _ = h.Put(key, val)
	}

	t.Log("h = ", h)

	for i := 0; i < 64; i++ {
		var key = midNumEnts[i].key
		var expected_val = midNumEnts[i].val

		var val, found = h.Get(key)
		if !found {
			t.Errorf("Did NOT find val for key=\"%s\"", key)
		}
		if val != expected_val {
			t.Errorf("val,%d != expected_val,%d", val, expected_val)
		}
	}
}
