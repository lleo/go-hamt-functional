/*
Package hamt32_functional implements a functional Hash Array Mapped Trie (HAMT).
It is called hamt32_functional because this package is using 32bits of hash for
the indexes into each level of the Trie. The term functional is used to imply
immutable and persistent.

All 32bits of the hash is used as we fold the 2 high bits into the lower 30bits
as described in http://www.isthe.com/chongo/tech/comp/fnv/index.html#xor-fold .

The 30bits of hash are separated into six 5bit values that constitue the hash
path of any Key in this Trie. However, not all six levels of the Trie are used.
As many levels (six or less) are used to find a unique location
for the leaf to be placed within the Trie.

If all six levels of the Trie are used for two or more key/val pairs then a
special collision leaf will be found at the sixth level of the Trie. It keeps
all key/val pairs in a slice.
*/
package hamt32_functional

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/lleo/go-hamt/hamt_key"
)

func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("[hamt] ")
	log.SetFlags(log.Lshortfile)
}

// The number of bits to partition the hashcode and to index each table. By
// logical necessity this MUST be 5 bits because 2^5 == 32; the number of
// entries in a table.
const NBITS uint = 5

// The Capacity of a table; 2^5 == 32;
const TABLE_CAPACITY uint = 1 << NBITS

const mask30 = 1<<30 - 1

// The maximum depth of a HAMT ranges between 0 and 5, for 6 levels total.
const MAXDEPTH uint = 5

const assert_const bool = true

func assert(test bool, msg string) {
	if assert_const {
		if !test {
			panic(msg)
		}
	}
}

func hashPathMask(depth uint) uint32 {
	return uint32(1<<((depth)*NBITS)) - 1
}

// Create a string of the form "/%02d/%02d..." to describe a hashPath of
// a given depth.
//
// If you want hashPathStrig() to include the current idx, you Must
// add one to depth. You may need to do this because you are creating
// a table to be put at the idx'th slot of the current table.
func hashPathString(hashPath uint32, depth uint) string {
	if depth == 0 {
		return "/"
	}
	var strs = make([]string, depth+1)

	for d := int32(depth); d >= 0; d-- {
		var idx = index(hashPath, uint(d))
		strs[d] = fmt.Sprintf("%02d", idx)
	}
	return "/" + strings.Join(strs, "/")
}

func hash30String(h30 uint32) string {
	return hashPathString(h30, MAXDEPTH)
}

func nodeMapString(nodeMap uint32) string {
	var strs = make([]string, 7)

	var top2 = nodeMap >> 30
	strs[0] = fmt.Sprintf("%02b", top2)

	const tenBitMask uint32 = 1<<10 - 1
	for i := int(0); i < 5; i++ {
		var ui = uint(i)
		tenBitVal := (nodeMap & (tenBitMask << (ui * 10))) >> (ui * 10)
		strs[5-ui] = fmt.Sprintf("%010b", tenBitVal)
	}

	return strings.Join(strs, " ")
}

func hashPathMatches(hashPath uint32, hashPathStr string) bool {
	//make sure hashPathStr is well formed
	if !strings.HasPrefix(hashPathStr, "/") {
		log.Printf("HashPathMatches: hashPathStr,%q does not start with a \"/\"\n", hashPathStr)
		return false
	}
	if strings.HasSuffix(hashPathStr, "/") {
		log.Printf("HashPathMatches: hashPathStr,%q ends with a \"/\"\n", hashPathStr)
		return false
	}

	//convert hashPathStr to a uint32 bit string
	var strs = strings.Split(hashPathStr, "/")
	strs = strs[1:] //take the first empty off
	var bitStr uint32
	for i, s := range strs {
		//var nn int64
		//var err error
		nn, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			log.Printf("strconv.PareseInt(s=%q, 10, 32) failed", s)
			return false
		}
		n := uint32(nn)
		if n > 31 {
			log.Printf("the i,%d entry,%d of hashPathStr,%s is >31", i, n, hashPathStr)
			return false
		}
		bitStr |= n << ((NBITS - uint(i)) * NBITS)
		//log.Printf("HashPathMatches: bitStr=%032b", bitStr)
	}

	if (hashPath & bitStr) == bitStr {
		return true
	} else {
		return false
	}
}

//indexMask() generates a NBITS(5-bit) mask for a given depth
func indexMask(depth uint) uint32 {
	return uint32(uint8(1<<NBITS)-1) << (depth * NBITS)
}

//index() calculates a NBITS(5-bit) integer based on the hash and depth
func index(h30 uint32, depth uint) uint {
	var idxMask = indexMask(depth)
	var idx = uint((h30 & idxMask) >> (depth * NBITS))
	return idx
}

