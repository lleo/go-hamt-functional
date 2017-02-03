package hamt64

import (
	"fmt"
	"log"
	"strings"
)

// The compressedTable is a low memory usage version of a fullTable. It applies
// to tables with less than TableCapacity/2 number of entries in the table.
//
// It records which table entries are populated using a bit map called nodeMap.
//
// It stores the nodes in a go slice starting with the node corresponding to
// the Least Significant Bit(LSB) is the first node in the slice. While precise
// and accurate this discription does not help boring regular programmers. Most
// bit patterns are drawn from the Most Significant Bits(MSB) to the LSB; in
// orther words for a uint64 from the 31st bit to the 0th bit left to right. So
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
// that each 64 entry table is indexed by 6bit section (2^==64) of the key hash,
// there is a function to calculate the index called index(hash, depth);
//
type compressedTable struct {
	hashPath uint64 // depth*Nbits of hash to get to this location in the Trie
	nodeMap  uint64
	nodes    []nodeI
	grade    bool
}

func newRootCompressedTable(grade bool, lf leafI) tableI {
	var idx = index(lf.Hash60(), 0)

	var ct = new(compressedTable)
	ct.grade = grade
	//ct.hashPath = 0
	ct.nodeMap = 1 << idx
	ct.nodes = make([]nodeI, 1)
	ct.nodes[0] = lf

	return ct
}

func newCompressedTable(grade bool, depth uint, leaf1 leafI, leaf2 flatLeaf) tableI {
	var retTable = new(compressedTable)
	retTable.grade = grade
	retTable.hashPath = leaf1.Hash60() & hashPathMask(depth)

	var curTable = retTable
	var hashPath = retTable.hashPath
	var d uint
	for d = depth; d < MaxDepth; d++ {
		var idx1 = index(leaf1.Hash60(), d)
		var idx2 = index(leaf2.Hash60(), d)

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
		// idx1 == idx2 && loop

		curTable.nodes = make([]nodeI, 1)

		hashPath = buildHashPath(hashPath, idx1, d)

		var newTable = new(compressedTable)
		newTable.grade = grade
		newTable.hashPath = hashPath

		curTable.nodeMap = 1 << idx1 //Set the idx1'th bit
		curTable.nodes[0] = newTable

		curTable = newTable
	}
	// We either BREAK out of the loop,
	// OR we hit d == MaxDepth.
	if d == MaxDepth {
		var idx1 = index(leaf1.Hash60(), d)
		var idx2 = index(leaf2.Hash60(), d)

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

			return retTable
		}
		// idx1 == idx2

		// NOTE: This condition should never result. The condition is
		// leaf1.Hash60() == leaf2.Hash60() all the way to MaxDepth;
		// because Hamt.newTable() is called only once, and after a
		// leaf1.Hash60() == leaf2.Hash60() check. It is here for completeness.
		log.Printf("compressed_table.go:newCompressedTable: SHOULD NOT BE CALLED")
		if leaf1.Hash60() != leaf2.Hash60() {
			log.Printf("madDepth=%d; d=%d; idx1=%d; idx2=%d", MaxDepth, d, idx1, idx2)
			log.Panicf("newCompressedTable: %s != %s", hash60String(leaf1.Hash60()), hash60String(leaf2.Hash60()))
		}
		var newLeaf, _ = leaf1.put(leaf2.key, leaf2.val)
		curTable.nodes = make([]nodeI, 1)
		curTable.nodeMap |= 1 << idx1
		curTable.nodes[0] = newLeaf
	}

	return retTable
}

