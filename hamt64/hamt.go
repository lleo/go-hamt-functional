/*
Package hamt64 implements a functional Hash Array Mapped Trie (HAMT).
It is called hamt64 because this package is using 64 nodes for each level of
the Trie. The term functional is used to imply immutable and persistent.

The 60bits of hash are separated into ten 6bit values that constitue the hash
path of any Key in this Trie. However, not all ten levels of the Trie are used.
As many levels (ten or less) are used to find a unique location
for the leaf to be placed within the Trie.

If all ten levels of the Trie are used for two or more key/val pairs then a
special collision leaf will be used to store those key/val pairs,  at the tenth
level of the Trie.
*/
package hamt64

import (
	"fmt"
	"strings"

	"github.com/lleo/go-hamt/key"
)

// Nbits constant is the number of bits(6) a 60bit hash value is split into,
// to provied the indexes of a HAMT.
const Nbits uint = 6

// MaxDepth constant is the maximum depth(9) of Nbits values that constitute
// the path in a HAMT, from [0..MaxDepth]for a total of MaxDepth+1(10) levels.
// Nbits*(MaxDepth+1) == HASHBITS (ie 6*(9+1) == 60).
const MaxDepth uint = 9

// TableCapacity constant is the number of table entries in a each node of
// a HAMT datastructure; its value is 1<<Nbits (ie 2^6 == 64).
const TableCapacity uint = 1 << Nbits

func hashPathMask(depth uint) uint64 {
	return uint64(1<<((depth)*Nbits)) - 1
}

// Create a string of the form "/%02d/%02d..." to describe a hashPath of
// a given depth.
//
// If you want hashPathStrig() to include the current idx, you Must
// add one to depth. You may need to do this because you are creating
// a table to be put at the idx'th slot of the current table.
func hashPathString(hashPath uint64, depth uint) string {
	if depth == 0 {
		return "/"
	}
	var strs = make([]string, depth)

	for d := uint(0); d < depth; d++ {
		var idx = index(hashPath, d)
		strs[d] = fmt.Sprintf("%02d", idx)
	}

	return "/" + strings.Join(strs, "/")
}

func hash60String(h60 uint64) string {
	return hashPathString(h60, MaxDepth)
}

//indexMask() generates a Nbits(6-bit) mask for a given depth
func indexMask(depth uint) uint64 {
	return uint64(uint8(1<<Nbits)-1) << (depth * Nbits)
}

//index() calculates a Nbits(6-bit) integer based on the hash and depth
func index(h60 uint64, depth uint) uint {
	var idxMask = indexMask(depth)
	var idx = uint((h60 & idxMask) >> (depth * Nbits))
	return idx
}

//buildHashPath(hashPath, idx, depth)
func buildHashPath(hashPath uint64, idx, depth uint) uint64 {
	return hashPath | uint64(idx<<(depth*Nbits))
}

type keyVal struct {
	key key.Key
	val interface{}
}

// Configuration contants to be passed to `hamt64.New(int) *Hamt`.
const (
	// HybridTables indicates the structure should use compressedTable
	// initially, then upgrade to fullTable when appropriate.
	HybridTables = iota //0
	// CompTablesOnly indicates the structure should use compressedTables ONLY.
	// This was intended just save space, but also seems to be faster; CPU cache
	// locality maybe?
	CompTablesOnly //1
	// FullTableOnly indicates the structure should use fullTables ONLY.
	// This was intended to be for speed, as compressed tables use a software
	// bitCount function to access individual cells. Turns out, not so much.
	FullTablesOnly //2
)

// TableOptionName is a map of the table option value Hybrid, CompTablesOnly,
// or FullTableOnly to a string representing that option.
//      var options = hamt64.FullTablesOnly
//      hamt64.TableOptionName[hamt64.FullTablesOnly] == "FullTablesOnly"
var TableOptionName = make(map[int]string, 3)

func init() {
	TableOptionName[HybridTables] = "HybridTables"
	TableOptionName[CompTablesOnly] = "CompTablesOnly"
	TableOptionName[FullTablesOnly] = "FullTablesOnly"
}

