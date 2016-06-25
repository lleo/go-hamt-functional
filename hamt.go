/*
Package hamt implements a functional Hash Array Mapped Trie (HAMT). Functional
is defined as immutable and persistent. FIXME more explanation.

A Hash Array Mapped Trie (HAMT) is a data structure to map a key (a byte slice)
to value (a interface{}; for a generic item). To do this we use a Trie data-
structure; where each node has a table of SIZE sub-nodes; where each sub-node is
a new table or a leaf; leaf's contain the key/value pairs.

We take a byte slice key and hash it into a 64 bit value. We split the hash value
into ten 6-bit values. Each 6-bit values is used as an index into a 64 entry
table for each node in the Trie structure. For each index, if there is no
collision (ie. no previous entry) in that Trie nodes table's position, then we
put a leaf (aka a value entry); if there is a collision, then we put a new node
table of the Trie structure and use the next 6-bits of the hash value to
calculate the index into that table. This means that the Trie is at most ten
levels deap, AND only as deep as is needed; for savings in memory and access
time. Algorithmically, this allows for a O(1) hash table.

We use go's "hash/fnv" FNV1 implementation for the hash.

Typically HAMT's can be implemented in 64/6 bit and 32/5 bit versions. I've
implemented this as a 64/6 bit version.
*/
package hamt

import (
	"hash/fnv"
)

// The number of bits to partition the hashcode and to index each table. By
// logical necessity this MUST be 6 bits because 2^6 == 64; the number of
// entries in a table.
const NBITS uint = 6

// The Capacity of a table; 2^6 == 64;
const TABLE uint = 1 << NBITS

const sixtyBitMask = 1<<60 - 1

// The maximum depthof a HAMT ranges between 0 and 9, for 10 levels
// total and cunsumes 60 of the 64 bit hashcode.
const MAXDEPTH uint = 9

const ASSERT_CONST bool = treu

func ASSERT(test bool, msg string) {
	if ASSERT_CONST {
		if test {
			panic(string)
		}
	}
}

// This function calcuates a 64-bit unsigned integer hashcode from an
// arbitrary list of bytes. It uses the FNV1 hashing algorithm:
//   https://golang.org/pkg/hash/fnv/
//   https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function
//
func hash64(bs []byte) uint64 {
	var h = fnv.New64()
	h.Write(bs)
	return h.Sum64()
}

func hashPathEqual(depth uint, a, b uint64) bool {
	pathMask := 1<<(depth*NBITS) - 1

	return (a & pathMask) == (b & pathMask)
}

func makeHashPath(depth uint, hash uint64) uint64 {
	return hash & hashPathMask(depth)
}

func hashPathMask(depth) uint64 {
	return 1<<(depth*NBITS) - 1
}

// The hash60Equal function compares the significant 60 bits of hash that
// HAMT uses for navagating the Trie.
//
func hash60Equal(a, b uint64) bool {
	return (a & sixtyBitMask) == (b & sixtyBitMask)
}

//idxMask() generates a NBITS(6-bit) mask for a given depth
func idxMask(depth) uint64 {
	var m = uint64(uint8(1<<NBITS) - 1)
	return m << (depth * NBITS)
}

//index() calculates a NBITS(6-bit) integer based on the hash and depth
func index(hash uint64, depth uint) uint {
	var m = idxMask(depth)
	idx = (hash & m) >> (depth * NBITS)
	return idx
}

type Hamt struct {
	root     tableI
	nentries uint
}

var EMPTY = Hamt{nil}

func (h Hamt) IsEmpty() bool {
	return h.root == nil
}

func (h Hamt) copyUp(oldTable, newTable tableI, path pathT) Hamt {
	if path.isEmpty() {
		if h.root != oldTable {
			panic("WTF! path.isEmpty() and h.root != oldTable")
		}
		return Hamt{root: newTable}
	}

	oldParent := path.pop()

	//newParent := oldParent.replace(oldNode, newNode)
	newParent := oldParent.set(oldNode.hashcode(), newNode)

	return h.copyUp(oldParent, newParent, path)
}

func (h Hamt) copyUpLeaf(oldLeaf, newLeaf, path) Hamt {
	if path.isEmpty() {
		panic("copyUpLeaf() call where path.isEmpty()")
	}

	var oldTable = path.pop()
	var newTable = oldTable.set(hash, newLeaf)

	return h.copyUp(oldTable, newTable, path)
}

