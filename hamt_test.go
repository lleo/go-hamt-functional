package hamt_functional

import (
	"log"
	"math/rand"
	"os"
	"testing"

	hamt64 "github.com/lleo/go-hamt-functional/hamt64_functional"
	"github.com/lleo/go-hamt/string_key"

	"github.com/lleo/stringutil"
)

var genRandomizedKvs func(kvs []keyVal) []keyVal

var numMidKvs int
var numHugeKvs int
var midKvs []keyVal
var hugeKvs []keyVal

var M map[string]int
var H hamt64.Hamt

func TestMain(m *testing.M) {
	// SETUP
	genRandomizedKvs = genRandomizedKvsInPlace

	var logFile, err = os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	//Lgr.SetOutput(logFile)

	midKvs = make([]keyVal, 0, 32)
	var s0 = stringutil.Str("aaa")
	//numMidKvs := 10000 //ten thousand
	numMidKvs = 1000 // 10 million
	for i := 0; i < numMidKvs; i++ {
		var key = string_key.StringKey(s0)
		var val = i + 1
		midKvs = append(midKvs, keyVal{key, val})
		s0 = s0.DigitalInc(1) //get off "" first
	}

	hugeKvs = make([]keyVal, 0, 32)
	var s1 = stringutil.Str("aaa")
	//numHugeKvs = 1024
	numHugeKvs = 1 * 1024 * 1024 // one mega-entries
	//numHugeKvs = 256 * 1024 * 1024 //256 MB
	for i := 0; i < numHugeKvs; i++ {
		var key = string_key.StringKey(s1)
		var val = i + 1
		hugeKvs = append(hugeKvs, keyVal{key, val})
		s1 = s1.DigitalInc(1)
	}

	// Build map & hamt
	M = make(map[string]int)
	H = hamt64.EMPTY
	var s = stringutil.Str("aaa")
	for i := 0; i < numHugeKvs; i++ {
		M[string(s)] = i + 1
		H, _ = H.Put(string_key.StringKey(s), i+1)
		s = s.DigitalInc(1)
	}

	// RUN
	xit := m.Run()

	// TEARDOWN

	os.Exit(xit)
}

//First genRandomizedKvs() copies []keyVal passed in. Then it randomizes that
//copy in-place. Finnally, it returns the randomized copy.
func genRandomizedKvsInPlace(kvs []keyVal) []keyVal {
	randKvs := make([]keyVal, len(kvs))
	copy(randKvs, kvs)

	//From: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle#The_modern_algorithm
	for i := len(randKvs) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		randKvs[i], randKvs[j] = randKvs[j], randKvs[i]
	}

	return randKvs
}

