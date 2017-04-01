package hamt64

import (
	"github.com/lleo/go-hamt/key"
)

// nodeI is the interface for every entry in a table; so table entries are
// either a leaf or a table or nil.
//
// The nodeI interface can be for compressedTable, fullTable, flatLeaf, or
// collisionLeaf.
//
// The tableI interface is for compressedTable and fullTable.
//
// The Hash60() method for leaf structs is the 60 most significant bits of
// the keys hash.
//
// The Hash60() method for table structs is the depth*nBits of the hash path
// that leads to the table's position in the Trie.
//
// For leafs Hash60() is the 60 bits returned by a keys Hash60().
// For collisionLeafs this is the definition of what a collision is.
//
type nodeI interface {
	Hash60() uint64
	String() string
}

// Every leafI is a nodeI
type leafI interface {
	nodeI
	get(key key.Key) (interface{}, bool)
	put(key key.Key, val interface{}) (leafI, bool) //bool == added? key/val pair
	del(key key.Key) (leafI, interface{}, bool)     //bool == deleted? key
	keyVals() []key.KeyVal
}

type KeyVals []key.KeyVal

func (kvs KeyVals) contains(k0 key.Key) bool {
	for _, kv := range kvs {
		var k1 = kv.Key
		if k0.Equals(k1) {
			return true
		}
	}
	return false
}

// Every tableI is a nodeI.
//
type tableI interface {
	nodeI

	LongString(indent string, recurse bool) string

	nentries() uint // get the number of nodeI entries

	// Get an Ordered list of index and node pairs. This slice MUST BE Ordered
	// from lowest index to highest.
	entries() []tableEntry

	Get(idx uint) nodeI

	insert(idx uint, entry nodeI) tableI
	replace(idx uint, entry nodeI) tableI
	remove(idx uint) tableI
}

type tableEntry struct {
	idx  uint
	node nodeI
}