//buildHashPath(hashPath, idx, depth)
func buildHashPath(hashPath uint32, idx, depth uint) uint32 {
	return hashPath | uint32(idx<<(depth*NBITS))
}

type keyVal struct {
	key hamt_key.Key
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
	//return h.root == nil
	return h == EMPTY
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
func (h Hamt) Get(key hamt_key.Key) (interface{}, bool) {
	if h.IsEmpty() {
		return nil, false
	}

	var h30 = key.Hash30()

	// We know h.root != nil (above IsEmpty test) and h.root is a tableI
	// intrface compliant struct.
	var curTable = h.root

	for depth := uint(0); depth <= MAXDEPTH; depth++ {
		var idx = index(h30, depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			break
		}

		//if curNode ISA leafI
		if leaf, ok := curNode.(leafI); ok {
			//if hashPathEqual(depth, h30, leaf.hashcode()) {
			if leaf.hashcode() == h30 {
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

func (h Hamt) Put(key hamt_key.Key, val interface{}) (Hamt, bool) {
	var nh = h.copy()
	var inserted = true //true == inserted key/val pair; false == replaced val

	var h30 = key.Hash30()
	var depth uint = 0
	var newLeaf = newFlatLeaf(h30, key, val)

	if h.IsEmpty() {
		nh.root = newCompressedTable(depth, h30, newLeaf)
		nh.nentries++
		return *nh, inserted
	}

	var path = newPathT()
	var hashPath uint32 = 0
	var curTable = h.root

	for depth = 0; depth <= MAXDEPTH; depth++ {
		var idx = index(h30, depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			var newTable = curTable.set(idx, newLeaf)
			nh.nentries++
			nh.copyUp(curTable, newTable, path)
			return *nh, inserted
		}

		if oldLeaf, ok := curNode.(leafI); ok {

			if oldLeaf.hashcode() == h30 {
				log.Printf("HOLY SHIT!!! Two keys collided with this same hash30 orig key=\"%s\" new key=\"%s\" h30=0x%016x", oldLeaf.(flatLeaf).key, key, h30)

				var newLeaf leafI
				newLeaf, inserted = oldLeaf.put(key, val)
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

			var newLeaf = newFlatLeaf(h30, key, val)

			//Can I calculate the hashPath from path? Should I go there? ;}

			collisionTable := newCompressedTable2(depth+1, hashPath, oldLeaf, *newLeaf)

			newTable := curTable.set(idx, collisionTable)

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

// Hamt.Del(key) returns a new Hamt, the value deleted, and a boolean that
// specifies whether or not the key was deleted (eg it didn't exist to start
// with). Therefor you must always test deleted before using the new *Hamt
// value.
func (h Hamt) Del(key hamt_key.Key) (Hamt, interface{}, bool) {
	var nh = h.copy()
	var val interface{}
	var deleted bool

	var h30 = key.Hash30()
	var depth uint = 0

	var path = newPathT()
	var curTable = h.root

	for depth = 0; depth <= MAXDEPTH; depth++ {
		var idx = index(h30, depth)
		var curNode = curTable.get(idx)

		if curNode == nil {
			return h, nil, false
		}

		if oldLeaf, ok := curNode.(leafI); ok {
			if oldLeaf.hashcode() != h30 {
				// Found a leaf, but not the leaf I was looking for.
				log.Printf("h.Del(%q): depth=%d; h30=%s", key, depth, hash30String(h30))
				log.Printf("h.Del(%q): idx=%d", key, idx)
				log.Printf("h.Del(%q): curTable=\n%s", key, curTable.LongString("", depth))
				log.Printf("h.Del(%q): Found a leaf, but not the leaf I was looking for; depth=%d; idx=%d; oldLeaf=%s", key, depth, idx, oldLeaf)
				return h, nil, false
			}

			var newLeaf leafI
			newLeaf, val, deleted = oldLeaf.del(key)

			//var idx = index(oldLeaf.hashcode(), depth)
			var newTable = curTable.set(idx, newLeaf)

			nh.copyUp(curTable, newTable, path)

			if deleted {
				nh.nentries--
			}

			return *nh, val, deleted
		}

		// curTable now becomes the parentTable and we push it on to the path
		path.push(curTable)

		// the curNode MUST BE a tableI so we coerce and set it to curTable
		curTable = curNode.(tableI)
	}
	// depth == MAXDEPTH+1 & no leaf with key was found
	// So after a thourough search no key/value exists to delete.

	return h, nil, false
}
