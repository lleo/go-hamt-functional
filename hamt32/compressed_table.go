package hamt32

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
// orther words for a uint32 from the 31st bit to the 0th bit left to right. So
// for a 8bit number 1 is writtent as 00000001 (where the LSB is 1) and 128 is
// written as 10000000 (where the MSB is 1).
//
// So the number of entries in the node slice is equal to the number of bits set
// in the nodeMap. You can count the number of bits in the nodeMap, a 32bit word,
// by calculating the Hamming Weight (another obscure name; google it). The
// simple most generice way of calculating the Hamming Weight of a 32bit work is
// implemented in the BitCount32(uint32) function defined bellow.
//
// To figure out the index of a node in the nodes slice from the index of the bit
// in the nodeMap we first find out if that bit in the nodeMap is set by
// calculating if "nodeMap & (1<<idx) > 0" is true the idx'th bit is set. Given
// that each 32 entry table is indexed by 5bit section (2^5==32) of the key hash,
// there is a function to calculate the index called index(hash, depth);
//
type compressedTable struct {
	hashPath uint32 // depth*Nbits of hash to get to this location in the Trie
	depth    uint
	nodeMap  uint32
	nodes    []nodeI
}

func createRootCompressedTable(lf leafI) tableI {
	var idx = index(lf.Hash30(), 0)

	var ct = new(compressedTable)
	//ct.hashPath = 0
	//ct.depth = 0
	ct.nodeMap = 1 << idx
	ct.nodes = make([]nodeI, 1)
	ct.nodes[0] = lf

	return ct
}