// downgradeToCompressedTable() converts fullTable structs that have less than
// TableCapacity/2 tableEntry's. One important thing we know is that none of
// the entries will collide with another.
//
// The ents []tableEntry slice is guaranteed to be in order from lowest idx to
// highest. tableI.entries() also adhears to this contract.
func downgradeToCompressedTable(hashPath uint64, ents []tableEntry) *compressedTable {
	var nt = new(compressedTable)
	nt.grade = true
	nt.hashPath = hashPath
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

func (t compressedTable) Hash60() uint64 {
	return t.hashPath
}

func (t compressedTable) copy() *compressedTable {
	var nt = new(compressedTable)
	nt.grade = t.grade
	nt.hashPath = t.hashPath
	nt.nodeMap = t.nodeMap
	nt.nodes = append(nt.nodes, t.nodes...)
	return nt
}

func nodeMapString(nodeMap uint64) string {
	var strs = make([]string, 4)

	var top2 = nodeMap >> 60
	strs[0] = fmt.Sprintf("%02b", top2)

	const tenBitMask uint64 = 1<<10 - 1
	for i := uint(0); i < 3; i++ {
		tenBitVal := (nodeMap & (tenBitMask << (i * 10))) >> (i * 10)
		strs[3-i] = fmt.Sprintf("%010b", tenBitVal)
	}

	return strings.Join(strs, " ")
}

//String() is required for nodeI depth
func (t compressedTable) String() string {
	// compressedTale{hashPath:/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d, nentries:%d,}
	return fmt.Sprintf("compressedTable{hashPath:%s, nentries()=%d}",
		hash60String(t.hashPath), t.nentries())
}

// LongString() is required for tableI
func (t compressedTable) LongString(indent string, depth uint) string {
	var strs = make([]string, 2+len(t.nodes))

	strs[0] = indent + fmt.Sprintf("compressedTable{hashPath=%s, nentries()=%d, nodeMap=%s,", hashPathString(t.hashPath, depth), t.nentries(), nodeMapString(t.nodeMap))

	for i, n := range t.nodes {
		if t, ok := n.(tableI); ok {
			strs[1+i] = indent + fmt.Sprintf("\tt.nodes[%d]:\n%s", i, t.LongString(indent+"\t", depth+1))
		} else {
			strs[1+i] = indent + fmt.Sprintf("\tt.nodes[%d]: %s", i, n.String())
		}
	}

	strs[len(strs)-1] = indent + "}"

	return strings.Join(strs, "\n")
}

func (t compressedTable) nentries() uint {
	//return bitCount64(t.nodeMap)
	return uint(len(t.nodes))
}

// This function MUST return the slice of tableEntry structs from lowest
// tableEntry.idx to highest tableEntry.idx .
func (t compressedTable) entries() []tableEntry {
	var n = t.nentries()
	var ents = make([]tableEntry, n)

	for i, j := uint(0), uint(0); i < TableCapacity; i++ {
		var nodeBit = uint64(1 << i)

		if (t.nodeMap & nodeBit) > 0 {
			ents[j] = tableEntry{i, t.nodes[j]}
			j++
		}
	}

	return ents
}

func (t compressedTable) get(idx uint) nodeI {
	var nodeBit = uint64(1 << idx)

	if (t.nodeMap & nodeBit) == 0 {
		return nil
	}

	// Create a mask to mask off all bits below idx'th bit
	var m = uint64(1<<idx) - 1

	// Count the number of bits in the nodeMap below the idx'th bit
	var i = bitCount64(t.nodeMap & m)

	var node = t.nodes[i]

	return node
}

func (t compressedTable) insert(idx uint, entry nodeI) tableI {
	// t.nodeMap & 1<<idx == 0
	var nodeBit = uint64(1 << idx)
	var bitMask = nodeBit - 1
	var i = bitCount64(t.nodeMap & bitMask)

	var nt = t.copy()
	nt.nodeMap |= nodeBit

	// insert newnode into the i'th spot of nt.nodes[]
	nt.nodes = append(nt.nodes[:i], append([]nodeI{entry}, nt.nodes[i:]...)...)

	if t.grade && uint(len(nt.nodes)) >= UpgradeThreshold {
		// promote compressedTable to fullTable
		return upgradeToFullTable(nt.hashPath, nt.entries())
	}

	return nt
}

func (t compressedTable) replace(idx uint, entry nodeI) tableI {
	// t.nodeMap & 1<<idx > 0
	var nodeBit = uint64(1 << idx)
	var bitMask = nodeBit - 1
	var i = bitCount64(t.nodeMap & bitMask)

	var nt = t.copy()

	nt.nodes[i] = entry

	return nt
}

func (t compressedTable) remove(idx uint) tableI {
	// t.nodeMap & 1<<idx > 0
	var nodeBit = uint64(1 << idx)
	var bitMask = nodeBit - 1
	var i = bitCount64(t.nodeMap & bitMask)

	var nt = t.copy()

	nt.nodeMap &^= nodeBit
	nt.nodes = append(nt.nodes[:i], nt.nodes[i+1:]...)

	if nt.nodeMap == 0 {
		return nil
	}

	return nt
}

//POPCNT Implementation
// copied from https://github.com/jddixon/xlUtil_go/blob/master/popCount.go
//  was MIT License

const (
	hexiFives  = uint64(0x5555555555555555)
	hexiThrees = uint64(0x3333333333333333)
	hexiOnes   = uint64(0x0101010101010101)
	hexiFs     = uint64(0x0f0f0f0f0f0f0f0f)
)

// The bitCount64() function is a software based implementation of the POPCNT
// instruction. It returns the number of bits set in a uint64 word.
//
// This is copied from https://github.com/jddixon/xlUtil_go/blob/master/popCount.go
func bitCount64(n uint64) uint {
	n = n - ((n >> 1) & hexiFives)
	n = (n & hexiThrees) + ((n >> 2) & hexiThrees)
	return uint((((n + (n >> 4)) & hexiFs) * hexiOnes) >> 56)
}