func (h Hamt) Get(key []byte) (interface{}, bool) {
	if h.IsEmpty() {
		return nil, false
	}

	var hash = hash64(key)

	// We know h.root != nil (above IsEmpty test) and h.root is a tableI
	// intrface compliant struct.
	var curTable = h.root
	var curNode = curTable.get(hash)

	for depth := 0; curNode != nil && depth <= MAXDEPTH; depth++ {
		//if curNode ISA leafI
		if leaf, ok := curNode.(leafI); ok {
			if hashPathEqual(depth, hash, leaf.hashcode()) {
				return leaf.get(key)
			}
			return nil, false
		}

		//else curNode MUST BE A tableI
		var curTable = curNode.(tableI)

		curNode = curTable.get(key, hash)
	}

	// curNode == nil
	return nil, false
}

func (h Hamt) Put(key []byte, val interface{}) (newHamt Hamt, newVal bool) {
	newHamt = Hamt{nil} //always creating a new Hamt
	newVal = true       //true == inserted key/val pair; false == replaced val

	//var depth uint = 0
	var hash = hash64(key)

	/**********************************************************
	 * If the Trie root is empty, insert a table with a leaf. *
	 **********************************************************/
	if h.IsEmpty() {
		leaf := NewFlatLeaf(hash, key, val)
		newHamt.root = NewCompressedTable(0, hash, leaf)
		return
	}

	var path = newPathT()
	var curTable = h.root
	var depth uint
	var hashPath uint64 = 0
	var newLeaf = NewFlatLeaf(hash, key, val)

	for depth = 0; depth < MAXDEPTH; depth++ {

		curNode := curTable.get(hash, key)

		/***********************************************
		 * If the table entry is empty, insert a leaf. *
		 ***********************************************/
		if curNode == nil {
			newTable = curTable.set(hash, newLeaf)
			h.copyUp(curTable, newTable, path)
			return
		}

		/**********************************
		 * If the table entry ISA leaf... *
		 **********************************/
		if oldLeaf, ok := curNode.(leafI); ok {

			/**************************************************************
			 * ...AND hashcode COLLISION. Either we are part of
			 * the way down to MAXDEPTH and part of the hashcode matches,
			 * or we are at MAXDEPTH and 60 of 64 bits match.
			 **************************************************************/
			if hash60Equal(oldLeaf.hashcode(), hash) {
				//if oldLeaf is a flatLeaf this will promote it to a
				// collisionLeaf, else it already is a collisionLeaf.
				path.push(curTable)
				newLeaf, newVal = oldLeaf.put(key, val)
				newHamt.copyUpLeaf(oldLeaf, newLeaf, path)
				return
			}

			var newLeaf = NewFlatLeaf(hash, key, val)
			curTable = NewCompressedTable(depth, hash, oldLeaf, newLeaf)

			h.copyUp(curTable, newTable, path)
			return
		}

		path.push(curTable)

		/***********************************
		 * The tale entry MUST BE A tableI *
		 ***********************************/
		curTable = curNode.(tableI)

	} //end: for

	return newHamt, false
}

func (h Hamt) Del(key []byte) (Hamt, interface{}, bool) {
	return Hamt{nil}, nil, false
}

// The nodeI interface can be compressedTable, fullTable, flatLeaf, or
// collisionLeaf
//
// For table structs hashcode() is the depth*NBITS of the hash path that leads
// to thistable's position in the Trie.
//
// For leafs hashcode() is the top 60 bits out of the hash64(key).
// For collisionLeafs this is the definition of what a collision is.
//
type nodeI interface {
	hashcode() uint64
}

type tableI interface {
	nodeI
	copy() tableI
	get(hash uint64) nodeI
	set(hash uint64, entry nodeI) tableI
	clear(hash uint64) (tableI, nodeI)
	nentries() uint
}

type compressedTable struct {
	hashPath uint64 // depth*NBITS of hash to get to this location in the Trie
	depth    uint
	nodeMap  uint64
	nodes    []nodeI
}

// Utility structure only used for NewCompressedTable() constructor.
//
type tableEntry struct {
	idx  uint
	hash uint64
	key  []byte
	val  interface{}
}

type tableS []tableEntry

func NewCompressedTable(depth uint, hashPath uint64, lf leafI) tableI {
	ASSERT(depth > -1 && depth < MAXDEPTH+1, "parameter 0 >= depth >= 9")

	var idx = index(lf.hashcode(), depth)

	var ct = new(compressedTable)
	ct.hashPath = hashPath
	ct.depth = depth
	ct.nodeMap = 1 << idx
	ct.nodes = make([]nodeI, 1)
	ct.nodes[0] = lf

	return ct
}