// UpgradeThreshold is a variable that defines when a compressedTable meats
// or exceeds that number of entries, then that table will be upgraded to
// a fullTable. This only applies when HybridTables option is chosen.
// The current value is TableCapacity/2.
var UpgradeThreshold = TableCapacity / 2

// DowngradeThreshold is a variable that defines when a fullTable becomes
// lower than that number of entries, then that table will be downgraded to
// a compressedTable. This only applies when HybridTables option is chosen.
// The current value is TableCapacity/4.
var DowngradeThreshold = TableCapacity / 4

type Hamt struct {
	root            tableI
	nentries        uint
	grade, fullinit bool
}

// New creates a new hamt64.Hamt data structure with the table option set to
// either:
//
// `hamt64.HybridTables`:
// Initially start out with compressedTable, but when the table is half full
// upgrade to fullTable. If a fullTable shrinks to TableCapacity/8(4) entries
// downgrade to compressedTable.
//
// `hamt64.CompTablesOnly`:
// Use compressedTable ONLY with no up/downgrading to/from fullTable. This
// uses the least amount of space.
//
// `hamt64.FullTablesOnly`:
// Only use fullTable no up/downgrading from/to compressedTables. This is
// the fastest performance.
//
func New(opt int) *Hamt {
	var grade, fullinit bool
	if opt == CompTablesOnly {
		grade = false
		fullinit = false
	} else if opt == FullTablesOnly {
		grade = false
		fullinit = true
	} else /* opt == HybridTables */ {
		grade = true
		fullinit = false
	}
	return &Hamt{nil, 0, grade, fullinit}
}

func (h Hamt) IsEmpty() bool {
	//return h.root == nil
	//return h.nentries == 0
	return h.root == nil && h.nentries == 0
}

func (h Hamt) copy() *Hamt {
	var nh = new(Hamt)

	//nh.root = h.root //this is ok because all tables are immutable
	//nh.nentries = h.nentries
	//nh.grade = h.grade
	//nh.fullinit = h.fullinit
	*nh = h

	return nh
}

func (h Hamt) newRootTable(leaf leafI) tableI {
	if h.fullinit {
		return newRootFullTable(h.grade, leaf)
	}
	return newRootCompressedTable(h.grade, leaf)
}

func (h Hamt) newTable(depth uint, leaf1 leafI, k key.Key, v interface{}) tableI {
	//var hashPath = k.Hash60() & hashPathMask(depth)
	var leaf2 = *newFlatLeaf(k, v)

	if h.fullinit {
		return newFullTable(h.grade, depth, leaf1, leaf2)
	}
	return newCompressedTable(h.grade, depth, leaf1, leaf2)
}

// copyUp is ONLY called on a fresh copy of the current Hamt. Hence, modifying
// it is allowed.
func (h *Hamt) copyUp(oldTable, newTable tableI, path pathT) {
	if path.isEmpty() {
		h.root = newTable
		return
	}

	var depth = uint(len(path))
	var parentDepth = depth - 1

	var oldParent = path.pop()

	var parentIdx = index(oldTable.Hash60(), parentDepth)
	var newParent tableI
	if newTable == nil {
		newParent = oldParent.remove(parentIdx)
	} else {
		newParent = oldParent.replace(parentIdx, newTable)
	}

	h.copyUp(oldParent, newParent, path) //recurses at most MaxDepth-1 times

	return
}

// Get(k) retrieves the value for a given key from the Hamt. The bool
// represents whether the key was found.
func (h Hamt) Get(k key.Key) (interface{}, bool) {
	if h.IsEmpty() {
		return nil, false
	}

	var h60 = k.Hash60()

	var curTable = h.root

	for depth := uint(0); depth <= MaxDepth; depth++ {
		var idx = index(h60, depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			break
		}

		if leaf, ok := curNode.(leafI); ok {
			if leaf.Hash60() == h60 {
				return leaf.get(k)
			}
			return nil, false
		}

		//else curNode MUST BE A tableI
		curTable = curNode.(tableI)
	}
	// curNode == nil || depth > MaxDepth

	return nil, false
}

//var debugKey = stringkey.New("hbud")

