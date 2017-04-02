/*
Package hamt32 implements a functional Hash Array Mapped Trie (HAMT).
It is called hamt32 because this package is using 32 nodes for each level of
the Trie. The term functional is used to imply immutable and persistent.

The 30bits of hash are separated into six 5bit values that constitue the hash
path of any Key in this Trie. However, not all six levels of the Trie are used.
As many levels (six or less) are used to find a unique location
for the leaf to be placed within the Trie.

If all six levels of the Trie are used for two or more key/val pairs then a
special collision leaf will be used to store those key/val pairs,  at the sixth
level of the Trie.
*/
package hamt32

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lleo/go-hamt/key"
	"github.com/pkg/errors"
)

// Nbits constant is the number of bits(5) a 30bit hash value is split into,
// to provied the indexes of a HAMT.
const Nbits uint = 5

// MaxDepth constant is the maximum depth(5) of Nbits values that constitute
// the path in a HAMT, from [0..MaxDepth]for a total of MaxDepth+1(6) levels.
// Nbits*(MaxDepth+1) == HASHBITS (ie 5*(5+1) == 30).
const MaxDepth uint = 5

// TableCapacity constant is the number of table entries in a each node of
// a HAMT datastructure; its value is 1<<Nbits (ie 2^5 == 32).
const TableCapacity uint = 1 << Nbits

func hashPathMask(depth uint) uint32 {
	return uint32(1<<(depth*Nbits)) - 1
}

// Create a string of the form "/%02d/%02d..." to describe a hashPath of
// a given depth.
//
// If you want hashPathString() to include the current idx, you Must
// add one to depth. You may need to do this because you are creating
// a table to be put at the idx'th slot of the current table.
func hashPathString(hashPath uint32, depth uint) string {
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

func h30ToString(h30 uint32) string {
	return hashPathString(h30, MaxDepth)
}

func StringToH30(s string) uint32 {
	if !strings.HasPrefix(s, "/") {
		panic(errors.New("does not start with '/'"))
	}
	var s0 = s[1:]
	var as = strings.Split(s0, "/")

	var h30 uint32 = 0
	for i, s1 := range as {
		var ui, err = strconv.ParseUint(s1, 10, int(Nbits))
		if err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("strconv.ParseUint(%q, %d, %d) failed", s1, 10, Nbits)))
		}
		h30 |= uint32(ui << (uint(i) * Nbits))
		//fmt.Printf("%d: h30 = %q %2d %#02x %05b\n", i, s1, ui, ui, ui)
	}

	return h30
}

//// h30ToBitStr is for printf debugging use
//func h30ToBitStr(h30 uint32) string {
//	var strs = make([]string, MaxDepth+1)
//
//	for depth := uint(0); depth <= MaxDepth; depth++ {
//		var idx = index(h30, depth)
//		strs[MaxDepth-depth] = fmt.Sprintf("%05b", idx)
//	}
//
//	return strings.Join(strs, " ")
//}

//indexMask() generates a Nbits(5-bit) mask for a given depth
func indexMask(depth uint) uint32 {
	return uint32((1<<Nbits)-1) << (depth * Nbits)
}

//index() calculates a Nbits(5-bit) integer based on the hash and depth
func index(h30 uint32, depth uint) uint {
	var idxMask = indexMask(depth)
	var idx = uint((h30 & idxMask) >> (depth * Nbits))
	return idx
}

//buildHashPath(hashPath, idx, depth)
//hashPath will be depth Nbits long
func buildHashPath(hashPath uint32, idx, depth uint) uint32 {
	var mask uint32 = (1 << ((depth - 1) * Nbits)) - 1
	hashPath = hashPath & mask

	return hashPath | uint32(idx<<((depth-1)*Nbits))
}

type keyVal struct {
	key key.Key
	val interface{}
}

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

func (h Hamt) Root() tableI {
	return h.root
}

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
	//var hashPath = k.Hash30() & hashPathMask(depth)
	//var leaf2 = *newFlatLeaf(k, v)

	if FullTableInit {
		return createFullTable(depth, leaf1, leaf2)
	}
	return createCompressedTable(depth, leaf1, leaf2)
}

// copyUp is ONLY called on a fresh copy of the current Hamt. Hence, modifying
// it is allowed.
func (nh *Hamt) persist(oldTable, newTable tableI, path tableStack) {
	if path.isEmpty() {
		nh.root = newTable
		return
	}

	var depth = uint(path.len())
	var parentDepth = depth - 1

	var parentIdx = index(oldTable.Hash30(), parentDepth)

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

	for depth = 0; depth < MaxDepth; depth++ {
		path.push(curTable)
		idx = index(h30, depth)
		curNode = curTable.Get(idx)

		switch n := curNode.(type) {
		case nil:
			return path, nil, idx
		case leafI:
			return path, n, idx
		case tableI:
			curTable = n
			// exit switch then loop for
		default:
			panic(fmt.Sprintf("switch default case: depth=%d; idx=%d; curNode unknown type = %T; value = %v; path=%s", depth, idx, n, n, path))
		}
	}
	if depth == MaxDepth {
		path.push(curTable)
		idx = index(h30, depth)
		curNode = curTable.Get(idx)

		if curNode == nil {
			return path, nil, idx
		} else if leaf, isLeaf := curNode.(leafI); isLeaf {
			return path, leaf, idx
		} else {
			panic(fmt.Sprintf("depth,%d == MaxDepth: %d; idx=%d; unknown type = %T; value = %v; path=%s", depth, MaxDepth, idx, curNode, curNode, path))
		}
	}

	panic("SHOULD NEVER GET HERE!")
}

// Get(k) retrieves the value for a given key from the Hamt. The bool
// represents whether the key was found.
func (h Hamt) Get(k key.Key) (interface{}, bool) {
	var _, leaf, _ = h.find(k)

	//var depth = path.len()
	//var curTable = path.pop()

	if leaf == nil {
		return nil, false
	}

	var val, found = leaf.get(k)
	if !found {
		return nil, false
	}

	return val, true
}

// Put new key/val pair into Hamt, returning a new persistant Hamt and a bool
// indicating if the key/val pair was added(true) or mearly updated(false).
func (h Hamt) Put(k key.Key, v interface{}) (Hamt, bool) {
	var nh Hamt = h //copy by value

	var path, leaf, idx = h.find(k)

	if path == nil { // h.IsEmpty()
		nh.root = createRootTable(newFlatLeaf(k, v))
		nh.nentries++
		return nh, true
	}

	var curTable = path.pop()
	var depth = uint(path.len())

	var newTable tableI
	var added bool

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

	return nh, added
}

// Hamt.Del(k) returns a new Hamt, the value deleted, and a boolean that
// specifies whether or not the key was deleted (eg it didn't exist to start
// with). Therefor you must always test deleted before using the new *Hamt
// value.
func (h Hamt) Del(k key.Key) (Hamt, interface{}, bool) {
	var nh Hamt = h // copy by value

	var path, leaf, idx = h.find(k)

	if path == nil { // h.IsEmpty()
		return nh, nil, false
	}

	var curTable = path.pop()
	//var depth = uint(path.len())

	var newTable tableI
	var val interface{}
	var deleted bool

	if leaf == nil {
		//return nh, val, deleted
		return h, nil, false
	} else {
		var newLeaf leafI
		newLeaf, val, deleted = leaf.del(k)

		if !deleted {
			//return nh, val, deleted
			return h, nil, false
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

	return nh, val, deleted
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
