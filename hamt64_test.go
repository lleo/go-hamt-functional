package hamt_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/lleo/go-hamt-functional/hamt64"
	"github.com/lleo/go-hamt-key/stringkey"
)

func TestPutOne64(t *testing.T) {
	var h = hamt64.Hamt{}
	var k = stringkey.New("aaa")
	var v = 1
	var inserted bool

	h, inserted = h.Put(k, v)
	if !inserted {
		t.Fatalf("failed to h.Put(%q, %v)", k.Str(), v)
	}

	//log.Printf("TestPutOne: h =\n%s", h.LongString(""))
}

// Used progs/findCollisionDepth0.go to find collison at depth 0 for "aaa".
// s = aax; H60 = /52/10/59/43/43/33/17/51/28/35
// s = acb; H60 = /52/60/58/43/43/33/49/50/28/35
func TestCollisionCreatesTable64(t *testing.T) {
	var h = hamt64.Hamt{}
	var k0 = stringkey.New("aax")
	var k1 = stringkey.New("acb")

	var inserted bool
	h, inserted = h.Put(k0, 0)
	if !inserted {
		t.Fatalf("failed to insert %s\n", k0)
	}

	h, inserted = h.Put(k1, 1)
	if !inserted {
		t.Fatalf("failed to insert %s\n", k1)
	}

}

func TestHash60Collision(t *testing.T) {
	//var name = "TestMaxDepthCollision"

	//var s60 = "/00/28/10/00/26/13"
	//var h60 = key.ParseHashVal60(s60)

	var k0 = stringkey.New("ewwd") // val=103647,  H60=/00/28/10/00/26/13
	var v0 = 103647

	var k1 = stringkey.New("fwdyy") // val=3148780, H60=/00/28/10/00/26/13
	var v1 = 3148780

	var h hamt64.Hamt
	var added bool

	h, added = h.Put(k0, v0)
	if !added {
		t.Fatalf("Failed to Put(%s, %d)\n", k0, v0)
	}
	h, added = h.Put(k1, v1)
	if !added {
		t.Fatalf("Failed to Put(%s, %d)\n", k1, v1)
	}

	//log.Printf("%s: h60=%#x h=\n%s", name, h60, h.LongString(""))
}

func TestBuildHamt64(t *testing.T) {
	var name = "TestBuildHamt64:" + CFG
	StartTime[name] = time.Now()

	TestHamt64 = hamt64.Hamt{} // global TestHamt64

	for _, kv := range KVS {
		var k = kv.Key
		var v = kv.Val

		var addded bool
		TestHamt64, addded = TestHamt64.Put(k, v)
		if !addded {
			t.Fatalf("failed to add (probably already exists; bad!) TestHamt64.Put(%s)\n", k)
		}
	}

	if TestHamt64.Nentries() != uint(len(KVS)) {
		t.Fatalf("TestHamt64.entries,%d != len(KVS),%d", TestHamt64.Nentries(), len(KVS))
	}

	RunTime[name] = time.Since(StartTime[name])
}

func TestLookupAll64(t *testing.T) {
	var name = "TestLookupAll:" + CFG

	if TestHamt64.IsEmpty() || TestHamt64.Nentries() != uint(len(KVS)) {
		TestHamt64 = createHamt64(name, KVS, TYP)
	}

	StartTime[name] = time.Now()

	for _, kv := range KVS {
		var k = kv.Key
		var v = kv.Val

		var val, found = TestHamt64.Get(k)
		if !found {
			t.Fatalf("Failed to TestHamt64.Get(k): k=%s\n", k)
		}
		if val != v {
			t.Fatalf("Found value not equal to expected value: v,%d != val,%d\n", k, val)
		}
	}

	RunTime[name] = time.Since(StartTime[name])
	//RunTime[name+"-full"] = time.Since(StartTime[name+"-full"])
}

func TestDeleteAll64(t *testing.T) {
	var name = "TestDeleteAll" + CFG

	if TestHamt64.IsEmpty() || TestHamt64.Nentries() != uint(len(KVS)) {
		TestHamt64 = createHamt64(name, KVS, TYP)
	}

	StartTime[name] = time.Now()
	var h = TestHamt64

	for _, kv := range KVS {
		var k = kv.Key
		var v = kv.Val

		var val interface{}
		var deleted bool
		h, val, deleted = h.Del(k)
		if !deleted {
			t.Fatal("Failed to delete k=%s", k)
		}
		if val != v {
			t.Fatal("For k=%s, returned val,%d != stored v,%d", k, val, v)
		}
	}

	RunTime[name] = time.Since(StartTime[name])
}

func BenchmarkHamt64Get(b *testing.B) {
	var name = fmt.Sprintf("BenchmarkHamt64Get#%d", b.N)
	log.Printf("BenchmarkHamt64Get: b.N=%d", b.N)

	var kvs = buildKeyVals(name, b.N, "aaa", 0)
	var lookupHamt64 = createHamt64(name, kvs, TYP)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var key = kvs[i].Key
		var val = kvs[i].Val
		var v, found = lookupHamt64.Get(key)
		if !found {
			b.Fatalf("H.Get(%s) not found", key)
		}
		if v != val {
			b.Fatalf("val,%v != kvs[%d].val,%v", v, i, val)
		}
	}
}

func BenchmarkHamt64Put(b *testing.B) {
	var name = fmt.Sprintf("BenchmarkHamt64Put#%d", b.N)
	log.Printf("BenchmarkHamt64Put: b.N=%d", b.N)

	var kvs = buildKeyVals(name, b.N, "aaa", 0)

	b.ResetTimer()

	var h = hamt64.Hamt{}
	for i := 0; i < b.N; i++ {
		key := kvs[i].Key
		val := kvs[i].Val

		var added bool
		h, added = h.Put(key, val)
		if !added {
			b.Fatalf("failed to h.Put(%s, %v)", key, val)
		}
	}
}

func BenchmarkHamt64Del(b *testing.B) {
	var name = fmt.Sprintf("BenchmarkHamt64Del#%d", b.N)
	log.Printf("BenchmarkHamt64Del: b.N=%d", b.N)

	var kvs = buildKeyVals(name, b.N+1, "aaa", 0)
	var deleteHamt64 = createHamt64(name, kvs, TYP)

	b.ResetTimer()

	StartTime["run BenchmarkHamt64Del"] = time.Now()

	var h = deleteHamt64

	for i := 0; i < b.N; i++ {
		kv := kvs[i]
		key := kv.Key
		val := kv.Val

		var v interface{}
		var deleted bool
		h, v, deleted = h.Del(key)
		if !deleted {
			b.Fatalf("failed to find and delet key=%s", key)
		}
		if v != val {
			b.Fatalf("deleted key=%s but the value found was wrong v=%d, expected val=%d", key, v, val)
		}
	}

	if h.IsEmpty() {
		b.Fatal("h.IsEmpty() => true; hence this wasn't a valid benchmark")
	}

	RunTime["run BenchmarkHamt64Del"] = time.Since(StartTime["run BenchmarkHamt64Del"])
}
