package hamt64

import (
	"fmt"
	"log"
	"strings"

	"github.com/lleo/go-hamt-key"
)

type fullTable struct {
	hashPath key.HashVal60 // depth*nBits of hash to get to this location in the Trie
	depth    uint
	numEnts  uint
	nodes    [TableCapacity]nodeI
}

func createRootFullTable(leaf leafI) tableI {
	var idx = leaf.Hash60().Index(0)

	var ft = new(fullTable)
	//ft.hashPath = 0
	//ft.depth = 0
	ft.numEnts = 1
	ft.nodes[idx] = leaf

	return ft
}

func createFullTable(depth uint, leaf1 leafI, leaf2 flatLeaf) tableI {
	var retTable = new(fullTable)
	retTable.hashPath = leaf1.Hash60() & key.HashPathMask60(depth-1)
	retTable.depth = depth

	var curTable = retTable
	var hashPath = retTable.hashPath
	var d uint
	for d = depth; d < MaxDepth; d++ {
		var idx1 = leaf1.Hash60().Index(d)
		var idx2 = leaf2.Hash60().Index(d)

		if idx1 != idx2 {
			curTable.nodes[idx1] = leaf1
			curTable.nodes[idx2] = leaf2

			curTable.numEnts = 2

			break
		}
		// idx1 == idx2 && continue

		//hashPath = hashPath.BuildHashPath(idx1, d)
		hashPath = leaf1.Hash60() & key.HashPathMask60(d)

		var newTable = new(fullTable)
		newTable.hashPath = hashPath
		newTable.depth = d + 1

		curTable.numEnts = 1
		curTable.nodes[idx1] = newTable

		curTable = newTable
	}
	// We either BREAK out of the loop,
	// OR we hit d == MaxDepth.
	if d == MaxDepth {
		var idx1 = leaf1.Hash60().Index(d)
		var idx2 = leaf2.Hash60().Index(d)

		if idx1 != idx2 {
			curTable.nodes[idx1] = leaf1
			curTable.nodes[idx2] = leaf2

			curTable.numEnts = 2

			return retTable
		}
		// idx1 == idx2

		// NOTE: This condition should never result. The condition is
		// leaf1.Hash60() == leaf2.Hash60() all the way to MaxDepth;
		// because Hamt.createTable() is called only once, and after a
		// leaf1.Hash60() == leaf2.Hash60() check. It is here for completeness.
		log.Printf("full_table.go:createFullTable: SHOULD NOT BE CALLED")

		// Check if the path of leaf1 is not equal to the one leaf2 just traversed.
		if leaf1.Hash60() != leaf2.Hash60() {
			log.Printf("MaxDepth=%d; d=%d; idx1=%d; idx2=%d", MaxDepth, d, idx1, idx2)
			log.Panicf("createFullTable: %s,0x%06x != %s,0x%06x",
				leaf1.Hash60(), leaf1.Hash60(), leaf2.Hash60(), leaf2.Hash60())
		}

		// Just for completeness; leaf1.Hash60() == leaf2.hash60()
		var newLeaf, _ = leaf1.put(leaf2.key, leaf2.val)
		curTable.nodes[idx1] = newLeaf
	}

	return retTable
}

func upgradeToFullTable(hashPath key.HashVal60, depth uint, tabEnts []tableEntry) tableI {
	var ft = new(fullTable)
	ft.hashPath = hashPath
	ft.depth = depth
	ft.numEnts = uint(len(tabEnts))

	for _, ent := range tabEnts {
		ft.nodes[ent.idx] = ent.node
	}

	return ft
}

// Hash60() is required for nodeI
func (t fullTable) Hash60() key.HashVal60 {
	return t.hashPath
}

// copy() is required for nodeI
func (t fullTable) copy() *fullTable {
	var nt = new(fullTable)
	nt.hashPath = t.hashPath
	nt.depth = t.depth
	nt.numEnts = t.numEnts
	//for i := 0; i < len(t.nodes); i++ {
	//	nt.nodes[i] = t.nodes[i]
	//}
	nt.nodes = t.nodes

	return nt
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

// Get() is required for tableI
func (t fullTable) Get(idx uint) nodeI {
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

//func (t fullTable) remove(idx uint) nodeI {
func (t fullTable) remove(idx uint) tableI {
	// t.nodes[idx] != nil
	var nt = t.copy()
	nt.nodes[idx] = nil
	nt.numEnts--

	if GradeTables && nt.numEnts < DowngradeThreshold {
		return downgradeToCompressedTable(nt.hashPath, nt.depth, nt.entries())
	}

	if nt.numEnts == 0 {
		return nil
	}

	return nt
}

// String() is required for nodeI
func (t fullTable) String() string {
	// fullTable{hashPath:/%d/%d/%d/%d/%d/%d, nentries:%d,}
	return fmt.Sprintf("fullTable{hashPath:%s, nentries()=%d, depth=%d}", t.hashPath.HashPathString(t.depth), t.nentries(), t.depth)
}

// LongString() is required for tableI
func (t fullTable) LongString(indent string, recurse bool) string {
	//var strs = make([]string, 2+len(t.nodes))
	var strs = make([]string, 2+t.nentries())

	strs[0] = indent + fmt.Sprintf("fullTable{hashPath:%s, nentries()=%d, t.depth=%d,", t.hashPath.HashPathString(t.depth), t.nentries(), t.depth)

	var j int
	for i, n := range t.nodes {
		//if n == nil {
		//	strs[1+i] = indent + fmt.Sprintf(halfIndent+"t.nodes[%d]: nil", i)
		//} else {
		if n != nil {
			if tt, ok := n.(tableI); ok {
				if recurse {
					strs[1+j] = indent + fmt.Sprintf(halfIndent+"t.nodes[%d]:\n%s", i, tt.LongString(indent+fullIndent, recurse))
				} else {
					strs[1+j] = indent + fmt.Sprintf(halfIndent+"t.nodes[%d]: %s", i, tt.String())
				}
			} else {
				strs[1+j] = indent + fmt.Sprintf(halfIndent+"t.nodes[%d]: %s", i, n)
			}
			j++
		}
	}

	strs[len(strs)-1] = indent + "}"

	return strings.Join(strs, "\n")
}
