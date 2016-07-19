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
package hamt_functional

import (
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"strings"
)

var Lgr = log.New(os.Stderr, "[hamt] ", log.Lshortfile)

// The number of bits to partition the hashcode and to index each table. By
// logical necessity this MUST be 6 bits because 2^6 == 64; the number of
// entries in a table.
const NBITS64 uint = 6

// The Capacity of a table; 2^6 == 64;
const TABLE_CAPACITY uint = 1 << NBITS64

const mask60 = 1<<60 - 1

// The maximum depthof a HAMT ranges between 0 and 9, for 10 levels
// total and cunsumes 60 of the 64 bit hashcode.
const MAXDEPTH uint = 9

const ASSERT_CONST bool = true

func ASSERT(test bool, msg string) {
	if ASSERT_CONST {
		if !test {
			panic(msg)
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

// Creatig non-64bit hash from:
// http://www.isthe.com/chongo/tech/comp/fnv/index.html#xor-fold
func hash60(bs []byte) uint64 {
	var h64 = hash64(bs)
	return (h64 >> 60) ^ (h64 & mask60)
}

func hashPathEqual(depth uint, a, b uint64) bool {
	pathMask := uint64(1<<(depth*NBITS64)) - 1

	return (a & pathMask) == (b & pathMask)
}

//func makeHashPath(h60 uint64, depth uint) uint64 {
//	return hash & hashPathMask(depth)
//}

func hashPathMask(depth uint) uint64 {
	return uint64(1<<((depth)*NBITS64)) - 1
}

// Create a string of the form "/%02d/%02d..." to describe a hashPath of
// a given depth.
//
// If you want hashPathStrig() to include the current idx, you Must
// add one to depth. You may need to do this because you are creating
// a table to be put at the idx'th slot of the current table.
func hashPathString(hashPath uint64, depth uint) string {
	if depth == 0 {
		return ""
	}
	var strs = make([]string, depth)

	for d := depth; d > 0; d-- {
		var idx = index(hashPath, d-1)
		strs[d-1] = fmt.Sprintf("%02d", idx)
	}
	return "/" + strings.Join(strs, "/")
}

func hash60String(h60 uint64) string {
	return hashPathString(h60, 10)
}

func nodeMapString(nodeMap uint64) string {
	var strs = make([]string, 7)

	var top4 = nodeMap >> 60
	strs[0] = fmt.Sprintf("%04b", top4)

	const tenBitMask uint64 = 1<<10 - 1
	for i := int(0); i < 6; i++ {
		var ui = uint(i)
		tenBitVal := (nodeMap & (tenBitMask << (ui * 10))) >> (ui * 10)
		strs[6-ui] = fmt.Sprintf("%010b", tenBitVal)
	}

	return strings.Join(strs, " ")
}

//indexMask() generates a NBITS64(6-bit) mask for a given depth
func indexMask(depth uint) uint64 {
	//var eightBitMask = uint64(uint8(1<<NBITS64) - 1)
	//return eightBitMask << (depth * NBITS64)
	return uint64(uint8(1<<NBITS64)-1) << (depth * NBITS64)
}

//index() calculates a NBITS64(6-bit) integer based on the hash and depth
func index(h60 uint64, depth uint) uint {
	var idxMask = indexMask(depth)
	var idx = uint((h60 & idxMask) >> (depth * NBITS64))
	return idx
}

//buildHashPath(hashPath, idx, depth)
func buildHashPath(hashPath uint64, idx, depth uint) uint64 {
	return hashPath | uint64(idx<<(depth*NBITS64))
}

type Hamt struct {
	root     tableI
	nentries uint
}

var EMPTY = Hamt{nil, 0}

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

func (h Hamt) IsEmpty() bool {
	return h.root == nil
}

func (h Hamt) copy() *Hamt {
	var nh = new(Hamt)
	nh.root = h.root
	nh.nentries = h.nentries
	return nh
}

func (h *Hamt) copyUp(oldTable, newTable tableI, path pathT) {
	if path.isEmpty() {
		h.root = newTable
		return
	}

	var depth = uint(len(path))
	var parentDepth = depth - 1

	oldParent := path.pop()

	var parentIdx = index(oldTable.hashcode(), parentDepth)
	var newParent = oldParent.set(parentIdx, newTable)
	h.copyUp(oldParent, newParent, path)

	return
}

// Get(key) retrieves the value for a given key from the Hamt. The bool
// represents whether the key was found.
func (h Hamt) Get(key []byte) (interface{}, bool) {
	if h.IsEmpty() {
		return nil, false
	}

	var h60 = hash60(key)

	// We know h.root != nil (above IsEmpty test) and h.root is a tableI
	// intrface compliant struct.
	var curTable = h.root

	for depth := uint(0); depth <= MAXDEPTH; depth++ {
		var idx = index(h60, depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			break
		}

		//if curNode ISA leafI
		if leaf, ok := curNode.(leafI); ok {
			if hashPathEqual(depth, h60, leaf.hashcode()) {
				return leaf.get(key)
			}
			return nil, false
		}

		//else curNode MUST BE A tableI
		curTable = curNode.(tableI)
	}
	// curNode == nil || depth > MAXDEPTH

	return nil, false
}

func (h Hamt) Put(key []byte, val interface{}) (nh *Hamt, inserted bool) {
	nh = h.copy()
	//inserted = true //true == inserted key/val pair; false == replaced val

	var h60 = hash60(key)
	var depth uint = 0
	var newLeaf = NewFlatLeaf(h60, key, val)

	if h.IsEmpty() {
		nh.root = newCompressedTable(depth, h60, newLeaf)
		nh.nentries++
		return nh, true
	}

	var path = newPathT()
	var hashPath uint64 = 0
	var curTable = h.root

	for depth = 0; depth < MAXDEPTH; depth++ {
		var idx = index(h60, depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			var newTable = curTable.set(idx, newLeaf)
			nh.nentries++
			nh.copyUp(curTable, newTable, path)
			return nh, true
		}

		if oldLeaf, ok := curNode.(leafI); ok {

			if oldLeaf.hashcode() == h60 {
				Lgr.Printf("HOLY SHIT!!! Two keys collided with this same hash60 orig key=\"%s\" new key=\"%s\" h60=0x%016x", oldLeaf.(flatLeaf).key, key, h60)

				var newLeaf leafI
				newLeaf, inserted = oldLeaf.put(key, val)
				if inserted {
					nh.nentries++
				}
				var newTable = curTable.set(idx, newLeaf)
				nh.copyUp(curTable, newTable, path)

				return
			}

			// Ok newLeaf & oldLeaf are colliding thus we create a new table and
			// we are going to insert it into this curTable.
			//
			// hashPath is already describes the curent depth; so to add the
			// idx onto hashPath, you must add +1 to the depth.
			hashPath = buildHashPath(hashPath, idx, depth)

			var newLeaf = NewFlatLeaf(h60, key, val)

			//Can I calculate the hashPath from path? Should I go there? ;}

			collisionTable := newCompressedTable2(depth+1, hashPath, oldLeaf, *newLeaf)

			newTable := curTable.set(idx, collisionTable)

			nh.nentries++
			nh.copyUp(curTable, newTable, path)
			return nh, true
		} //if curNode ISA leafI

		hashPath = buildHashPath(hashPath, idx, depth)

		path.push(curTable)

		// The table entry is NOT NIL & NOT LeafI THEREFOR MUST BE a tableI
		curTable = curNode.(tableI)

	} //end: for

	return nh, false
}

// Hamt.Del(key) returns a new Hamt, the value deleted, and a boolean that
// specifies whether or not the key was deleted (eg it didn't exist to start
// with). Therefor you must always test deleted before using the new *Hamt
// value.
func (h Hamt) Del(key []byte) (nh *Hamt, val interface{}, deleted bool) {
	nh = h.copy()

	var h60 = hash60(key)
	var depth uint = 0

	var path = newPathT()
	var curTable = h.root

	for depth = 0; depth <= MAXDEPTH; depth++ {
		var idx = index(h60, depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			return nil, nil, false
		}

		if oldLeaf, ok := curNode.(leafI); ok {
			if oldLeaf.hashcode() != h60 {
				// Found a leaf, but not the leaf I was looking for.
				Lgr.Printf("h.Del(%q): depth=%d; h60=%s", key, depth, hash60String(h60))
				Lgr.Printf("h.Del(%q): idx=%d", key, idx)
				Lgr.Printf("h.Del(%q): curTable=\n%s", key, curTable.LongString("", depth))
				Lgr.Printf("h.Del(%q): Found a leaf, but not the leaf I was looking for; depth=%d; idx=%d; oldLeaf=%s", key, depth, idx, oldLeaf)
				return nil, nil, false
			}

			var newLeaf, val, deleted = oldLeaf.del(key)

			//var idx = index(oldLeaf.hashcode(), depth)
			var newTable = curTable.set(idx, newLeaf)

			nh.copyUp(curTable, newTable, path)

			if deleted {
				nh.nentries--
			}

			return nh, val, deleted
		}

		// curTable now becomes the parentTable and we push it on to the path
		path.push(curTable)

		// the curNode MUST BE a tableI so we coerce and set it to curTable
		curTable = curNode.(tableI)
	}
	// depth == MAXDEPTH+1 & no leaf with key was found
	// So after a thourough search no key/value exists to delete.

	return nil, nil, false
}