func createCompressedTable(depth uint, leaf1 leafI, leaf2 flatLeaf) tableI {
	var retTable = new(compressedTable)
	retTable.hashPath = leaf1.Hash30() & hashPathMask(depth)
	retTable.depth = depth

	var curTable = retTable
	var hashPath = retTable.hashPath
	var d uint
	for d = depth; d < MaxDepth; d++ {
		var idx1 = index(leaf1.Hash30(), d)
		var idx2 = index(leaf2.Hash30(), d)

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

		hashPath = buildHashPath(hashPath, idx1, d+1)

		var newTable = new(compressedTable)
		newTable.hashPath = hashPath
		newTable.depth = d + 1

		curTable.nodeMap = 1 << idx1 //Set the idx1'th bit
		curTable.nodes[0] = newTable

		curTable = newTable
	}
	// We either BREAK out of the loop,
	// OR we hit d == MaxDepth.
	if d == MaxDepth {
		var idx1 = index(leaf1.Hash30(), d)
		var idx2 = index(leaf2.Hash30(), d)

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
		// leaf1.Hash30() == leaf2.Hash30() all the way to MaxDepth;
		// because Hamt.createTable() is called only once, and after a
		// leaf1.Hash30() == leaf2.Hash30() check. It is here for completeness.
		log.Printf("compressed_table.go:newCompressedTable: SHOULD NOT BE CALLED")

		// Check if the path of leaf1 is not equal to the one leaf2 just traversed.
		if leaf1.Hash30() != leaf2.Hash30() {
			log.Printf("madDepth=%d; d=%d; idx1=%d; idx2=%d", MaxDepth, d, idx1, idx2)
			log.Panicf("newCompressedTable: %s != %s", h30ToString(leaf1.Hash30()), h30ToString(leaf2.Hash30()))
		}

		// Just for completeness; leaf1.Hash30() == leaf2.hash30()
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
func downgradeToCompressedTable(hashPath uint32, depth uint, ents []tableEntry) *compressedTable {
	var nt = new(compressedTable)
	nt.hashPath = hashPath
	nt.depth = depth
	//nt.nodeMap = 0
	nt.nodes = make([]nodeI, len(ents))

	for i := 0; i < len(ents); i++ {
		var ent = ents[i]
		var nodeBit = uint32(1 << ent.idx)
		nt.nodeMap |= nodeBit
		nt.nodes[i] = ent.node
	}

	return nt
}

func (t compressedTable) Hash30() uint32 {
	return t.hashPath
}

func (t compressedTable) copyExceptNodes() *compressedTable {
	var nt = new(compressedTable)
	nt.hashPath = t.hashPath
	nt.depth = t.depth
	nt.nodeMap = t.nodeMap
	return nt
}

func (t compressedTable) copy() *compressedTable {
	var nt = t.copyExceptNodes()

	//nt.nodes = append(nt.nodes, t.nodes...)

	nt.nodes = make([]nodeI, len(t.nodes))
	copy(nt.nodes, t.nodes)

	return nt
}

func (t compressedTable) nentries() uint {
	//return bitCount32(t.nodeMap)
	return uint(len(t.nodes))
}

// This function MUST return the slice of tableEntry structs from lowest
// tableEntry.idx to highest tableEntry.idx .
func (t compressedTable) entries() []tableEntry {
	var n = t.nentries()
	var ents = make([]tableEntry, n)

	for i, j := uint(0), uint(0); i < TableCapacity; i++ {
		var nodeBit = uint32(1 << i)

		if (t.nodeMap & nodeBit) > 0 {
			ents[j] = tableEntry{i, t.nodes[j]}
			j++
		}
	}

	return ents
}

func (t compressedTable) Get(idx uint) nodeI {
	var nodeBit = uint32(1 << idx)

	if (t.nodeMap & nodeBit) == 0 {
		return nil
	}

	// Create a mask to mask off all bits below idx'th bit
	var m = uint32(1<<idx) - 1

	// Count the number of bits in the nodeMap below the idx'th bit
	var i = bitCount32(t.nodeMap & m)

	var node = t.nodes[i]

	return node
}

func (t compressedTable) insert(idx uint, entry nodeI) tableI {
	var nodeBit = uint32(1 << idx)
	var bitMask = nodeBit - 1
	var i = bitCount32(t.nodeMap & bitMask)

	var nt = t.copyExceptNodes()
	nt.nodeMap |= nodeBit

	// insert newnode into the i'th spot of nt.nodes[]

	// Slower append() way
	//nt.nodes = append(nt.nodes, t.nodes[:i]...)
	//nt.nodes = append(nt.nodes[:i], append([]nodeI{entry}, nt.nodes[i:]...)...)

	// Faster copy() way
	nt.nodes = make([]nodeI, len(t.nodes)+1)
	copy(nt.nodes, t.nodes[:i])
	nt.nodes[i] = entry
	copy(nt.nodes[i+1:], t.nodes[i:])

	if GradeTables && uint(len(nt.nodes)) >= UpgradeThreshold {
		// promote compressedTable to fullTable
		return upgradeToFullTable(nt.hashPath, nt.depth, nt.entries())
	}

	return nt
}

func (t compressedTable) replace(idx uint, entry nodeI) tableI {
	// t.nodeMap & 1<<idx > 0
	var nodeBit = uint32(1 << idx)
	var bitMask = nodeBit - 1
	var i = bitCount32(t.nodeMap & bitMask)

	var nt = t.copyExceptNodes()

	// Slower append() way
	//nt.nodes = append(nt.nodes, t.nodes...)

	// Faster copy() way
	nt.nodes = make([]nodeI, len(t.nodes))
	copy(nt.nodes, t.nodes)

	nt.nodes[i] = entry

	return nt
}

func (t compressedTable) remove(idx uint) tableI {
	var nodeBit = uint32(1 << idx)
	var bitMask = nodeBit - 1
	var i = bitCount32(t.nodeMap & bitMask)

	var nt = t.copyExceptNodes()

	nt.nodeMap &^= nodeBit

	// Slower append() way
	//nt.nodes = append(nt.nodes, t.nodes[:i]...)
	//nt.nodes = append(nt.nodes[:i], t.nodes[i+1:]...)

	// Faster copy() way
	nt.nodes = make([]nodeI, len(t.nodes)-1)
	copy(nt.nodes, t.nodes[:i])
	copy(nt.nodes[i:], t.nodes[i+1:])

	if nt.nodeMap == 0 {
		return nil
	}

	return nt
}

func nodeMapString(nodeMap uint32) string {
	var strs = make([]string, 4)

	var top2 = nodeMap >> 30
	strs[0] = fmt.Sprintf("%02b", top2)

	const tenBitMask uint32 = 1<<10 - 1
	for i := uint(0); i < 3; i++ {
		tenBitVal := (nodeMap & (tenBitMask << (i * 10))) >> (i * 10)
		strs[3-i] = fmt.Sprintf("%010b", tenBitVal)
	}

	return strings.Join(strs, " ")
}

//String() is required for nodeI depth
func (t compressedTable) String() string {
	// compressedTale{hashPath:/%d/%d/%d/%d/%d/%d, nentries:%d,}
	return fmt.Sprintf("compressedTable{hashPath:%s, nentries()=%d, depth=%d}",
		h30ToString(t.hashPath), t.nentries(), t.depth)
}

// LongString() is required for tableI
func (t compressedTable) LongString(indent string, recurse bool) string {
	var strs = make([]string, 2+len(t.nodes))

	strs[0] = indent + fmt.Sprintf("compressedTable{hashPath=%s, nentries()=%d, t.depth=%d, nodeMap=%s,", hashPathString(t.hashPath, t.depth), t.nentries(), t.depth, nodeMapString(t.nodeMap))

	for i, n := range t.nodes {
		if tt, ok := n.(tableI); ok {
			if recurse {
				strs[1+i] = indent + fmt.Sprintf(halfIndent+"t.nodes[%d]:\n%s", i, tt.LongString(indent+fullIndent, recurse))
			} else {
				strs[1+i] = indent + fmt.Sprintf(halfIndent+"t.nodes[%d]: %s", i, tt.String())
			}
		} else {
			strs[1+i] = indent + fmt.Sprintf(halfIndent+"t.nodes[%d]: %s", i, n.String())
		}
	}

	strs[len(strs)-1] = indent + "}"

	return strings.Join(strs, "\n")
}
