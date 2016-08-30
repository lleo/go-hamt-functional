package hamt64

import (
	"fmt"

	"github.com/lleo/go-hamt/key"
)

type flatLeaf struct {
	hash60 uint64 //hash60(key)
	key    key.Key
	val    interface{}
}

func newFlatLeaf(h60 uint64, key key.Key, val interface{}) *flatLeaf {
	var fl = new(flatLeaf)
	fl.hash60 = h60
	fl.key = key
	fl.val = val
	return fl
}

// hashcode() is required for nodeI
func (l flatLeaf) hashcode() uint64 {
	return l.hash60
}

// copy() is required for nodeI
func (l flatLeaf) copy() *flatLeaf {
	return newFlatLeaf(l.hash60, l.key, l.val)
}

func (l flatLeaf) String() string {
	return fmt.Sprintf("flatLeaf{hash60:%s, key:key.Key(\"%s\"), val:%v}", hash60String(l.hash60), l.key, l.val)
}

func (l flatLeaf) get(key key.Key) (interface{}, bool) {
	if l.key.Equals(key) {
		return l.val, true
	}
	return nil, false
}

// nentries() is required for tableI
func (l flatLeaf) put(k key.Key, v interface{}) (leafI, bool) {
	if l.key.Equals(k) {
		h60 := key.Hash60(k)
		nl := newFlatLeaf(h60, k, v)
		return nl, true
	}

	var nl = newCollisionLeaf(l.hash60, []keyVal{keyVal{l.key, l.val}, keyVal{k, v}})

	return nl, false //didn't replace
}

func (l flatLeaf) del(key key.Key) (leafI, interface{}, bool) {
	if l.key.Equals(key) {
		return nil, l.val, true //deleted entry
	}
	return nil, nil, false //didn't delete
}

func (l flatLeaf) keyVals() []keyVal {
	return []keyVal{keyVal{l.key, l.val}}
}
