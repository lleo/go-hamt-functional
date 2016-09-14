/*
Package hamt64 implements a functional Hash Array Mapped Trie (HAMT).
It is called hamt64 because this package is using 64bits of hash for
the indexes into each level of the Trie. The term functional is used to imply
immutable and persistent.

All 64bits of the hash is used as we fold the 4 high bits into the lower 60bits
as described in http://www.isthe.com/chongo/tech/comp/fnv/index.html#xor-fold .

The 60bits of hash are separated into ten 6bit values that constitue the hash
path of any Key in this Trie. However, not all ten levels of the Trie are used.
As many levels (ten or less) are used to find a unique location for the leaf to
be placed within the Trie.

If all ten levels of the Trie are used for two or more key/val pairs then a
special collision leaf will be used to store those key/val pairs, at the 10nth
level of the Trie.
*/
package hamt64

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/lleo/go-hamt/key"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[hamt] ")
	log.SetFlags(log.Lshortfile)
}

// NBITS constant is the number of bits, 6bits, a 60bit hash value is split
// into to provied the indexes of a HAMT.
const NBITS uint = 6

// MAXDEPTH constant is the maximum depth(9) of NBITS values that constitute
// the path in a HAMT, from [0..MAXDEPTH] for a total of MAXDEPTH+1(10) levels.
// NBITS*(MAXDEPTH+1) == HASHBITS (ie 6*(9+1) == 60).
const MAXDEPTH uint = 9

// TABLE_CAPACITY constant is the number of table entries in a each node of
// a HAMT datastructure; its value is 1<<NBITS (ie 2^6 == 64).
const TABLE_CAPACITY uint = 1 << NBITS

// GRADE_TABLES is a boolean to enable/disable of upgrading & downgrading.
// true turns upgrading/downgrading ON.
const GRADE_TABLES = false

const assert_const bool = true

func assert(test bool, msg string) {
	if assert_const {
		if !test {
			panic(msg)
		}
	}
}

func hashPathMask(depth uint) uint64 {
	return uint64(1<<((depth)*NBITS)) - 1
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
	var strs = make([]string, depth+1)

	for d := int(depth); d >= 0; d-- {
		var idx = index(hashPath, uint(d))
		strs[d] = fmt.Sprintf("%02d", idx)
	}
	return "/" + strings.Join(strs, "/")
}

func hash60String(hashPath uint64) string {
	return hashPathString(hashPath, MAXDEPTH)
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

//indexMask() generates a NBITS(6-bit) mask for a given depth
func indexMask(depth uint) uint64 {
	return uint64(uint8(1<<NBITS)-1) << (depth * NBITS)
}

//index() calculates a NBITS(6-bit) integer based on the hash and depth
func index(h60 uint64, depth uint) uint {
	var idxMask = indexMask(depth)
	var idx = uint((h60 & idxMask) >> (depth * NBITS))
	return idx
}

//buildHashPath(hashPath, idx, depth)
func buildHashPath(hashPath uint64, idx, depth uint) uint64 {
	return hashPath | uint64(idx<<(depth*NBITS))
}

type keyVal struct {
	key key.Key
	val interface{}
}

func (kv keyVal) String() string {
	return fmt.Sprintf("keyVal{%s, %v}", kv.key, kv.val)
}

type Hamt struct {
	root     tableI
	nentries uint
}

// The EMPTY Hamt struct is also the zero value of the Hamt struct and represents
// an empty Hash Arry Mapped Trie.
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

// Get(k) retrieves the value for a given key from the Hamt. The bool
// represents whether the key was found.
func (h Hamt) Get(k key.Key) (interface{}, bool) {
	if h.IsEmpty() {
		return nil, false
	}

	var h60 = key.Hash60(k)

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
			//if hashPathEqual(depth, h60, leaf.hashcode()) {
			if leaf.hashcode() == h60 {
				return leaf.get(k)
			}
			return nil, false
		}

		//else curNode MUST BE A tableI
		curTable = curNode.(tableI)
	}
	// curNode == nil || depth > MAXDEPTH

	return nil, false
}

