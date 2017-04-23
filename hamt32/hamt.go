/*
Package hamt32 implements a functional Hash Array Mapped Trie (HAMT).
It is called hamt32 because this package is using 32 nodes branching factor for
each level of the Trie. The term functional is used to imply immutable and
persistent.

The key to the hamt32 datastructure is imported from the
"github.com/lleogo-hamt-key" module. We get the 30 bits of hash value from key.
The 30bits of hash are separated into six 5 bit values that constitue the hash
path of any Key in this Trie. However, not all six levels of the Trie are used.
As many levels (six or less) are used to find a unique location
for the leaf to be placed within the Trie.

If all six levels of the Trie are used for two or more key/val pairs, then a
special collision leaf will be used to store those key/val pairs at the sixth
level of the Trie.
*/
package hamt32

import (
	"fmt"
	"log"

	"github.com/lleo/go-hamt-key"
)

// Nbits constant is the number of bits(5) a 30bit hash value is split into,
// to provied the indexes of a HAMT. We actually get this value from
// key.BitsPerLevel30 in "github.com/lleo/go-hamt-key".
//const Nbits uint = 5
const Nbits uint = key.BitsPerLevel30

// MaxDepth constant is the maximum depth(5) of Nbits values that constitute
// the path in a HAMT, from [0..MaxDepth] for a total of MaxDepth+1(6) levels.
// Nbits*(MaxDepth+1) == HASHBITS (ie 5*(5+1) == 30). We actually get this
// value from key.MaxDepth60 in "github.com/lleo/go-hamt-key".
//const MaxDepth uint = 5
const MaxDepth uint = key.MaxDepth30

// TableCapacity constant is the number of table entries in a each node of
// a HAMT datastructure; its value is 1<<Nbits (ie 2^5 == 32).
//const TableCapacity uint = 1 << Nbits
const TableCapacity uint = 1 << key.BitsPerLevel30

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

	var parentIdx = oldTable.Hash30().Index(parentDepth)

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

	var h30 = k.Hash30()
	var depth uint
	var curNode nodeI

DepthIter:
	for depth = 0; depth <= MaxDepth; depth++ {
		path.push(curTable)
		idx = h30.Index(depth)
		curNode = curTable.Get(idx)

		switch n := curNode.(type) {
		case nil:
			leaf = nil
			break DepthIter
		case leafI:
			leaf = n
			break DepthIter
		case tableI:
			if depth == MaxDepth {
				log.Printf("k = %s", k)
				log.Printf("path=%s", path)
				log.Printf("curTable=%s", curTable.LongString("", false))
				log.Printf("idx=%s", idx)
				log.Printf("curNode type=%T; value=%v", curNode, curNode)
				log.Panicf("SHOULD NOT BE REACHED; depth,%d == MaxDepth,%d & invalid type=%T;", depth, MaxDepth, curNode)
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
func (h Hamt) Get(k key.Key) (val interface{}, found bool) {
	var _, leaf, _ = h.find(k)

	if leaf == nil {
		//return nil, false
		return
	}

	val, found = leaf.get(k)
	return
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
		if leaf.Hash30() == k.Hash30() {
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
