package hamt64

import (
	"fmt"
	"log"
	"strings"
)

type fullTable struct {
	hashPath uint64 // depth*Nbits of hash to get to this location in the Trie
	numEnts  uint
	nodes    [TableCapacity]nodeI
}

func createRootFullTable(leaf leafI) tableI {
	var idx = index(leaf.Hash60(), 0)

	var ft = new(fullTable)
	//ft.hashPath = 0
	ft.numEnts = 1
	ft.nodes[idx] = leaf

	return ft
}

func createFullTable(depth uint, leaf1 leafI, leaf2 flatLeaf) tableI {
	var retTable = new(fullTable)
	retTable.hashPath = leaf1.Hash60() & hashPathMask(depth)

	var curTable = retTable
	var hashPath = retTable.hashPath
	var d uint
	for d = depth; d < MaxDepth; d++ {
		var idx1 = index(leaf1.Hash60(), d)
		var idx2 = index(leaf2.Hash60(), d)

		if idx1 != idx2 {
			curTable.nodes[idx1] = leaf1
			curTable.nodes[idx2] = leaf2

			curTable.numEnts = 2

			break
		}
		// idx1 == idx2 && continue

		hashPath = buildHashPath(hashPath, idx1, d)

		var newTable = new(fullTable)
		newTable.hashPath = hashPath

		curTable.numEnts = 1
		curTable.nodes[idx1] = newTable

		curTable = newTable
	}
	// We either BREAK out of the loop,
	// OR we hit d == MaxDepth.
	if d == MaxDepth {
		var idx1 = index(leaf1.Hash60(), d)
		var idx2 = index(leaf2.Hash60(), d)

		if idx1 != idx2 {
			curTable.nodes[idx1] = leaf1
			curTable.nodes[idx2] = leaf2

			curTable.numEnts = 2

			return retTable
		}
		// idx1 == idx2

		// NOTE: This condition should never result. The condition is
		// leaf1.Hash60() == leaf2.Hash60() all the way to MaxDepth;
		// because Hamt.newTable() is called only once, and after a
		// leaf1.Hash60() == leaf2.Hash60() check. It is here for completeness.
		log.Printf("full_table.go:newFullTable: SHOULD NOT BE CALLED")
		if leaf1.Hash60() != leaf2.Hash60() {
			log.Printf("madDepth=%d; d=%d; idx1=%d; idx2=%d", MaxDepth, d, idx1, idx2)
			log.Panicf("newFullTable: %s != %s", hash60String(leaf1.Hash60()), hash60String(leaf2.Hash60()))
		}
		var newLeaf, _ = leaf1.put(leaf2.key, leaf2.val)
		curTable.insert(idx1, newLeaf)
	}

	return retTable
}

func upgradeToFullTable(hashPath uint64, tabEnts []tableEntry) tableI {
	var ft = new(fullTable)
	ft.hashPath = hashPath
	ft.numEnts = uint(len(tabEnts))

	for _, ent := range tabEnts {
		ft.nodes[ent.idx] = ent.node
	}

	return ft
}

// Hash60() is required for nodeI
func (t fullTable) Hash60() uint64 {
	return t.hashPath
}

// copy() is required for nodeI
func (t fullTable) copy() *fullTable {
	var nt = new(fullTable)
	nt.hashPath = t.hashPath
	nt.numEnts = t.numEnts
	//for i := 0; i < len(t.nodes); i++ {
	//	nt.nodes[i] = t.nodes[i]
	//}
	nt.nodes = t.nodes

	return nt
}

// String() is required for nodeI
func (t fullTable) String() string {
	// fullTable{hashPath:/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d, nentries:%d,}
	return fmt.Sprintf("fullTable{hashPath:%s, nentries()=%d}", hash60String(t.hashPath), t.nentries())
}

// LongString() is required for tableI
func (t fullTable) LongString(indent string, depth uint) string {
	var strs = make([]string, 2+len(t.nodes))

	strs[0] = indent + fmt.Sprintf("fullTable{hashPath:%s, nentries()=%d,", hashPathString(t.hashPath, depth), t.nentries())

	for i, n := range t.nodes {
		if t.nodes[i] == nil {
			strs[1+i] = indent + fmt.Sprintf("\tt.nodes[%d]: nil", i)
		} else {
			if t, ok := t.nodes[i].(tableI); ok {
				strs[1+i] = indent + fmt.Sprintf("\tt.nodes[%d]:\n%s", i, t.LongString(indent+"\t", depth+1))
			} else {
				strs[1+i] = indent + fmt.Sprintf("\tt.nodes[%d]: %s", i, n)
			}
		}
	}

	strs[len(strs)-1] = indent + "}"

	return strings.Join(strs, "\n")
}

// nentries() is required for tableI
func (t fullTable) nentries() uint {
	return t.numEnts
}

// This function MUST return the slice of tableEntry structs from lowest
// tableEntry.idx to highest tableEntry.idx .
func (t fullTable) entries() []tableEntry {
	var n = t.nentries()
	var ents = make([]tableEntry, n)
	for i, j := uint(0), 0; i < TableCapacity; i++ {
		if t.nodes[i] != nil {
			//The difference with compressedTable is t.nodes[i] vs. t.nodes[j]
			ents[j] = tableEntry{i, t.nodes[i]}
			j++
		}
	}
	return ents
}

// get(uint64) is required for tableI
func (t fullTable) get(idx uint) nodeI {
	return t.nodes[idx]
}

func (t fullTable) insert(idx uint, entry nodeI) tableI {
	// t.nodes[idx] == nil
	var nt = t.copy()
	nt.nodes[idx] = entry
	nt.numEnts++
	return nt
}

func (t fullTable) replace(idx uint, entry nodeI) tableI {
	// t.nodes[idx] != nil
	var nt = t.copy()
	nt.nodes[idx] = entry
	return nt
}

func (t fullTable) remove(idx uint) tableI {
	// t.nodes[idx] != nil
	var nt = t.copy()
	nt.nodes[idx] = nil
	nt.numEnts--

	if GradeTables && nt.numEnts < DowngradeThreshold {
		return downgradeToCompressedTable(nt.hashPath, nt.entries())
	}

	if nt.numEnts == 0 {
		return nil
	}

	return nt
}
