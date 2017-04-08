package hamt32

import (
	"fmt"

	"github.com/lleo/go-hamt-key"
)

type flatLeaf struct {
	key key.Key
	val interface{}
}

func newFlatLeaf(key key.Key, val interface{}) *flatLeaf {
	var fl = new(flatLeaf)
	fl.key = key
	fl.val = val
	return fl
	//return &flatLeaf{key, val}
}

func (l flatLeaf) Key() key.Key {
	return l.key
}

func (l flatLeaf) Val() interface{} {
	return l.val
}

// Hash30() is required for nodeI
func (l flatLeaf) Hash30() key.HashVal30 {
	return l.key.Hash30()
}

// copy() is required for nodeI
func (l flatLeaf) copy() *flatLeaf {
	return newFlatLeaf(l.key, l.val)
}

func (l flatLeaf) String() string {
	return fmt.Sprintf("flatLeaf{key:key.Key(\"%s\"), val:%v}", l.key, l.val)
}

func (l flatLeaf) get(key key.Key) (interface{}, bool) {
	if l.key.Equals(key) {
		return l.val, true
	}
	return nil, false
}

// put inserts a new key/val pair. Returns new leaf node and a bool indicating if
// the key/val pair was added?(true), or was a previous key/val pair updated?(false).
func (l flatLeaf) put(k key.Key, v interface{}) (leafI, bool) {
	if l.key.Equals(k) {
		nl := newFlatLeaf(k, v)
		return nl, false // did NOT add k/v pair
	}

	var nl = newCollisionLeaf([]key.KeyVal{key.KeyVal{l.key, l.val}, key.KeyVal{k, v}})

	return nl, true // added k,v pair
}

func (l flatLeaf) del(key key.Key) (leafI, interface{}, bool) {
	if l.key.Equals(key) {
		return nil, l.val, true //deleted entry
	}
	return nil, nil, false //didn't delete
}

func (l flatLeaf) keyVals() []key.KeyVal {
	return []key.KeyVal{key.KeyVal{l.key, l.val}}
}
