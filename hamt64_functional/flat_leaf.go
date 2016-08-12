package hamt64_functional

import (
	"fmt"

	"github.com/lleo/go-hamt/hamt_key"
)

type flatLeaf struct {
	hash60 uint64 //hash60(key)
	key    hamt_key.Key
	val    interface{}
}

func NewFlatLeaf(h60 uint64, key hamt_key.Key, val interface{}) *flatLeaf {
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
	return NewFlatLeaf(l.hash60, l.key, l.val)
}

func (l flatLeaf) String() string {
	return fmt.Sprintf("flatLeaf{hash60:%s, key:hamt_key.Key(\"%s\"), val:%v}", hash60String(l.hash60), l.key, l.val)
}

func (l flatLeaf) get(key hamt_key.Key) (interface{}, bool) {
	if l.key.Equals(key) {
		return l.val, true
	}
	return nil, false
}

// nentries() is required for tableI
func (l flatLeaf) put(key hamt_key.Key, val interface{}) (leafI, bool) {
	if l.key.Equals(key) {
		h60 := key.Hash60()
		nl := NewFlatLeaf(h60, key, val)
		return nl, true
	}

	var nl = NewCollisionLeaf(l.hash60, []keyVal{keyVal{l.key, l.val}, keyVal{key, val}})
	//var nl = new(collisionLeaf)
	//nl.hash60 = l.hashcode()
	//nl.kvs = append(nl.kvs, keyVal{l.key, l.val})
	//nl.kvs = append(nl.kvs, keyVal{key, val})
	return nl, false //didn't replace
}

func (l flatLeaf) del(key hamt_key.Key) (leafI, interface{}, bool) {
	if l.key.Equals(key) {
		return nil, l.val, true //deleted entry
	}
	return nil, nil, false //didn't delete
}

func (l flatLeaf) keyVals() []keyVal {
	return []keyVal{keyVal{l.key, l.val}}
}
