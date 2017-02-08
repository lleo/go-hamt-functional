package hamt64

import (
	"fmt"
	"strings"

	"github.com/lleo/go-hamt/key"
)

type collisionLeaf struct {
	kvs []key.KeyVal
}

func newCollisionLeaf(kvs []key.KeyVal) *collisionLeaf {
	leaf := new(collisionLeaf)
	leaf.kvs = append(leaf.kvs, kvs...)

	return leaf
}

func (l collisionLeaf) Hash60() uint64 {
	return l.kvs[0].Key.Hash60()
}

func (l collisionLeaf) String() string {
	var kvstrs = make([]string, len(l.kvs))
	for i := 0; i < len(l.kvs); i++ {
		kvstrs[i] = l.kvs[i].String()
	}
	var jkvstr = strings.Join(kvstrs, ",")

	return fmt.Sprintf("{kvs:[]kv{%s}}", jkvstr)
}

func (l collisionLeaf) get(key key.Key) (interface{}, bool) {
	for i := 0; i < len(l.kvs); i++ {
		if l.kvs[i].Key.Equals(key) {
			return l.kvs[i].Val, true
		}
	}
	return nil, false
}

func (l collisionLeaf) copy() *collisionLeaf {
	var nl = new(collisionLeaf)

	// keep key.KeyVal containers, only this splice is new
	nl.kvs = append(nl.kvs, l.kvs...)

	return nl
}

// put insertes a new key,val pair into the leaf node, and returns a new leaf
// and a bool representing if the new leaf is bigger (ie accumulated key/val pair).
func (l collisionLeaf) put(k key.Key, val interface{}) (leafI, bool) {
	var nl = l.copy()

	// check if k is exact match of current key
	// if exact match create new key.KeyVal container and update Val
	// and return new leaf & bool
	for i := 0; i < len(l.kvs); i++ {
		if l.kvs[i].Key.Equals(k) { // Key.Equal() checks equal-by-value

			// new key.KeyVal container, and keep the old l.kvs[i].Key object.
			nl.kvs[i] = key.KeyVal{k, val}

			return nl, false // key,val was not added, merely replaced Val
		}
	}

	nl.kvs = append(nl.kvs, key.KeyVal{k, val})
	return nl, true // k,val was added
}

// del method searches current list of key.KeyVal objects, if k found
// remove matching key.KeyVal container, and return a new leafI, the removed
// value, and a bool indicating if the k was found&removed.
func (l collisionLeaf) del(k key.Key) (leafI, interface{}, bool) {

	if len(l.kvs) == 2 {
		// exhaustive search
		// if k found new leaf will be a flatLeaf.
		if l.kvs[0].Key.Equals(k) {
			return newFlatLeaf(l.kvs[1].Key, l.kvs[1].Val), l.kvs[0].Val, true
		}
		if l.kvs[1].Key.Equals(k) {
			return newFlatLeaf(l.kvs[0].Key, l.kvs[0].Val), l.kvs[1].Val, true
		}

		// k not found, hence no deletion occured
		return nil, nil, false
	}

	var nl = l.copy()

	for i := 0; i < len(l.kvs); i++ {
		if l.kvs[i].Key.Equals(k) {
			var retVal = l.kvs[i].Val

			// removing the i'th element of a slice; wiki/SliceTricks "Delete"
			nl.kvs = append(nl.kvs[:i], nl.kvs[i+1:]...)

			return nl, retVal, true
		}
	}

	return nil, nil, false
}

func (l collisionLeaf) keyVals() []key.KeyVal {
	return l.kvs
}