func TestHamt64PutDelCrazy(t *testing.T) {
	var key = string_key.StringKey("aaaaaaaaaaaaaaaaaaaaaabbcdefghijkl")
	var val = 14126
	var h = hamt64.EMPTY

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

func TestHamt64PutOnceGetOnce(t *testing.T) {
	key := string_key.StringKey("foo")

	h, _ := hamt64.EMPTY.Put(key, 1)

	val, ok := h.Get(key)

	if !ok {
		t.Fatal("failed to retrieve \"foo\"")
	}

	if val != 1 {
		t.Fatal("failed to rerieve the correct val for \"foo\"")
	}
}

func TestHamt64PutThriceFlatGetThrice(t *testing.T) {
	var keys = []string_key.StringKey{string_key.StringKey("foo"), string_key.StringKey("bar"), string_key.StringKey("baz")}
	var vals = []int{1, 2, 3}

	var h = hamt64.EMPTY

	for i := range keys {
		h, _ = h.Put(keys[i], vals[i])
	}

	t.Logf("h =\n%s", h.LongString(""))

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
func TestHamt64PutGetTwoTableDeepCollision(t *testing.T) {
	var h = hamt64.EMPTY

	h, _ = h.Put(string_key.StringKey("d"), 4)
	h, _ = h.Put(string_key.StringKey("aa"), 27)

	t.Log("h =\n%s", h.LongString(""))

	var val interface{}
	var found bool
	val, found = h.Get(string_key.StringKey("d"))
	if !found {
		t.Error("failed to find val for key=\"d\"")
	}
	if val != 4 {
		t.Error("h.Get(\"d\") failed to retrieve val = 4")
	}

	val, found = h.Get(string_key.StringKey("aa"))
	if !found {
		t.Error("failed to find val for key=\"aa\"")
	}
	if val != 27 {
		t.Error("h.Get(\"d\") failed to retrieve val = 27")
	}

	return
}

// Where Many == 64
func TestHamt64PutManyGetMany(t *testing.T) {
	var h = hamt64.EMPTY

	for i := 0; i < 64; i++ {
		var key = midKvs[i].key
		var val = midKvs[i].val
		h, _ = h.Put(key, val)
	}

	for i := 0; i < 64; i++ {
		var key = midKvs[i].key
		var expected_val = midKvs[i].val

		var val, found = h.Get(key)
		if !found {
			t.Errorf("Did NOT find val for key=\"%s\"", key)
		}
		if val != expected_val {
			t.Errorf("val,%d != expected_val,%d", val, expected_val)
		}
	}
}

func TestHamt64PutOnceDelOnce(t *testing.T) {
	var h = hamt64.EMPTY

	var key = string_key.StringKey("a")
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

func TestHamt64PutOnceDelOnceIsEmpty(t *testing.T) {
	var h = hamt64.EMPTY

	var key = string_key.StringKey("a")
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

func TestHamt64PutThriceFlatDelThriceIsEmpty(t *testing.T) {
	var keys = []string_key.StringKey{string_key.StringKey("foo"), string_key.StringKey("bar"), string_key.StringKey("baz")}
	var vals = []int{1, 2, 3}

	var h = hamt64.EMPTY

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
func TestHamt64PutDelOneTableDeepCollisionIsEmpty(t *testing.T) {
	var h = hamt64.EMPTY

	h, _ = h.Put(string_key.StringKey("c"), 3)
	h, _ = h.Put(string_key.StringKey("fg"), 38)

	var val interface{}
	var deleted bool

	h, val, deleted = h.Del(string_key.StringKey("c"))

	if !deleted {
		t.Error("failed to delete for key=\"c\"")
	}
	if val != 3 {
		t.Error("h.Get(\"c\") failed to retrieve val = 3")
	}

	h, val, deleted = h.Del(string_key.StringKey("fg"))

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
func TestHamt64PutDelTwoTableDeepCollisionIsEmpty(t *testing.T) {
	var h = hamt64.EMPTY

	h, _ = h.Put(string_key.StringKey("d"), 4)
	h, _ = h.Put(string_key.StringKey("aa"), 27)

	t.Logf("h =\n%s", h.LongString(""))

	var val interface{}
	var deleted bool
	h, val, deleted = h.Del(string_key.StringKey("d"))
	if !deleted {
		t.Error("failed to delete for key=\"d\"")
	}
	if val != 4 {
		t.Error("h.Get(\"d\") failed to retrieve val = 4")
	}

	t.Logf("After h.Del(%q): h =\n%s", "d", h.LongString(""))

	h, val, deleted = h.Del(string_key.StringKey("aa"))
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
func TestHamt64PutManyDelManyIsEmpty(t *testing.T) {
	var h = hamt64.EMPTY

	for i := 0; i < 64; i++ {
		var key = midKvs[i].key
		var val = midKvs[i].val
		h, _ = h.Put(key, val)
	}

	t.Log("h =\n", h.LongString(""))

	for i := 0; i < 64; i++ {
		var key = midKvs[i].key
		var expected_val = midKvs[i].val

		var val interface{}
		var deleted bool
		h, val, deleted = h.Del(key)
		if !deleted {
			t.Fatalf("Did NOT find&delete for key=\"%s\"; hence h no longer valid", key)
		}
		if val != expected_val {
			t.Errorf("val,%d != expected_val,%d", val, expected_val)
		}

		t.Log("h =\n", h.LongString(""))
	}
	t.Log("### Testing compressedTable Shrinkage ###")

	t.Log(h)

	if !h.IsEmpty() {
		t.Fatal("NOT h.IsEmpty()")
	}
}

func TestHamt64PutGetHuge(t *testing.T) {
	var h = hamt64.EMPTY

	for i := 0; i < numHugeKvs; i++ {
		h, _ = h.Put(hugeKvs[i].key, hugeKvs[i].val)
	}

	for i := 0; i < numHugeKvs; i++ {
		var key = hugeKvs[i].key
		var val = hugeKvs[i].val
		var vv, found = h.Get(key)
		if !found {
			t.Fatalf("key,%q not found", key)
		}
		if val != vv {
			t.Fatalf("key,%q found incorrect vv,%d != val,%d", key, vv, val)
		}
	}
}

func TestHamt64PutDelHugeIsEmpty(t *testing.T) {
	var h = hamt64.EMPTY

	for i := 0; i < numHugeKvs; i++ {
		h, _ = h.Put(hugeKvs[i].key, hugeKvs[i].val)
	}

	//Lgr.Println("TEST: h = ", h.LongString(""))

	for i := 0; i < numHugeKvs; i++ {
		var key = hugeKvs[i].key
		var expected_val = hugeKvs[i].val

		//var h1 hamt64.Hamt
		//var val interface{}
		//var deleted bool
		var h1, val, deleted = h.Del(key)
		if !deleted {
			t.Errorf("Did NOT find&delete for key=\"%s\"", key)
		}
		if val != expected_val {
			t.Errorf("val,%d != expected_val,%d", val, expected_val)
		}
		h = h1
	}
	//t.Log("### Testing compressedTable Shrinkage ###")

	if !h.IsEmpty() {
		//Lgr.Println("TestEmptyPutDelTrumpIsEmpty Failed cur !h.IsEmpty()")
		//Lgr.Println(h.LongString(""))
		t.Fatal("NOT h.IsEmpty()")
	}
}

func TestHamt32PutGetHuge(t *testing.T) {
	var h = hamt32.EMPTY

	for i := 0; i < numHugeKvs; i++ {
		h, _ = h.Put(hugeKvs[i].key, hugeKvs[i].val)
	}

	for i := 0; i < numHugeKvs; i++ {
		var key = hugeKvs[i].key
		var val = hugeKvs[i].val
		var vv, found = h.Get(key)
		if !found {
			t.Fatalf("key,%q not found", key)
		}
		if val != vv {
			t.Fatalf("key,%q found incorrect vv,%d != val,%d", key, vv, val)
		}
	}
}

func dTestHamt32PutDelHugeIsEmpty(t *testing.T) {
	var h = hamt32.EMPTY

	for i := 0; i < numHugeKvs; i++ {
		h, _ = h.Put(hugeKvs[i].key, hugeKvs[i].val)
	}

	//Lgr.Println("TEST: h = ", h.LongString(""))

	for i := 0; i < numHugeKvs; i++ {
		var key = hugeKvs[i].key
		var expected_val = hugeKvs[i].val

		//var h1 hamt64.Hamt
		//var val interface{}
		//var deleted bool
		var h1, val, deleted = h.Del(key)
		if !deleted {
			t.Errorf("Did NOT find&delete for key=\"%s\"", key)
		}
		if val != expected_val {
			t.Errorf("val,%d != expected_val,%d", val, expected_val)
		}
		h = h1
	}
	//t.Log("### Testing compressedTable Shrinkage ###")

	if !h.IsEmpty() {
		//Lgr.Println("TestEmptyPutDelTrumpIsEmpty Failed cur !h.IsEmpty()")
		//Lgr.Println(h.LongString(""))
		t.Fatal("NOT h.IsEmpty()")
	}
}

//func TestMapPut(t *testing.T) {
//	var m = make(map[string]int)
//	for i := 0; i < numHugeKvs; i++ {
//		m[string(hugeKvs[i].key)] = i + 1
//	}
//}
//
//func TestHamtPut(t *testing.T) {
//	var h = hamt64.EMPTY
//	for i := 0; i < numHugeKvs; i++ {
//		h, _ = h.Put(hugeKvs[i].key, i+1)
//	}
//}

func BenchmarkMapGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var j = int(rand.Int31()) % numHugeKvs
		var s = hugeKvs[j].key.String()
		var v = M[s]
		v += 1
	}
}

func BenchmarkHamt64Get(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var j = int(rand.Int31()) % numHugeKvs
		var k = hugeKvs[j].key
		var _, _ = H.Get(k)
	}
}

func BenchmarkMapPut(b *testing.B) {
	var m = make(map[string]int)
	var s = stringutil.Str("aaa")
	for i := 0; i < b.N; i++ {
		m[string(s)] = i + 1
		s = s.DigitalInc(1)
	}
}

func BenchmarkHamt64Put(b *testing.B) {
	var h = hamt64.EMPTY
	var s = stringutil.Str("aaa")
	for i := 0; i < b.N; i++ {
		h, _ = h.Put(string_key.StringKey(s), i+1)
		s = s.DigitalInc(1)
	}
}