// Hamt.Put(k, v) returns a new Hamt structure and FIXME
func (h Hamt) Put(k key.Key, v interface{}) (Hamt, bool) {
	var nh = h.copy()
	var inserted = true //true == inserted key/val pair; false == replaced val

	var h60 = key.Hash60(k)
	var depth uint = 0
	var newLeaf = newFlatLeaf(h60, k, v)

	if h.IsEmpty() {
		nh.root = newTable(depth, h60, newLeaf)
		nh.nentries++
		return *nh, inserted
	}

	var path = newPathT()
	var hashPath uint64 = 0
	var curTable = h.root

	for depth = 0; depth <= MAXDEPTH; depth++ {
		var idx = index(h60, depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			var newTable = curTable.set(idx, newLeaf)
			nh.nentries++
			nh.copyUp(curTable, newTable, path)
			return *nh, inserted
		}

		if oldLeaf, ok := curNode.(leafI); ok {

			if oldLeaf.hashcode() == h60 {
				log.Printf("HOLY SHIT!!! Two keys collided with this same hash60 orig key=\"%s\" new key=\"%s\" h60=0x%016x", oldLeaf.(flatLeaf).key, k, h60)

				var newLeaf leafI
				newLeaf, inserted = oldLeaf.put(k, v)
				if inserted {
					nh.nentries++
				}
				var newTable = curTable.set(idx, newLeaf)
				nh.copyUp(curTable, newTable, path)

				return *nh, inserted
			}

			// Ok newLeaf & oldLeaf are colliding thus we create a new table and
			// we are going to insert it into this curTable.
			//
			// hashPath is already describes the curent depth; so to add the
			// idx onto hashPath, you must add +1 to the depth.
			hashPath = buildHashPath(hashPath, idx, depth)

			var newLeaf = newFlatLeaf(h60, k, v)

			//Can I calculate the hashPath from path? Should I go there? ;}

			tmpTable := newTable2(depth+1, hashPath, oldLeaf, *newLeaf)

			newTable := curTable.set(idx, tmpTable)

			nh.nentries++
			nh.copyUp(curTable, newTable, path)
			return *nh, inserted
		} //if curNode ISA leafI

		hashPath = buildHashPath(hashPath, idx, depth)

		path.push(curTable)

		// The table entry is NOT NIL & NOT LeafI THEREFOR MUST BE a tableI
		curTable = curNode.(tableI)

	} //end: for

	inserted = false
	return *nh, inserted
}

// Hamt.Del(k) returns a new Hamt, the value deleted, and a boolean that
// specifies whether or not the key was deleted (eg it didn't exist to start
// with). Therefor you must always test deleted before using the new Hamt
// value.
func (h Hamt) Del(k key.Key) (Hamt, interface{}, bool) {
	var nh = h.copy()
	var v interface{}
	var deleted bool

	var h60 = key.Hash60(k)
	var depth uint = 0

	var path = newPathT()
	var curTable = h.root

	for depth = 0; depth <= MAXDEPTH; depth++ {
		var idx = index(h60, depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			return h, nil, false
		}

		if oldLeaf, ok := curNode.(leafI); ok {
			if oldLeaf.hashcode() != h60 {
				// Found a leaf, but not the leaf I was looking for.
				log.Printf("h.Del(%q): depth=%d; h60=%s", k, depth, hash60String(h60))
				log.Printf("h.Del(%q): idx=%d", k, idx)
				log.Printf("h.Del(%q): curTable=\n%s", k, curTable.LongString("", depth))
				log.Printf("h.Del(%q): Found a leaf, but not the leaf I was looking for; depth=%d; idx=%d; oldLeaf=%s", k, depth, idx, oldLeaf)
				return h, nil, false
			}

			var newLeaf leafI
			newLeaf, v, deleted = oldLeaf.del(k)

			//var idx = index(oldLeaf.hashcode(), depth)
			var newTable = curTable.set(idx, newLeaf)

			nh.copyUp(curTable, newTable, path)

			if deleted {
				nh.nentries--
			}

			return *nh, v, deleted
		}

		// curTable now becomes the parentTable and we push it on to the path
		path.push(curTable)

		// the curNode MUST BE a tableI so we coerce and set it to curTable
		curTable = curNode.(tableI)
	}
	// depth > MAXDEPTH & no leaf with key was found
	// So after a thourough search no key/value exists to delete.

	return h, nil, false
}
