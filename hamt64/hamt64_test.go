package hamt64_test

import (
	"log"
	"testing"

	"github.com/lleo/go-hamt-functional/hamt64"
)

func TestHamt64NewFullTablesOnly(t *testing.T) {
	var h = hamt64.New(hamt64.FullTablesOnly)
	if h == nil {
		t.Fatal("no new Hamt struct")
	}
}

func TestHamt64NewCompTablesOnly(t *testing.T) {
	var h = hamt64.New(hamt64.CompTablesOnly)
	if h == nil {
		t.Fatal("no new Hamt struct")
	}
}

func TestHamt64NewHybridTables(t *testing.T) {
	var h = hamt64.New(hamt64.HybridTables)
	if h == nil {
		t.Fatal("no new Hamt struct")
	}
}

func TestBuildHamt64(t *testing.T) {
	log.Println("TestBuildHamt64:")
	var h = hamt64.New(TableOption)

	var added bool
	for _, kv := range hugeKvs {
		*h, added = h.Put(kv.Key, kv.Val)
		if !added {
			t.Fatalf("failed to h.Put(%s, %v)", kv.Key, kv.Val)
		}
	}
	//log.Println(h.LongString(""))

	var val interface{}
	var removed bool
	for _, kv := range hugeKvs {
		*h, val, removed = h.Del(kv.Key)
		if !removed {
			t.Fatalf("failed to h.Del(%s)", kv.Key)
		}
		if val != kv.Val {
			t.Fatalf("val,%d != kv.Val,%d", val, kv.Val)
		}
	}

	log.Printf("h = %s", h.LongString(""))

	if !h.IsEmpty() {
		t.Fatalf("!h.IsEmpty()")
	}
}
