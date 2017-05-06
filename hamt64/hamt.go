/*
Package hamt64 implements a functional Hash Array Mapped Trie (HAMT).
It is called hamt64 because this package is using 64 nodes for each level of
the Trie. The term functional is used to imply immutable and persistent.

The key to the hamt64 datastructure is imported from the
"github.com/lleogo-hamt-key" module. We get the 60 bits of hash value from key.
The 60bits of hash are separated into ten 6 bit values that constitue the hash
path of any Key in this Trie. However, not all ten levels of the Trie are used.
As many levels (ten or less) are used to find a unique location
for the leaf to be placed within the Trie.

If all ten levels of the Trie are used for two or more key/val pairs, then a
special collision leaf will be used to store those key/val pairs at the tenth
level of the Trie.
*/
package hamt64

import (
	"fmt"
	"log"

	"github.com/lleo/go-hamt-key"
)

// Nbits constant is the number of bits(6) a 60bit hash value is split into,
// to provied the indexes of a HAMT. We actually get this value from
// key.BitsPerLevel60 in "github.com/lleo/go-hamt-key".
//const Nbits uint = 6
const Nbits uint = key.BitsPerLevel60

// MaxDepth constant is the maximum depth(6) of Nbits values that constitute
// the path in a HAMT, from [0..MaxDepth] for a total of MaxDepth+1(9) levels.
// Nbits*(MaxDepth+1) == HASHBITS (ie 6*(6+1) == 60). We actually get this
// value from key.MaxDepth60 in "github.com/lleo/go-hamt-key".
//const MaxDepth uint = 6
const MaxDepth uint = key.MaxDepth60

// TableCapacity constant is the number of table entries in a each node of
// a HAMT datastructure; its value is 1<<Nbits (ie 2^6 == 64).
//const TableCapacity uint = 1 << Nbits
const TableCapacity uint = 1 << key.BitsPerLevel60

// GradeTables variable controls whether Hamt structures will upgrade/
// downgrade compressed/full tables. This variable and FullTableInit
// should not be changed during the lifetime of any Hamt structure.
// Default: true
var GradeTables = true

// FullTableInit variable controls whether the initial new table type is
// fullTable, else the initial new table type is compressedTable.
// Default: false
var FullTableInit = false

// UpgradeThreshold is a variable that defines when a compressedTable meats
// or exceeds that number of entries, then that table will be upgraded to
// a fullTable. This only applies when HybridTables option is chosen.
// The current value is TableCapacity/2.
var UpgradeThreshold = TableCapacity * 2 / 3

// DowngradeThreshold is a variable that defines when a fullTable becomes
// lower than that number of entries, then that table will be downgraded to
// a compressedTable. This only applies when HybridTables option is chosen.
// The current value is TableCapacity/4.
var DowngradeThreshold = TableCapacity / 4

type Hamt struct {
	root     tableI
	nentries uint
}

func (h Hamt) IsEmpty() bool {
	//return h.root == nil
	//return h.nentries == 0
	//return h.root == nil && h.nentries == 0
	return h == Hamt{}
}

//func (h Hamt) Root() tableI {
//	return h.root
//}

func (h Hamt) Nentries() uint {
	return h.nentries
}

func createRootTable(leaf leafI) tableI {
	if FullTableInit {
		return createRootFullTable(leaf)
	}
	return createRootCompressedTable(leaf)
}

//func createTable(depth uint, leaf1 leafI, k key.Key, v interface{}) tableI {
func createTable(depth uint, leaf1 leafI, leaf2 flatLeaf) tableI {
	if FullTableInit {
		return createFullTable(depth, leaf1, leaf2)
	}
	return createCompressedTable(depth, leaf1, leaf2)
}

// persist() is ONLY called on a fresh copy of the current Hamt.
// Hence, modifying it is allowed.
func (nh *Hamt) persist(oldTable, newTable tableI, path tableStack) {
	if path.isEmpty() {
		nh.root = newTable
		return
	}

	var depth = uint(path.len())
	var parentDepth = depth - 1

	var parentIdx = oldTable.Hash60().Index(parentDepth)

	var oldParent = path.pop()
	var newParent tableI

	if newTable == nil {
		newParent = oldParent.remove(parentIdx)
	} else {
		newParent = oldParent.replace(parentIdx, newTable)
	}

	nh.persist(oldParent, newParent, path) //recurses at most MaxDepth-1 times

	return
}

