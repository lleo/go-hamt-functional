package hamt32_test

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/lleo/go-hamt-functional/hamt32"
	"github.com/lleo/go-hamt/stringkey"
	"github.com/lleo/stringutil"
)

func TestBuildHamt32(t *testing.T) {
	log.Println("TestBuildHamt32:")
	var h = hamt32.Hamt{}

	var added bool
	for _, kv := range KVS {
		h, added = h.Put(kv.Key, kv.Val)
		if !added {
			t.Fatalf("failed to h.Put(%s, %v)", kv.Key, kv.Val)
		}
	}

	//log.Println(h.LongString(""))

	var val interface{}
	var removed bool
	for _, kv := range KVS {
		//log.Printf("kv.Key = %s", kv.Key)
		h, val, removed = h.Del(kv.Key)
		if !removed {
			t.Fatalf("failed to h.Del(%s)", kv.Key)
		}
		if val != kv.Val {
			t.Fatalf("val,%d != kv.Val,%d", val, kv.Val)
		}
	}

	//log.Printf("h = %s", h.LongString(""))

	if !h.IsEmpty() {
		t.Fatalf("!h.IsEmpty()")
	}
}

func BenchmarkHamt32Get(b *testing.B) {
	log.Printf("BenchmarkHamt32Get: b.N=%d", b.N)

	for i := 0; i < b.N; i++ {
		//var j = int(rand.Int31()) % numKvs
		var j = rand.Int() % numKvs
		var key = KVS[j].Key
		var val = KVS[j].Val
		var v, found = TestHamt32.Get(key)
		if !found {
			b.Fatalf("H.Get(%s) not found", key)
		}
		if v != val {
			b.Fatalf("val,%v != KVS[%d].val,%v", v, j, val)
		}
	}
}

func BenchmarkHamt32Put(b *testing.B) {
	log.Printf("BenchmarkHamt32Put: b.N=%d", b.N)

	var h = hamt32.Hamt{}
	var s = "aaa"
	for i := 0; i < b.N; i++ {
		key := stringkey.New(s)
		val := i
		h, _ = h.Put(key, val)
		s = stringutil.DigitalInc(s)
	}
}

func BenchmarkHamt32Del(b *testing.B) {
	log.Printf("BenchmarkHamt32Del: b.N=%d", b.N)

	var h = TestHamt32

	var randomizedKVS = genRandomizedKvs(KVS)

	b.ResetTimer()

	StartTime["run BenchmarkHamt32Del"] = time.Now()
	for i := 0; i < b.N; i++ {
		kv := randomizedKVS[i%numKvs]
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
		b.Fatal("TestHamt32.IsEmpty() => true; hence this wasn't a valid benchmark")
	}

	RunTime["run BenchmarkHamt32Del"] = time.Since(StartTime["run BenchmarkHamt32Del"])
}
