package hamt

import (
	"log"
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
var hugeNumEnts []keyVal

func TestMain(m *testing.M) {
	// SETUP

	var logFile, err = os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	Lgr.SetOutput(logFile)

	midNumEnts = make([]keyVal, 0, 32)
	var s0 = util.Str("")
	//nEnts := 10000 //ten thousand
	var midEnts = 1000
	for i := 0; i < midEnts; i++ {
		s0 = s0.Inc(1) //get off "" first
		var key = []byte(s0)
		var val = i + 1
		midNumEnts = append(midNumEnts, keyVal{key, val})
	}

	hugeNumEnts = make([]keyVal, 0, 32)
	var s1 = util.Str("")
	//var hugeEnts = 1024
	var hugeEnts = 32 * 1024
	//var hugeEnts = 256 * 1024 * 1024 //256 MB
	for i := 0; i < hugeEnts; i++ {
		s1 = s1.Inc(1)
		var key = []byte(s1)
		var val = i + 1
		hugeNumEnts = append(hugeNumEnts, keyVal{key, val})
	}

	// RUN
	xit := m.Run()

	// TEARDOW

	os.Exit(xit)
}

func dTestEmptyPutDelCrazy(t *testing.T) {
	var key = []byte("aaaaaaaaaaaaaaaaaaaaaabbcdefghijkl")
	var val = 14126
	var h = &EMPTY

	h, _ = h.Put(key, val)

	var v interface{}
	var d bool
	h, v, d = h.Del(key)
	if !d {
		t.Fatalf("failed to retrieve %q", key)
	}
	if v != val {
		t.Fatalf("failed to retrieve the correct val,%d for %q", val, key)
	}

	if !h.IsEmpty() {
		t.Fatal("hash is not Empty")
	}
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

func TestEmptyPutThriceFlatGetThrice(t *testing.T) {
	var keys = [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}
	var vals = []int{1, 2, 3}

	var h *Hamt = &EMPTY

	for i := range keys {
		h, _ = h.Put(keys[i], vals[i])
	}

	t.Logf("h.root =\n%s", h.root.LongString(""))

	for i := range vals {
		var val, found = h.Get(keys[i])

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

	t.Log("h.root =\n%s", h.root.LongString(""))

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

func TestEmptyPutOnceDelOnce(t *testing.T) {
	var h = &EMPTY

	var key = []byte("a")
	var val interface{} = 1

	h, _ = h.Put(key, val)

	var v interface{}
	var deleted, found bool

	h, v, deleted = h.Del(key)
	if !deleted {
		t.Fatalf("key=%q not deleted from h.", key)
	}
	if v != val {
		t.Fatalf("Returned deleted value val,%d != v,%d .", val, v)
	}

	v, found = h.Get(key)
	if found {
		t.Fatalf("h.Get(%q) retrived a value v=%v.", key, v)
	}
}

func TestEmptyPutOnceDelOnceIsEmpty(t *testing.T) {
	var h = &EMPTY

	var key = []byte("a")
	var val interface{} = 1

	h, _ = h.Put(key, val)

	var v interface{}
	var deleted, found bool

	h, v, deleted = h.Del(key)
	if !deleted {
		t.Fatalf("key=%q not deleted from h.", key)
	}
	if v != val {
		t.Fatalf("Returned deleted value val,%d != v,%d .", val, v)
	}

	v, found = h.Get(key)
	if found {
		t.Fatalf("h.Get(%q) retrived a value v=%v.", key, v)
	}

	if !h.IsEmpty() {
		t.Fatal("NOT h.IsEmpty()")
	}
}

func TestEmptyPutThriceFlatDelThriceIsEmpty(t *testing.T) {
	var keys = [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}
	var vals = []int{1, 2, 3}

	var h *Hamt = &EMPTY

	for i := range keys {
		h, _ = h.Put(keys[i], vals[i])
	}

	for i := range vals {
		var val interface{}
		var deleted bool
		h, val, deleted = h.Del(keys[i])

		if !deleted {
			t.Fatalf("failed to delete key \"%s\" from h", keys[i])
		}
		if val != vals[i] {
			t.Fatalf("deleted val for \"%s\" val,%d != vals[%d],%d from h", keys[i], val, i, vals[i])
		}

	}

	if !h.IsEmpty() {
		t.Fatal("h is NOT empty")
	}
}

// "c":3 && "fg":38 at depth 1
func TestPutDelOneTableDeepCollisionIsEmpty(t *testing.T) {
	var h = &EMPTY

	h, _ = h.Put([]byte("c"), 3)
	h, _ = h.Put([]byte("fg"), 38)

	var val interface{}
	var deleted bool

	h, val, deleted = h.Del([]byte("c"))

	if !deleted {
		t.Error("failed to delete for key=\"c\"")
	}
	if val != 3 {
		t.Error("h.Get(\"c\") failed to retrieve val = 3")
	}

	h, val, deleted = h.Del([]byte("fg"))

	if !deleted {
		t.Error("failed to delete for key=\"fg\"")
	}
	if val != 38 {
		t.Error("h.Get(\"fg\") failed to retrieve val = 38")
	}

	if !h.IsEmpty() {
		t.Error("h is NOT empty")
	}
}

// "d":4 && "aa":27 collide at depth 2
func TestPutDelTwoTableDeepCollisionIsEmpty(t *testing.T) {
	var h = &EMPTY

	h, _ = h.Put([]byte("d"), 4)
	h, _ = h.Put([]byte("aa"), 27)

	t.Logf("h =\n%s", h.LongString(""))

	var val interface{}
	var deleted bool
	h, val, deleted = h.Del([]byte("d"))
	if !deleted {
		t.Error("failed to delete for key=\"d\"")
	}
	if val != 4 {
		t.Error("h.Get(\"d\") failed to retrieve val = 4")
	}

	t.Logf("After h.Del(%q): h =\n%s", "d", h.LongString(""))

	h, val, deleted = h.Del([]byte("aa"))
	if !deleted {
		t.Error("failed to delete for key=\"aa\"")
	}
	if val != 27 {
		t.Error("h.Get(\"d\") failed to retrieve val = 27")
	}

	t.Logf("After h.Del(%q): h =\n%s", "aa", h.LongString(""))

	if !h.IsEmpty() {
		t.Error("h is NOT empty")
	}

	return
}

// Where Many == 64
func TestEmptyPutManyDelManyIsEmpty(t *testing.T) {
	var h = &EMPTY

	for i := 0; i < 64; i++ {
		var key = midNumEnts[i].key
		var val = midNumEnts[i].val
		h, _ = h.Put(key, val)
	}

	t.Log("h.root =\n", h.root.LongString(""))

	for i := 0; i < 64; i++ {
		var key = midNumEnts[i].key
		var expected_val = midNumEnts[i].val

		var val interface{}
		var deleted bool
		h, val, deleted = h.Del(key)
		if !deleted {
			t.Errorf("Did NOT find&delete for key=\"%s\"", key)
		}
		if val != expected_val {
			t.Errorf("val,%d != expected_val,%d", val, expected_val)
		}

		if h.root == nil {
			t.Log("h.root == nil")
		} else {
			t.Log("h.root ==\n", h.root.LongString(""))
		}
	}
	t.Log("### Testing compressedTable Shrinkage ###")

	t.Log(h)

	if !h.IsEmpty() {
		t.Fatal("NOT h.IsEmpty()")
	}
}

func TestEmptyPutDelTrumpIsEmpty(t *testing.T) {
	var h = &EMPTY

	for i := 0; i < len(hugeNumEnts); i++ {
		h, _ = h.Put(hugeNumEnts[i].key, hugeNumEnts[i].val)
	}

	Lgr.Println("h.root =")
	Lgr.Println(h.root.LongString(""))

	for i := 0; i < len(hugeNumEnts); i++ {
		var key = hugeNumEnts[i].key
		var expected_val = hugeNumEnts[i].val

		var val interface{}
		var deleted bool
		h, val, deleted = h.Del(key)
		if !deleted {
			t.Errorf("Did NOT find&delete for key=\"%s\"", key)
		}
		if val != expected_val {
			t.Errorf("val,%d != expected_val,%d", val, expected_val)
		}
	}
	//t.Log("### Testing compressedTable Shrinkage ###")

	if !h.IsEmpty() {
		Lgr.Println(h.LongString(""))
		t.Fatal("NOT h.IsEmpty()")
	}
}

// collided depth 3:
//     "b",2 & "rstuvvw",670
//     "gg",39 & "yzz",152     <=== I like this one
//     "mm",51 & "efggh",283
//     "stt",169 & "abcddefgh",940
