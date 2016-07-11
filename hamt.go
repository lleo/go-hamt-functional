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
const NBITS uint = 6

// The Capacity of a table; 2^6 == 64;
const TABLE_CAPACITY uint = 1 << NBITS

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

func hash60(bs []byte) uint64 {
	var h64 = hash64(bs)
	return (h64 >> 60) ^ (h64 & mask60)
}

func hashPathEqual(depth uint, a, b uint64) bool {
	pathMask := uint64(1<<(depth*NBITS)) - 1

	return (a & pathMask) == (b & pathMask)
}

func makeHashPath(depth uint, hash uint64) uint64 {
	return hash & hashPathMask(depth)
}

func hashPathMask(depth uint) uint64 {
	return uint64(1<<((depth)*NBITS)) - 1
}

func hashPathString(hashPath uint64, depth uint) string {
	if depth == 0 {
		return "0"
	}
	var strs = make([]string, depth)

	for i := depth; i > 0; i-- {
		var idx = index(hashPath, i-1)
		//strs[i-1] = fmt.Sprintf("%06b", idx)
		strs[i-1] = fmt.Sprintf("%02d", idx)
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

//indexMask() generates a NBITS(6-bit) mask for a given depth
func indexMask(depth uint) uint64 {
	//var eightBitMask = uint64(uint8(1<<NBITS) - 1)
	//return eightBitMask << (depth * NBITS)
	return uint64(uint8(1<<NBITS)-1) << (depth * NBITS)
}

//index() calculates a NBITS(6-bit) integer based on the hash and depth
func index(h60 uint64, depth uint) uint {
	var idxMask = indexMask(depth)
	var idx = uint((h60 & idxMask) >> (depth * NBITS))
	return idx
}

//buildHashPath()
func buildHashPath(hashPath uint64, depth, idx uint) uint64 {
	// This code is not valid where idx == 0 for each depth
	/*
		hashPathMask_ := hashPathMask(depth)
		if ((hashPath & hashPathMask_) - 1) > 0 {
			panic("hashPath > depth")
		}
	*/

	return hashPath & uint64(idx<<(depth*NBITS))
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
		str += indent + h.root.LongString(indent)
		str += indent + "}"
		return str
	} else {
		str = indent + fmt.Sprintf("Hamt{ nentries: %d, root: %s }", h.nentries, h.root)
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

	oldParent := path.pop()

	// Cases
	// #1 newTable == nil
	// #2 newTable.nentries() == 1
	// #3 newTable.nentries() > 1
	//
	// In case #1 and #3 oldParent.set() & copyUp(oldParent, newParent)
	//
	// In case #2 find the oneEntry	in newTable
	//            and oldParent.set(oldTable.hashcode(), oneEntry)
	//            and copyUp(oldParent, newParent)

	if newTable.nentries() == 1 {
		// find the oneEntry in	newTable

		//var idx = uint(math.Log2(float64(newTable.nodeMap)))
		//var node = newTable.get(hashPath)
		//var newParent = oldParent.set(newTable.hashcode(), node)
		//h.copyUp(oldParent, newParent, path)

		var nt = newTable.(*compressedTable)

		for idx := uint(0); idx < TABLE_CAPACITY; idx++ {
			var hashPath = nt.hashPath | uint64(idx<<(nt.depth()*NBITS))

			if (nt.nodeMap & (1 << idx)) > 0 {
				var node = nt.get(hashPath)
				var newParent = oldParent.set(nt.hashcode(), node)
				h.copyUp(oldParent, newParent, path)
				break
			}
		}

	} else { // newTable == nil OR newTable.nentries() > 1
		var newParent = oldParent.set(oldTable.hashcode(), newTable)
		h.copyUp(oldParent, newParent, path)
	}

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
	var curNode = curTable.get(h60)

	for depth := uint(0); curNode != nil && depth <= MAXDEPTH; depth++ {
		//if curNode ISA leafI
		if leaf, ok := curNode.(leafI); ok {
			if hashPathEqual(depth, h60, leaf.hashcode()) {
				return leaf.get(key)
			}
			return nil, false
		}

		//else curNode MUST BE A tableI
		var curTable = curNode.(tableI)

		curNode = curTable.get(h60)
	}

	// curNode == nil
	return nil, false
}

func (h Hamt) Put(key []byte, val interface{}) (nh *Hamt, inserted bool) {
	nh = h.copy()
	//inserted = true //true == inserted key/val pair; false == replaced val

	var h60 = hash60(key)
	var depth uint = 0
	var newLeaf = NewFlatLeaf(h60, key, val)

	/**********************************************************
	 * If the Trie root is empty, insert a table with a leaf. *
	 **********************************************************/
	if h.IsEmpty() {
		nh.root = NewCompressedTable(depth, h60, newLeaf)
		nh.nentries++
		return nh, true
	}

	var path = newPathT()
	var curTable = h.root

	for depth = 0; depth < MAXDEPTH; depth++ {
		var curNode = curTable.get(h60)

		if curNode == nil {
			var newTable = curTable.set(h60, newLeaf)
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
				var newTable = curTable.set(oldLeaf.hashcode(), newLeaf)
				nh.copyUp(curTable, newTable, path)

				return
			}

			var newLeaf = NewFlatLeaf(h60, key, val)

			var hashPath = h60 & hashPathMask(depth+1)

			collisionTable := NewCompressedTable2(depth+1, hashPath, oldLeaf, *newLeaf)

			newTable := curTable.set(h60, collisionTable)

			// SANITY TEST
			// checking oldLeaf & newLeaf share the same hashPath upto depth+1
			var tmpHashPath uint64
			for i := uint(0); i < depth+1; i++ {
				idx1 := index(oldLeaf.hashcode(), i)
				idx2 := index(newLeaf.hashcode(), i)
				if idx1 != idx2 {
					//tmpHashPath1 := tmpHashPath | uint64(idx1<<i*NBITS)
					//tmpHashPath2 := tmpHashPath | uint64(idx2<<i*NBITS)
					Lgr.Printf("h=\n%s", h.LongString(""))
					Lgr.Printf("collisionTable=\n%s", collisionTable.LongString(""))
					Lgr.Printf("curTable=\n%s", curTable.LongString(""))
					Lgr.Printf("newTable=\n%s", newTable.LongString(""))
					Lgr.Printf("idx1=%d; idx2=%d", idx1, idx2)
					Lgr.Printf("depth+1=%d; i=%d", depth+1, i)
					Lgr.Panicf("attempted to call NewCompressedTable2 with two leaves with diffent hashPaths; oldLeaf=%s newLeaf=%s", oldLeaf, newLeaf)
				}
				tmpHashPath |= uint64(idx1 << (i * NBITS))
			}

			nh.nentries++
			nh.copyUp(curTable, newTable, path)
			return nh, true
		}

		path.push(curTable)

		// The table entry is NOT NIL & NOT LeafI THEREFOR MUST BE a tableI
		curTable = curNode.(tableI)

	} //end: for

	return nh, false
}

func (h Hamt) Del(key []byte) (nh *Hamt, val interface{}, deleted bool) {
	if h.IsEmpty() {
		return nil, nil, false
	}

	nh = h.copy()

	var h60 = hash60(key)
	var depth uint = 0

	var path = newPathT()
	var curTable = h.root
	var curNode = curTable.get(h60)

	for depth = 0; curNode != nil && depth <= MAXDEPTH; depth++ {
		if oldLeaf, ok := curNode.(leafI); ok {
			if oldLeaf.hashcode() == h60 {
				var newLeaf, val, deleted = oldLeaf.del(key)
				var newTable = curTable.set(oldLeaf.hashcode(), newLeaf)

				nh.copyUp(curTable, newTable, path)

				if deleted {
					nh.nentries--
				}

				return nh, val, deleted
			}

			// Found a leaf, but not the leaf I was looking for.
			return nil, nil, false
		}

		// curTable now becomes the parentTable and we push it on to the path
		path.push(curTable)

		// the curNode MUST BE a tableI so we coerce and set it to curTable
		curTable = curNode.(tableI)

		curNode = curTable.get(h60)
	}
	// curNode == nil

	return nil, nil, false
}

// nodeI is the interface for every entry in a table; so table entries are
// either a leaf or a table or nil.
//
// The nodeI interface can be for compressedTable, fullTable, flatLeaf, or
// collisionLeaf.
//
// The tableI interface is for compressedTable and fullTable.
//
// The hashcode() method for leaf structs is the 60 most significant bits of
// the keys hash.
//
// The hashcode() method for table structs is the depth*NBITS of the hash path
// that leads to the table's position in the Trie.
//
// For leafs hashcode() is the 60 bits returned by hash60(key).
// For collisionLeafs this is the definition of what a collision is.
//
type nodeI interface {
	hashcode() uint64
	String() string
	LongString(string) string
}

// Every tableI is a nodeI.
//
type tableI interface {
	nodeI
	depth() uint
	nentries() uint // get the number of node entries

	// Get an Ordered list of index and node pairs. This slice MUST BE Ordered
	// from lowest index to highest.
	entries() []tableEntry

	get(hash uint64) nodeI               // get an entry
	set(hash uint64, entry nodeI) tableI // set and entry
	del(hash uint64) (tableI, nodeI)     // delete an entry

	//get(idx uint) nodeI               // get an entry
	//set(idx uint, entry nodeI) tableI // set and entry
	//del(idx uint) (tableI, nodeI)     // delete an entry
}

// The compressedTable is a low memory usage version of a fullTable. It applies
// to tables with less than TABLE_CAPACITY/2 number of entries in the table.
//
// It records which table entries are populated using a bit map called nodeMap.
//
// It stores the nodes in a go slice starting with the node corresponding to
// the Least Significant Bit(LSB) is the first node in the slice. While precise
// and accurate this discription does not help boring regular programmers. Most
// bit patterns are drawn from the Most Significant Bits(MSB) to the LSB; in
// orther words for a uint64 from the 63rd bit to the 0th bit left to right. So
// for a 8bit number 1 is writtent as 00000001 (where the LSB is 1) and 128 is
// written as 10000000 (where the MSB is 1).
//
// So the number of entries in the node slice is equal to the number of bits set
// in the nodeMap. You can count the number of bits in the nodeMap, a 64bit word,
// by calculating the Hamming Weight (another obscure name; google it). The
// simple most generice way of calculating the Hamming Weight of a 64bit work is
// implemented in the BitCount64(uint64) function defined bellow.
//
// To figure out the index of a node in the nodes slice from the index of the bit
// in the nodeMap we first find out if that bit in the nodeMap is set by
// calculating if "nodeMap & (1<<idx) > 0" is true the idx'th bit is set. Given
// that each 64 entry table is indexed by 6bit section (2^6==64) of the key hash,
// there is a function to calculate the index called index(hash, depth);
//
type compressedTable struct {
	hashPath uint64 // depth*NBITS of hash to get to this location in the Trie
	depth_   uint
	nodeMap  uint64
	nodes    []nodeI
}

type tableEntry struct {
	idx  uint
	node nodeI
}

func NewCompressedTable(depth uint, h60 uint64, lf leafI) tableI {
	//ASSERT(depth < MAXDEPTH+1, "uint parameter 0 >= depth >= 9")

	var idx = index(lf.hashcode(), depth)

	var ct = new(compressedTable)
	ct.hashPath = h60 & hashPathMask(depth)
	ct._depth = depth
	ct.nodeMap = 1 << idx
	ct.nodes = make([]nodeI, 1)
	ct.nodes[0] = lf

	return ct
}

func NewCompressedTable2(depth uint, hashPath uint64, leaf1 leafI, leaf2 flatLeaf) tableI {
	//ASSERT(depth < MAXDEPTH+1, "uint parameter 0 >= depth >= 9")

	var retTable = new(compressedTable) // return compressedTable
	retTable.hashPath = hashPath & hashPathMask(depth)
	retTable._depth = depth

	var curTable = retTable //*compressedTable
	var d uint
	for d = depth; d < MAXDEPTH; d++ {
		var idx1 = index(leaf1.hashcode(), d)
		var idx2 = index(leaf2.hashcode(), d)

		if idx1 != idx2 {
			curTable.nodes = make([]nodeI, 2)

			curTable.nodeMap |= 1 << idx1
			curTable.nodeMap |= 1 << idx2
			if idx1 < idx2 {
				curTable.nodes[0] = leaf1
				curTable.nodes[1] = leaf2
			} else {
				curTable.nodes[0] = leaf2
				curTable.nodes[1] = leaf1
			}

			break
		}
		// idx1 == idx2 && continue

		curTable.nodes = make([]nodeI, 1)

		var newTable = new(compressedTable)

		//newTable.hashPath = buildHashPath(hashPath, d+1, idx1)
		newTable.hashPath = hashPath | uint64(idx1<<(d*NBITS))
		newTable._depth = d + 1

		curTable.nodeMap = 1 << idx1 //Set the idx1'th bit
		curTable.nodes[0] = newTable

		curTable = newTable
	}
	// We either BREAK out of the loop,
	// OR we hit d = MAXDEPTH.
	if d == MAXDEPTH {
		// leaf1.hashcode() == leaf2.hashcode()
		leaf, _ := leaf1.put(leaf2.key, leaf2.val)
		curTable.set(leaf.hashcode(), leaf)
	}

	return retTable
}

// DowngradeToCompressedTable() converts fullTable structs that have less than
// TABLE_CAPACITY/2 tableEntry's. One important thing we know is that none of
// the entries will collide with another.
//
// The ents []tableEntry slice is guaranteed to be in order from lowest idx to
// highest. tableI.entries() also adhears to this contract.
func DowngradeToCompressedTable(depth uint, hashPath uint64, ents []tableEntry) *compressedTable {
	var nt = new(compressedTable)
	nt.hashPath = hashPath
	nt._depth = depth
	//nt.nodeMap = 0
	nt.nodes = make([]nodeI, len(ents))

	for i := 0; i < len(ents); i++ {
		var ent = ents[i]
		var nodeBit = uint64(1 << ent.idx)
		nt.nodeMap |= nodeBit
		nt.nodes[i] = ent.node
	}

	return nt
}

func (t compressedTable) depth() uint {
	return t._depth
}

func (t compressedTable) hashcode() uint64 {
	return t.hashPath
}

func (t compressedTable) copy() *compressedTable {
	var nt = new(compressedTable)
	nt.hashPath = t.hashPath
	nt._depth = t._depth
	nt.nodeMap = t.nodeMap
	nt.nodes = append(nt.nodes, t.nodes...)
	return nt
}

//String() is required for nodeI
func (t compressedTable) String() string {
	// compressedTale{depth:%d, hashPath:[%d,%d,%d], nentries:%d,}
	return fmt.Sprintf("compressedTable{depth:%d, hashPath:%s, nentries()=%d}",
		t.depth(), hashPathString(t.hashPath, t.depth()), t.nentries())
}

// LongString() is required for nodeI
func (t compressedTable) LongString(indent string) string {
	var strs = make([]string, 3+len(t.nodes))

	strs[0] = indent + fmt.Sprintf("compressedTable{depth=%d, hashPath=%s, nentries()=%d", t.depth(), hashPathString(t.hashPath, t.depth()), t.nentries())

	strs[1] = indent + "\tnodeMap=" + nodeMapString(t.nodeMap) + ","

	//strs[2] = indent+fmt.Sprintf("\tlen(t.nodes)=%d,", len(t.nodes))

	for i, n := range t.nodes {
		if _, ok := n.(tableI); ok {
			strs[2+i] = indent + fmt.Sprintf("\tt.nodes[%d]:\n%s", i, n.LongString(indent+"\t"))
		} else {
			strs[2+i] = indent + fmt.Sprintf("\tt.nodes[%d]: %s", i, n.String())
		}
	}

	strs[len(strs)-1] = indent + "}"

	return strings.Join(strs, "\n")
}

func (t compressedTable) nentries() uint {
	return BitCount64(t.nodeMap)
}

// This function MUST return the slice of tableEntry structs from lowest
// tableEntry.idx to highest tableEntry.idx .
func (t compressedTable) entries() []tableEntry {
	var n = t.nentries()
	var ents = make([]tableEntry, n)

	for i, j := uint(0), uint(0); i < TABLE_CAPACITY; i++ {
		var nodeBit = uint64(1 << i)

		if (t.nodeMap & nodeBit) > 0 {
			ents[j] = tableEntry{i, t.nodes[j]}
			j++
		}
	}

	return ents
}

func (t compressedTable) get(hashPath uint64) nodeI {
	// Get the regular index of the node we want to access
	var idx = index(hashPath, t.depth()) // 0..63
	var nodeBit = uint64(1 << idx)

	if (t.nodeMap & nodeBit) == 0 {
		return nil
	}

	// Create a mask to mask off all bits below idx'th bit
	var m = uint64(1<<idx) - 1

	// Count the number of bits in the nodeMap below the idx'th bit
	var i = BitCount64(t.nodeMap & m)

	var node = t.nodes[i]

	return node
}

// set(uint64, nodeI) is required for tableI
func (t compressedTable) set(hashPath uint64, nn nodeI) tableI {
	var nt = t.copy()

	// Get the regular index of the node we want to access
	var idx = index(hashPath, t.depth()) // 0..63
	var nodeBit = uint64(1 << idx)       // idx is the slot
	var bitMask = nodeBit - 1            // mask all bits below the idx'th bit

	// Calculate the index into compressedTable.nodes[] for this entry
	var i = BitCount64(t.nodeMap & bitMask)

	if nn != nil {
		if (t.nodeMap & nodeBit) == 0 {
			nt.nodeMap |= nodeBit

			// insert newnode into the i'th spot of nt.nodes[]
			nt.nodes = append(nt.nodes[:i], append([]nodeI{nn}, nt.nodes[i:]...)...)

			if BitCount64(nt.nodeMap) >= TABLE_CAPACITY/2 {
				// promote compressedTable to fullTable
				return NewFullTable(nt.depth(), nt.hashPath, nt.entries())
			}
		} else /* if (t.nodeMap & nodeBit) > 0 */ {
			// don't need to touch nt.nodeMap
			nt.nodes[i] = nn //overwrite i'th slice entry
		}
	} else /* if nn == nil */ {
		if (t.nodeMap & nodeBit) > 0 {

			nt.nodeMap &^= nodeBit //unset nodeBit via bitClear &^ op
			nt.nodes = append(nt.nodes[:i], nt.nodes[i+1:]...)

			if nt.nodeMap == 0 {
				return nil
			}
		} else if (t.nodeMap & nodeBit) == 0 {
			// do nothing
			return t
		}
	}

	return nt
}

func (t compressedTable) del(hash uint64) (tableI, nodeI) {
	return nil, nil
}

type fullTable struct {
	hashPath uint64 // depth*NBITS of hash to get to this location in the Trie
	depth_   uint
	nodeMap  uint64
	nodes    [TABLE_CAPACITY]nodeI
}

func NewFullTable(depth uint, hashPath uint64, tabEnts []tableEntry) tableI {
	var ft = new(fullTable)
	ft.hashPath = hashPath
	ft._depth = depth
	//ft.nodeMap = 0 //unnecessary

	for _, ent := range tabEnts {
		var nodeBit = uint64(1 << ent.idx)
		ft.nodeMap |= nodeBit
		ft.nodes[ent.idx] = ent.node
	}

	return ft
}

func (t fullTable) depth() uint {
	return t._depth
}

// hashcode() is required for nodeI
func (t fullTable) hashcode() uint64 {
	return t.hashPath
}

// copy() is required for nodeI
func (t fullTable) copy() *fullTable {
	var nt = new(fullTable)
	nt.hashPath = t.hashPath
	nt._depth = t._depth
	nt.nodeMap = t.nodeMap
	for i := 0; i < len(t.nodes); i++ {
		nt.nodes[i] = t.nodes[i]
	}
	return nt
}

// String() is required for nodeI
func (t fullTable) String() string {
	// fullTable{depth:%d hashPath:%s, nentries()=%d}
	return fmt.Sprintf("fullTable{depth:%d, hashPath:%s, nentries()=%d}", t.depth(), hashPathString(t.hashPath, t.depth()), t.nentries())
}

// LongString() is required for nodeI
func (t fullTable) LongString(indent string) string {
	var strs = make([]string, 3+len(t.nodes))

	strs[0] = indent + fmt.Sprintf("fullTable{depth:%d, hashPath:%s, nentries()=%d", t.depth(), hashPathString(t.hashPath, t.depth()), t.nentries())

	strs[1] = indent + "\tnodeMap=" + nodeMapString(t.nodeMap) + ","

	for i, n := range t.nodes {
		if t.nodes[i] == nil {
			strs[2+i] = indent + fmt.Sprintf("\tt.nodes[%d]: nil", i)
		} else {
			if tab, ok := t.nodes[i].(tableI); ok {
				strs[2+i] = indent + fmt.Sprintf("\tt.nodes[%d]:\n%s", i, tab.LongString(indent+"\t"))
			} else {
				strs[2+i] = indent + fmt.Sprintf("\tt.nodes[%d]: %s", i, n)
			}
		}
	}

	strs[len(strs)-1] = indent + "}"

	return strings.Join(strs, "\n")
}

// nentries() is required for tableI
func (t fullTable) nentries() uint {
	return BitCount64(t.nodeMap)
}

// entries() is required for tableI
//
// This function MUST return the slice of tableEntry structs from lowest
// tableEntry.idx to highest tableEntry.idx .
func (t fullTable) entries() []tableEntry {
	var n = t.nentries()
	var ents = make([]tableEntry, n)
	for i, j := uint(0), 0; i < TABLE_CAPACITY; i++ {
		var nodeBit = uint64(1 << i)
		if (t.nodeMap & nodeBit) > 0 {
			//The difference with compressedTable is t.nodes[i] vs. t.nodes[j]
			ents[j] = tableEntry{i, t.nodes[i]}
			j++
		}
	}
	return ents
}

// get(uint64) is required for tableI
func (t fullTable) get(hashPath uint64) nodeI {
	var idx = index(hashPath, t.depth())
	var nodeBit = uint64(1 << idx)

	if (t.nodeMap & nodeBit) == 0 {
		return nil
	}

	var node = t.nodes[idx]

	return node
}

// set(uint64, nodeI) is required for tableI
func (t fullTable) set(hashPath uint64, nn nodeI) tableI {
	var nt = t.copy()

	var idx = index(hashPath, nt.depth())
	var nodeBit = uint64(1 << idx)

	if nn != nil {
		nt.nodeMap |= nodeBit
		nt.nodes[idx] = nn
	} else /* if nn == nil */ {
		nt.nodeMap &^= nodeBit
		nt.nodes[idx] = nn

		if BitCount64(nt.nodeMap) < TABLE_CAPACITY/2 {
			return DowngradeToCompressedTable(nt.depth(), nt.hashPath, nt.entries())
		}
	}

	return nt
}

// del(uint64) is required for tableI
func (t fullTable) del(hash uint64) (tableI, nodeI) {
	return nil, nil
}

type leafI interface {
	nodeI
	get(key []byte) (interface{}, bool)
	put(key []byte, val interface{}) (leafI, bool) //bool == replace? val
	del(key []byte) (leafI, interface{}, bool)     //bool == deleted? key
}

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

func (l flatLeaf) LongString(indent string) string {
	return indent + l.String()
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

type keyVal struct {
	key []byte
	val interface{}
}

func (kv keyVal) String() string {
	return fmt.Sprintf("{[]byte(\"%s\"), val:%v}", kv.key, kv.val)
}

type collisionLeaf struct {
	hash60 uint64 //hash60(key)
	kvs    []keyVal
}

func NewCollisionLeaf(hash uint64, kvs []keyVal) *collisionLeaf {
	leaf := new(collisionLeaf)
	leaf.hash60 = hash & mask60
	leaf.kvs = append(leaf.kvs, kvs...)

	return leaf
}

func (l collisionLeaf) hashcode() uint64 {
	return l.hash60
}

//is this needed?
func (l collisionLeaf) copy() *collisionLeaf {
	var nl = new(collisionLeaf)
	nl.hash60 = l.hash60
	nl.kvs = append(nl.kvs, l.kvs...)
	return nl
}

func (l collisionLeaf) String() string {
	var kvstrs = make([]string, len(l.kvs))
	for i := 0; i < len(l.kvs); i++ {
		kvstrs[i] = l.kvs[i].String()
	}
	var jkvstr = strings.Join(kvstrs, ",")

	return fmt.Sprintf("{hash60:%s, kvs:[]kv{%s}}", hash60String(l.hash60), jkvstr)
}

func (l collisionLeaf) LongString(indent string) string {
	return indent + l.String()
}

func (l collisionLeaf) get(key []byte) (interface{}, bool) {
	for i := 0; i < len(l.kvs); i++ {
		if byteSlicesEqual(l.kvs[i].key, key) {
			return l.kvs[i].val, true
		}
	}
	return nil, false
}

func (l collisionLeaf) put(key []byte, val interface{}) (leafI, bool) {
	nl := new(collisionLeaf)
	nl.hash60 = l.hash60
	nl.kvs = append(nl.kvs, l.kvs...)

	for i, kv := range l.kvs {
		if byteSlicesEqual(kv.key, key) {
			nl.kvs[i].val = val
			return nl, true // val was replaced
		}
	}

	nl.kvs = append(nl.kvs, keyVal{key, val})
	return nl, false
}

func (l collisionLeaf) del(key []byte) (leafI, interface{}, bool) {
	if len(l.kvs) == 2 {
		if byteSlicesEqual(key, l.kvs[0].key) {
			return NewFlatLeaf(l.hash60, l.kvs[1].key, l.kvs[1].val), l.kvs[0].val, true
		}
		if byteSlicesEqual(key, l.kvs[1].key) {
			return NewFlatLeaf(l.hash60, l.kvs[0].key, l.kvs[0].val), l.kvs[1].val, true
		}
		return nil, nil, false
	}

	var nl = l.copy()

	for i := 0; i < len(nl.kvs); i++ {
		if byteSlicesEqual(key, nl.kvs[i].key) {
			var retVal = nl.kvs[i].val

			// removing the i'th element of a slice; wiki/SliceTricks "Delete"
			nl.kvs = append(nl.kvs[:i], nl.kvs[i+1:]...)

			return nl, retVal, true
		}
	}

	return nil, nil, false
}

//byteSlicesEqual function
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
