package hamt

import (
	"math/rand"
	"os"
	"testing"

	"github.com/lleo/go-hamt-functional/hamt32"
	"github.com/lleo/go-hamt-functional/hamt64"
	"github.com/lleo/go-hamt/string_key"

	"github.com/lleo/stringutil"
)

var numKvs int
var kvs []keyVal
var randKvs []keyVal

var MAP map[string]int
var H32 hamt32.Hamt
var H64 hamt64.Hamt

func TestMain(m *testing.M) {
	// SETUP

	var str = stringutil.Str("aaa")
	//numKvs = 1024
	numKvs = 1 * 1024 * 1024 // one mega-entries
	//numKvs = 256 * 1024 * 1024 //256 MB
	kvs = make([]keyVal, 0, numKvs)
	for i := 0; i < numKvs; i++ {
		var key = string_key.StringKey(str)
		var val = i + 1
		kvs = append(kvs, keyVal{key, val})
		str = str.DigitalInc(1)
	}

	// For Benchmarks; Build map & hamt
	//randKvs = genRandomizedKvs(kvs)
	MAP = make(map[string]int)
	H32 = hamt32.EMPTY
	H64 = hamt64.EMPTY
	for i := 0; i < len(kvs); i++ {
		MAP[kvs[i].key.String()] = i + 1
		H32, _ = H32.Put(kvs[i].key, kvs[i].val)
		H64, _ = H64.Put(kvs[i].key, kvs[i].val)
	}

	// RUN
	xit := m.Run()

	// TEARDOWN

	os.Exit(xit)
}

//First genRandomizedKvs() copies []keyVal passed in. Then it randomizes that
//copy in-place. Finnally, it returns the randomized copy.
func genRandomizedKvs(kvs []keyVal) []keyVal {
	randKvs := make([]keyVal, len(kvs))
	copy(randKvs, kvs)

	//From: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle#The_modern_algorithm
	for i := len(randKvs) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		randKvs[i], randKvs[j] = randKvs[j], randKvs[i]
	}

	return randKvs
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

func TestHamt64PutDelHugeIsEmpty(t *testing.T) {
	var h = hamt64.EMPTY

	//var k0 = string_key.StringKey("xema")

	for i := 0; i < numKvs; i++ {
		h, _ = h.Put(kvs[i].key, kvs[i].val)
	}

	//log.Println("TEST: h = ", h.LongString(""))

	for i := 0; i < numKvs; i++ {
		var key = kvs[i].key
		var expectedVal = kvs[i].val

		//if key.Equals(k0) {
		//	t.Logf("hit %s", k0)
		//}

		//var h1 hamt64.Hamt
		//var val interface{}
		//var deleted bool
		var h1, val, deleted = h.Del(key)
		if !deleted {
			t.Errorf("Did NOT find&delete for key=\"%s\"", key)
		} else {
			if val != expectedVal {
				t.Errorf("val,%d != expectedVal,%d", val, expectedVal)
			}
		}

		h = h1
	}
	//t.Log("### Testing compressedTable Shrinkage ###")

	if !h.IsEmpty() {
		//log.Println("TestEmptyPutDelTrumpIsEmpty Failed cur !h.IsEmpty()")
		//log.Println(h.LongString(""))
		t.Fatal("NOT h.IsEmpty()")
	}
}

func TestHamt32PutDelHugeIsEmpty(t *testing.T) {
	var h = hamt32.EMPTY

	for i := 0; i < numKvs; i++ {
		h, _ = h.Put(kvs[i].key, kvs[i].val)
	}

	//log.Println("TEST: h = ", h.LongString(""))

	for i := 0; i < numKvs; i++ {
		var key = kvs[i].key
		var expectedVal = kvs[i].val

		//var h1 hamt64.Hamt
		//var val interface{}
		//var deleted bool
		var h1, val, deleted = h.Del(key)
		if !deleted {
			t.Errorf("Did NOT find&delete for key=\"%s\"", key)
		}
		if val != expectedVal {
			t.Errorf("val,%d != expectedVal,%d", val, expectedVal)
		}
		h = h1
	}
	//t.Log("### Testing compressedTable Shrinkage ###")

	if !h.IsEmpty() {
		//log.Println("TestEmptyPutDelTrumpIsEmpty Failed cur !h.IsEmpty()")
		//log.Println(h.LongString(""))
		t.Fatal("NOT h.IsEmpty()")
	}
}

func BenchmarkMapGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var j = int(rand.Int31()) % numKvs
		var s = kvs[j].key.String()
		var v = MAP[s]
		v += 1 //screw golint! it recommends v++ and I hate v++ !
	}
}

func BenchmarkHamt32Get(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var j = int(rand.Int31()) % numKvs
		var k = kvs[j].key
		var _, _ = H32.Get(k)
	}
}

func BenchmarkHamt64Get(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var j = int(rand.Int31()) % numKvs
		var k = kvs[j].key
		var _, _ = H64.Get(k)
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

func BenchmarkHamt32Put(b *testing.B) {
	var h = hamt32.EMPTY
	var s = stringutil.Str("aaa")
	for i := 0; i < b.N; i++ {
		h, _ = h.Put(string_key.StringKey(s), i+1)
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
