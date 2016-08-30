package hamt32

import (
	"fmt"

	"github.com/lleo/go-hamt/key"
)

type flatLeaf struct {
	hash30 uint32 //hash30(key)
	key    key.Key
	val    interface{}
}

func newFlatLeaf(h30 uint32, key key.Key, val interface{}) *flatLeaf {
	var fl = new(flatLeaf)
	fl.hash30 = h30
	fl.key = key
	fl.val = val
	return fl
}

// hashcode() is required for nodeI
func (l flatLeaf) hashcode() uint32 {
	return l.hash30
}

// copy() is required for nodeI
func (l flatLeaf) copy() *flatLeaf {
	return newFlatLeaf(l.hash30, l.key, l.val)
}

func (l flatLeaf) String() string {
	return fmt.Sprintf("flatLeaf{hash30:%s, key:key.Key(\"%s\"), val:%v}", hash30String(l.hash30), l.key, l.val)
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
		h30 := key.Hash30(k)
		nl := newFlatLeaf(h30, k, v)
		return nl, true
	}

	var nl = newCollisionLeaf(l.hash30, []keyVal{keyVal{l.key, l.val}, keyVal{k, v}})

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
