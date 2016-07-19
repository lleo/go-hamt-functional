package hamt

import "fmt"

type flatLeaf struct {
	hash60 uint64 //hash60(key)
	key    []byte
	val    interface{}
}

func NewFlatLeaf(h60 uint64, key []byte, val interface{}) *flatLeaf {
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
	return fmt.Sprintf("flatLeaf{hash60:%s, key:[]byte(\"%s\"), val:%v}", hash60String(l.hash60), l.key, l.val)
}

func (l flatLeaf) get(key []byte) (interface{}, bool) {
	if byteSlicesEqual(l.key, key) {
		return l.val, true
	}
	return nil, false
}

// nentries() is required for tableI
func (l flatLeaf) put(key []byte, val interface{}) (leafI, bool) {
	if byteSlicesEqual(l.key, key) {
		h60 := hash60(key)
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

func (l flatLeaf) del(key []byte) (leafI, interface{}, bool) {
	if byteSlicesEqual(l.key, key) {
		return nil, l.val, true //deleted entry
	}
	return nil, nil, false //didn't delete
}