func NewCompressedTable2(depth uint, hashPath uint64, lf1, lf2 leafI) tableI {
	ASSERT(depth > -1 && depth < MAXDEPTH+1, "parameter 0 >= depth >= 9")

	var retTable = new(compressedTable) // return compressedTable
	retTable.hashPath = hashPath
	retTable.depth = depth

	var curTable = retTable //*compressedTable
	var d uint
	for d = depth; d < MAXDEPTH; d++ {
		var idx1 = index(lf1.hashcode(), d)
		var idx2 = index(lf2.hashcode(), d)
		curTable.nodes = make([]nodeI, 2)

		if idx1 != idx2 {
			curTable.nodeMap &= 1 << idx1
			curTable.nodeMap &= 1 << idx2
			if idx1 < idx2 {
				curTable.nodes[0] = lf1
				curTable.nodes[1] = lf2
			} else {
				curTable.nodes[0] = lf2
				curTable.nodes[1] = lf1
			}

			break
		}

		// idx1 == idx2 && continue
		var newTable = new(compressedTable)
		newTable.hashPath = hashPath & (idx1 << ((d + 1) * NBITS))
		newTable.depth = d + 1

		curTable.nodeMap = 1 << idx1
		curTable.nodes = append(curTable.nodes, newTable)

		curTable = newTable
	}
	if d == MAXDEPTH {
		// lf1.hash60 == curTable.hashPath
		// lf1.hashcode() == lf2.hashcode()
		lf := lf1.put(lf2)
		curTable.set(lf.hashcode(), lf)
	}

	return retTable
}

func (t compressedTable) hashcode() uint64 {
	return t.hashPath
}

func (t compressedTable) copy() tableI {
	var nt = new(compressedTable)
	nt.hashPath = t.hashPath
	nt.depth = t.depth
	nt.nodeMap = t.nodeMap
	nt.nodes = append(nt.nodes, t.nodes...)
	return nt
}

func (t compressedTable) get(hash uint64) nodeI {
	// Get the regular index of the node we want to access
	var idx = index(hash, t.depth) // 0..63
	var nodeBit = 1 << idx

	if (t.nodeMap & nodeBit) == 0 {
		return nil
	}

	// Create a mask to mask off all bits below idx'th bit
	var m = 1<<idx - 1

	// Count the number of bits in the nodeMap below the idx'th bit
	var i = BitCount64(t.nodeMap & m)

	var node = t.nodes[i]

	return node
}

func (t compressedTable) set(hash uint64, entry nodeI) tableI {
	nt := t.copy()

	// Get the regular index of the node we want to access
	var idx = index(hash, nt.depth) // 0..63
	var bit = 1 << idx              //is the slot

	var bitMask = 1<<idx - 1

	if (nt.nodeMap & bit) == 0 {
		// slot is empty, so put nodeI in slot
		nt.nodeMap &= bit
		var i = BitCount64(nt.nodeMap & bitMask)

		//insert entry into the i'th spot of t.nodes
		nt.nodes = append(nt.nodes[:i], append([]nodeI{entry}, nt.nodes[i:]))

		return nt
	}

	// slot is used
	//
	var i = BitCount64(t.nodeMap & bitMask)

	oldEnts := oldTable.entries().copy()
	newEnts := make([]nodeI, len(oldEnts))
	copy(newEnts, oldEnts)

	var nt = NewCompressedTable(t.depth+1, hash, newEnts)

	return nt
}

func (t compressedTable) clear(hash uint64) (tableI, nodeI) {
	return nil, nil, false
}

func (t compressedTable) nentries() uint {
	return BitCount64(n.nodeMap)
}

type fullTable struct {
	hashPath uint64 // depth*NBITS of hash to get to this location in the Trie
	depth    uint
	nodeMap  uint64
	nodes    [TABLE]nodeI
}

func NewFullTable(depth uint, hash uint64, nt tableS) tableI {
	return nil
}

func (t fullTable) hashcode() uint64 {
	return t.hashPath
}

func (t fullTable) copy() nodeI {
	return nil
}

func (t fullTable) get(hash uint64) nodeI {
	return nil, false
}

func (t fullTable) set(hash uint64, entry nodeI) nodeI {
	return nil
}

func (t fullTable) put(hash uint64, entry nodeI) tableI {
	return nil, false
}

func (t fullTable) del(hash uint64) (tableI, nodeI) {
	return nil, nil, false
}

type leafI interface {
	nodeI
	//copy() leafI
	get(key []byte) (interface{}, bool)
	put(key []byte, val interface{}) (leafI, bool)
	del(key []byte) (leafI, interface{}, bool)
}

