package hamt_test

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/lleo/go-hamt-functional/hamt32"
	"github.com/lleo/go-hamt-key/stringkey"
)

func TestPutOne32(t *testing.T) {
	var h = hamt32.Hamt{}
	var k = stringkey.New("aaa")
	var v = 1
	var inserted bool

	h, inserted = h.Put(k, v)
	if !inserted {
		t.Fatalf("failed to h.Put(%q, %v)", k.Str(), v)
	}

	//log.Printf("TestPutOne: h =\n%s", h.LongString(""))
}

// Used progs/findCollisionDepth0.go to find the first collison at depth 0.
// s = aah; H30 = /08/28/12/24/09/03
// s = aba; H30 = /08/26/11/24/25/01
func TestCollisionCreatesTable32(t *testing.T) {
	var h = hamt32.Hamt{}
	var k0 = stringkey.New("aah")
	var k1 = stringkey.New("aba")

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

func TestHash30Collision(t *testing.T) {
	//var name = "TestMaxDepthCollision"

	//var s30 = "/00/28/10/00/26/13"
	//var h30 = key.ParseHashVal30(s30)

	var k0 = stringkey.New("ewwd") // val=103327,  H30=/00/28/10/00/26/13
	var v0 = 103327

	var k1 = stringkey.New("fwdyy") // val=3148780, H30=/00/28/10/00/26/13
	var v1 = 3148780

	var h hamt32.Hamt
	var added bool

	h, added = h.Put(k0, v0)
	if !added {
		t.Fatalf("Failed to Put(%s, %d)\n", k0, v0)
	}
	h, added = h.Put(k1, v1)
	if !added {
		t.Fatalf("Failed to Put(%s, %d)\n", k1, v1)
	}

	//log.Printf("%s: h30=%#x h=\n%s", name, h30, h.LongString(""))
}

func TestBuildHamt32(t *testing.T) {
	var name = "TestBuildHamt32:" + CFG
	StartTime[name] = time.Now()

	TestHamt32 = hamt32.Hamt{} // global TestHamt32

	for _, kv := range KVS {
		var k = kv.Key
		var v = kv.Val

		var addded bool
		TestHamt32, addded = TestHamt32.Put(k, v)
		if !addded {
			t.Fatalf("failed to add (probably already exists; bad!) TestHamt32.Put(%s)\n", k)
		}
	}

	if TestHamt32.Nentries() != uint(len(KVS)) {
		t.Fatalf("TestHamt32.entries,%d != len(KVS),%d", TestHamt32.Nentries(), len(KVS))
	}

	RunTime[name] = time.Since(StartTime[name])
}

func TestLookupAllHamt32(t *testing.T) {
	var name = "TestLookupAllHamt32:" + CFG

	if TestHamt32.IsEmpty() || TestHamt32.Nentries() != uint(len(KVS)) {
		TestHamt32 = createHamt32(name, KVS, TYP)
	}

	StartTime[name] = time.Now()

	for _, kv := range KVS {
		var k = kv.Key
		var v = kv.Val

		var val, found = TestHamt32.Get(k)
		if !found {
			t.Fatalf("Failed to TestHamt32.Get(k): k=%s\n", k)
		}
		if val != v {
			t.Fatalf("Found value not equal to expected value: v,%d != val,%d\n", k, val)
		}
	}

	RunTime[name] = time.Since(StartTime[name])
	//RunTime[name+"-full"] = time.Since(StartTime[name+"-full"])
}

func TestDeleteAll32(t *testing.T) {
	var name = "TestDeleteAll" + CFG

	if TestHamt32.IsEmpty() || TestHamt32.Nentries() != uint(len(KVS)) {
		TestHamt32 = createHamt32(name, KVS, TYP)
	}

	StartTime[name] = time.Now()
	var h = TestHamt32

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

func BenchmarkHamt32Get(b *testing.B) {
	var name = fmt.Sprintf("BenchmarkHamt32Get#%d", b.N)
	log.Printf("BenchmarkHamt32Get: b.N=%d", b.N)

	var kvs = buildKeyVals(name, b.N, "aaa", 0)
	var lookupHamt32 = createHamt32(name, kvs, TYP)

	//kvs = genRandomizedKvs(kvs)
	b.ResetTimer()

	for _, kv := range kvs {
		var key = kv.Key
		var val = kv.Val
		var v, found = lookupHamt32.Get(key)
		if !found {
			b.Fatalf("H.Get(%s) not found", key)
		}
		if v != val {
			b.Fatalf("v,%v != val,%v for key=%s", v, val, key)
		}
	}
}

func BenchmarkHamt32Put(b *testing.B) {
	var name = fmt.Sprintf("BenchmarkHamt32Put#%d", b.N)
	log.Printf("BenchmarkHamt32Put: b.N=%d", b.N)

	var kvs = buildKeyVals(name, b.N, "aaa", 0)

	b.ResetTimer()

	var h = hamt32.Hamt{}
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

func BenchmarkHamt32Del(b *testing.B) {
	var name = fmt.Sprintf("BenchmarkHamt32Del#%d", b.N)
	log.Printf("BenchmarkHamt32Del: b.N=%d", b.N)

	var kvs = buildKeyVals(name, b.N+1, "aaa", 0)
	var deleteHamt32 = createHamt32(name, kvs, TYP)

	b.ResetTimer()

	StartTime["run BenchmarkHamt32Del"] = time.Now()

	var h = deleteHamt32

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

	RunTime["run BenchmarkHamt32Del"] = time.Since(StartTime["run BenchmarkHamt32Del"])
}
