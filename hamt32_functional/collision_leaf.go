package hamt32_functional

import (
	"fmt"
	"strings"

	"github.com/lleo/go-hamt/hamt_key"
)

type collisionLeaf struct {
	hash30 uint32 //hash30(key)
	kvs    []keyVal
}

func newCollisionLeaf(hash uint32, kvs []keyVal) *collisionLeaf {
	leaf := new(collisionLeaf)
	leaf.hash30 = hash & mask30
	leaf.kvs = append(leaf.kvs, kvs...)

	return leaf
}

func (l collisionLeaf) hashcode() uint32 {
	return l.hash30
}

//is this needed?
func (l collisionLeaf) copy() *collisionLeaf {
	var nl = new(collisionLeaf)
	nl.hash30 = l.hash30
	nl.kvs = append(nl.kvs, l.kvs...)
	return nl
}

func (l collisionLeaf) String() string {
	var kvstrs = make([]string, len(l.kvs))
	for i := 0; i < len(l.kvs); i++ {
		kvstrs[i] = l.kvs[i].String()
	}
	var jkvstr = strings.Join(kvstrs, ",")

	return fmt.Sprintf("{hash30:%s, kvs:[]kv{%s}}", hash30String(l.hash30), jkvstr)
}

func (l collisionLeaf) get(key hamt_key.Key) (interface{}, bool) {
	for i := 0; i < len(l.kvs); i++ {
		if l.kvs[i].key.Equals(key) {
			return l.kvs[i].val, true
		}
	}
	return nil, false
}

func (l collisionLeaf) put(key hamt_key.Key, val interface{}) (leafI, bool) {
	nl := new(collisionLeaf)
	nl.hash30 = l.hash30
	nl.kvs = append(nl.kvs, l.kvs...)

	for i := 0; i < len(l.kvs); i++ {
		if l.kvs[i].key.Equals(key) {
			nl.kvs[i].val = val
			return nl, false // key,val was not added, merely replaced
		}
	}

	nl.kvs = append(nl.kvs, keyVal{key, val})
	return nl, true // key,val was added
}

func (l collisionLeaf) del(key hamt_key.Key) (leafI, interface{}, bool) {
	if len(l.kvs) == 2 {
		if l.kvs[0].key.Equals(key) {
			return newFlatLeaf(l.hash30, l.kvs[1].key, l.kvs[1].val), l.kvs[0].val, true
		}
		if l.kvs[1].key.Equals(key) {
			return newFlatLeaf(l.hash30, l.kvs[0].key, l.kvs[0].val), l.kvs[1].val, true
		}
		return nil, nil, false
	}

	var nl = l.copy()

	for i := 0; i < len(nl.kvs); i++ {
		if l.kvs[i].key.Equals(key) {
			var retVal = nl.kvs[i].val

			// removing the i'th element of a slice; wiki/SliceTricks "Delete"
			nl.kvs = append(nl.kvs[:i], nl.kvs[i+1:]...)

			return nl, retVal, true
		}
	}

	return nil, nil, false
}

func (l *collisionLeaf) keyVals() []keyVal {
	return l.kvs
}