type flatLeaf struct {
	hash60 uint64 //hash64(key) & sixtyBitMask
	key    []byte
	val    interface{}
}

func NewFlatLeaf(hash uint64, key []byte, val interface{}) flatLeaf {
	hash60 = hash & sixtyBitMask
	var leaf = flatLeaf{hash60, key, val}
	return leaf
}

func (l flatLeaf) hashcode() uint64 {
	return l.hash60
}

//func (l flatLeaf) copy() leafI {
//	nl = new(flatLeaf)
//	nl.hash60 = l.hash60
//	nl.key = l.key
//	nl.val = l.val
//	return nl
//}

func (l flatLeaf) get(key []byte) (interface{}, bool) {
	if byteSlicesEqual(l.key, key) {
		return l.val, true
	}
	return nil, false
}

//
func (l floatLeaf) put(key []byte, value interface{}) (leafI, bool) {
	if byteSlicesEqual(l.key, key) {
		oval := l.val
		l.val = value

	}

	var cl = new(collisionLeaf)
	cl.hash60 = l.hashcode()
	//cl.kvs = make([]keyVal, 0, 2)
	cl.kvs = append(cl.kvs, keyVal{l.key, l.val})
	cl.kvs = append(cl.kvs, keyVal{key, value})
	return cl
}

func (l flatLeaf) del(key []byte) (leafI, interface{}, bool) {
	if byteSlicesEqual(l.key, key) {
		return nil, l.val, true
	}
	return nil, nil, false
}

type keyVal struct {
	key []byte
	val interface{}
}

type collisionLeaf struct {
	hash60 uint64 //hash64(key) & sixtyBitMask
	kvs    []keyVal
}

func NewCollisionLeaf(hash uint64, kvs []keyVal) collisionLeaf {
	leaf := new(collisionLeaf)
	leaf.hash60 = hash & sixtyBitMask
	leaf.kvs = append(leaf.kvs, kvs...)

	return leaf
}

func (l collisionLeaf) hashcode() uint64 {
	return l.hash60
}

func (l collisionLeaf) get(key []byte) (interface{}, bool) {
	for i := 0; i < len(l.kvs); i++ {
		if byteSlicesEqual(l.key[i], key) {
			return l.kvs[i].val, true
		}
	}
	return nil, false
}

func (l collisionLeaf) put(key []byte, val interface{}) (leafI, bool) {
	nl := new(collisionLeaf)
	nl.hash60 = l.hash60
	nl.kvs = append(nl.kvs, l.kvs...)
	nl.kvs = append(nl.kvs, keyVal{key, val})
	return nl
}

func (l collisionLeaf) del(key []byte) (leafI, interface{}, bool) {
	if len(l.kvs) == 2 {
		if byteSliceEqual(key, l.kvs[0].key) {
			return NewFlatLeaf(l.hash60, l.kvs[1].key, l.kvs[1].val), l.kvs[0].val, true
		}
		if byteSliceEqual(key, l.kvs[1].key) {
			return NewFlatLeaf(l.hash60, l.kvs[0].key, l.kvs[0].val), l.kvs[1].val, true
		}
		return nil, nil, false
	}

	var retval interface{}
	var removed bool = false
	var nl = new(collisionLeaf)

	nl.hash60 = l.hash60

	for i := 0; i < len(l.kvs); i++ {
		if !byteSliceEqual(key, l.kvs[i].key) {
			nl.kvs = append(nl.kvs, l.kvs[i])
		} else {
			retVal = l.kvs[i].val
			removed = true
		}
	}

	return nl, retVal, removed
}

//byteSliceEqual function
//
func byteSlicesEqual(k1 []byte, k2 []byte) bool {
	if len(k1) != len(k2) {
		return false
	}
	if len(k1) == 0 {
		return true
	}
	for i := 0; i < len(k1); i++ {
		if k1[i] != k2[i] {
			return false
		}
	}
	return true
}

//POPCNT Implementation
//

const (
	HEXI_FIVES  = uint64(0x5555555555555555)
	HEXI_THREES = uint64(0x3333333333333333)
	HEXI_ONES   = uint64(0x0101010101010101)
	HEXI_FS     = uint64(0x0f0f0f0f0f0f0f0f)
)

func BitCount64(n uint64) uint {
	n = n - ((n >> 1) & HEXI_FIVES)
	n = (n & HEXI_THREES) + ((n >> 2) & HEXI_THREES)
	return uint((((n + (n >> 4)) & HEXI_FS) * HEXI_ONES) >> 56)
}