func (h Hamt) find(k key.Key) (path tableStack, leaf leafI, idx uint) {
	if h.IsEmpty() {
		return nil, nil, 0
	}

	path = newTableStack()
	var curTable = h.root

	var h60 = k.Hash60()
	var depth uint
	var curNode nodeI

DepthIter:
	for depth = 0; depth <= MaxDepth; depth++ {
		path.push(curTable)
		idx = h60.Index(depth)
		curNode = curTable.get(idx)

		switch n := curNode.(type) {
		case nil:
			leaf = nil
			break DepthIter
		case leafI:
			leaf = n
			break DepthIter
		case tableI:
			if depth == MaxDepth {
				log.Panicf("SHOULD NOT BE REACHED; depth,%d == MaxDepth,%d & tableI entry found; %s", depth, MaxDepth, n)
			}
			curTable = n
			// exit switch then loop for
		default:
			log.Panicf("SHOULD NOT BE REACHED: depth=%d; curNode unknown type=%T;", depth, curNode)
		}
	}

	return
}

// Get(k) retrieves the value for a given key from the Hamt. The bool
// represents whether the key was found.
//func (h Hamt) Get(k key.Key) (val interface{}, found bool) {
//	var _, leaf, _ = h.find(k)
//
//	if leaf == nil {
//		//return nil, false
//		return
//	}
//
//	val, found = leaf.get(k)
//	return
//}

// Get(k) retrieves the value for a given key from the Hamt. The bool
// represents whether the key was found.
func (h Hamt) Get(k key.Key) (val interface{}, found bool) {
	if h.IsEmpty() {
		return //nil, false
	}

	var h30 = k.Hash30()

	var curTable = h.root

	for depth := uint(0); depth <= MaxDepth; depth++ {
		var idx = h30.Index(depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			return //nil, false
		}

		if leaf, isLeaf := curNode.(leafI); isLeaf {
			val, found = leaf.get(k)
			return
		}

		if depth == MaxDepth {
			panic("SHOULD NOT HAPPEN")
		}
		curTable = curNode.(tableI)
	}

	panic("SHOULD NEVER BE REACHED")
}

// Put inserts a key/val pair into Hamt, returning a new persistent Hamt and a
// bool indicating if the key/val pair was added(true) or mearly updated(false).
func (h Hamt) Put(k key.Key, v interface{}) (nh Hamt, added bool) {
	nh = h //copy by value

	var path, leaf, idx = h.find(k)

	if path == nil { // h.IsEmpty()
		nh.root = createRootTable(newFlatLeaf(k, v))
		nh.nentries++

		//return nh, true
		added = true
		return
	}

	var curTable = path.pop()
	var depth = uint(path.len())

	var newTable tableI

	if leaf == nil {
		newTable = curTable.insert(idx, newFlatLeaf(k, v))
		added = true
	} else {
		if leaf.Hash60() == k.Hash60() {
			var newLeaf leafI
			newLeaf, added = leaf.put(k, v)
			newTable = curTable.replace(idx, newLeaf)
		} else {
			var tmpTable = createTable(depth+1, leaf, *newFlatLeaf(k, v))
			newTable = curTable.replace(idx, tmpTable)
			added = true
		}
	}

	if added {
		nh.nentries++
	}

	nh.persist(curTable, newTable, path)

	//return nh, added
	return
}

// Hamt.Del(k) returns a Hamt structure, a value, and a boolean that specifies
// whether or not the key was found (and therefor deleted). If the key was
// found & deleted it returns the value assosiated with the key and a new
// persistent Hamt structure, otherwise it returns a nil value and the original
// (immutable) Hamt structure
func (h Hamt) Del(k key.Key) (nh Hamt, val interface{}, deleted bool) {
	nh = h // copy by value

	var path, leaf, idx = h.find(k)

	if path == nil { // h.IsEmpty()
		//return nh, nil, false
		return
	}

	var curTable = path.pop()
	//var depth = uint(path.len())

	var newTable tableI

	if leaf == nil {
		//return nh, val, found
		//return h, nil, false
		return
	} else {
		var newLeaf leafI
		newLeaf, val, deleted = leaf.del(k)

		if !deleted {
			//return nh, val, deleted
			//return h, nil, false
			return
		}

		if newLeaf == nil {
			newTable = curTable.remove(idx)
		} else {
			newTable = curTable.replace(idx, newLeaf)
		}
	}

	if deleted {
		nh.nentries--
	}

	nh.persist(curTable, newTable, path)

	//return nh, val, deleted
	return
}

func (h Hamt) String() string {
	return fmt.Sprintf("Hamt{ nentries: %d, root: %s }", h.nentries, h.root)
}

const halfIndent = "  "
const fullIndent = "    "

func (h Hamt) LongString(indent string) string {
	var str string
	if h.root != nil {
		str = indent + fmt.Sprintf("Hamt{ nentries: %d, root:\n", h.nentries)
		str += indent + h.root.LongString(indent+fullIndent, true)
		str += indent + "}end\n"
		return str
	} else {
		str = indent + fmt.Sprintf("Hamt{ nentries: %d, root: nil }", h.nentries)
	}
	return str
}