// Put new key/val pair into Hamt, returning a new persistant Hamt and a bool
// indicating if the key/val pair was added(true) or mearly updated(false).
func (h Hamt) Put(k key.Key, v interface{}) (Hamt, bool) {
	//var debug = debugKey.Equals(k)

	var nh = h.copy()

	if h.IsEmpty() {
		nh.root = h.newRootTable(newFlatLeaf(k, v))
		nh.nentries++
		return *nh, true
	}

	var newTable tableI
	var added bool

	// for-loop state is path, curTable and depth.
	var path = newPathT()
	var curTable = h.root
	var depth uint

	for depth = 0; depth < MaxDepth; depth++ {
		var idx = index(k.Hash60(), depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			newTable = curTable.insert(idx, newFlatLeaf(k, v))
			added = true
			break
		}

		if curLeaf, isLeaf := curNode.(leafI); isLeaf {
			if curLeaf.Hash60() == k.Hash60() {
				var newLeaf leafI
				newLeaf, added = curLeaf.put(k, v)
				newTable = curTable.replace(idx, newLeaf)
				break
			}

			var tmpTable = h.newTable(depth+1, curLeaf, k, v)
			newTable = curTable.replace(idx, tmpTable)
			added = true
			break
		}

		path.push(curTable)
		curTable = curNode.(tableI)
	}
	if depth == MaxDepth {
		var idx = index(k.Hash60(), depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			newTable = curTable.insert(idx, newFlatLeaf(k, v))
			added = true
		} else {
			var curLeaf = curNode.(leafI)
			var newLeaf leafI
			newLeaf, added = curLeaf.put(k, v)
			newTable = curTable.replace(idx, newLeaf)
		}
	}

	if added {
		nh.nentries++
	}
	nh.copyUp(curTable, newTable, path)

	return *nh, added
}

// Hamt.Del(k) returns a new Hamt, the value deleted, and a boolean that
// specifies whether or not the key was deleted (eg it didn't exist to start
// with). Therefor you must always test deleted before using the new *Hamt
// value.
func (h Hamt) Del(k key.Key) (Hamt, interface{}, bool) {
	var nh = h.copy()

	var val interface{}
	var deleted bool
	var newTable tableI

	// for-loop state is path, curTable, and depth.
	var path = newPathT()
	var curTable = h.root
	var depth uint
	for depth = 0; depth < MaxDepth; depth++ {
		var idx = index(k.Hash60(), depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			//return h, nil, false
			return *nh, val, deleted
		}

		if curLeaf, ok := curNode.(leafI); ok {
			var newLeaf leafI
			newLeaf, val, deleted = curLeaf.del(k)

			if !deleted {
				//return h, nil, false
				return *nh, val, deleted
			}

			if newLeaf == nil {
				newTable = curTable.remove(idx)
			} else {
				newTable = curTable.replace(idx, newLeaf)
			}

			break
		}

		// curNode is NOT nil & NOT a leafI, so curNode MUST BE a tableI.
		// We are going to loop, so update loop state like path and curTable.
		// for-loop will handle updating depth.

		path.push(curTable)
		curTable = curNode.(tableI)
	}
	if depth == MaxDepth {
		var idx = index(k.Hash60(), depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			return *nh, val, deleted
		}

		var curLeaf = curNode.(leafI)

		var newLeaf leafI
		newLeaf, val, deleted = curLeaf.del(k)

		if !deleted {
			//return h, nil, false
			return *nh, val, deleted
		}

		if newLeaf == nil {
			newTable = curTable.remove(idx)
		} else {
			newTable = curTable.replace(idx, newLeaf)
		}
	}

	nh.nentries--
	nh.copyUp(curTable, newTable, path)

	return *nh, val, deleted
}

func (h Hamt) String() string {
	return fmt.Sprintf("Hamt{ nentries: %d, root: %s }", h.nentries, h.root)
}

func (h Hamt) LongString(indent string) string {
	var str string
	if h.root != nil {
		str = indent + fmt.Sprintf("Hamt{ nentries: %d, root:\n", h.nentries)
		str += indent + h.root.LongString(indent, 0)
		str += indent + "}"
		return str
	} else {
		str = indent + fmt.Sprintf("Hamt{ nentries: %d, root: nil }", h.nentries)
	}
	return str
}
