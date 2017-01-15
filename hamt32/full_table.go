package hamt32

import (
	"fmt"
	"strings"
)

type fullTable struct {
	hashPath uint32 // depth*NBITS of hash to get to this location in the Trie
	numEnts  uint
	nodes    [TABLE_CAPACITY]nodeI
}

func newFullTable(depth uint, hashPath uint32, leaf leafI) tableI {
	var idx = index(hashPath, depth)

	var ft = new(fullTable)
	//var ft = pool.Get().(*fullTable)
	ft.hashPath = hashPath & hashPathMask(depth)
	ft.numEnts = 1
	ft.nodes[idx] = leaf

	return ft
}

func newFullTable2(depth uint, hashPath uint32, leaf1 leafI, leaf2 flatLeaf) tableI {
	var retTable = new(fullTable)
	//var retTable = pool.Get().(*fullTable)
	retTable.hashPath = hashPath & hashPathMask(depth)

	var curTable = retTable
	var d uint
	for d = depth; d <= MAXDEPTH; d++ {
		var idx1 = index(leaf1.hashcode(), d)
		var idx2 = index(leaf2.hashcode(), d)

		if idx1 != idx2 {
			curTable.nodes[idx1] = leaf1
			curTable.nodes[idx2] = leaf2

			curTable.numEnts = 2

			break
		}
		// idx1 == idx2 && continue

		hashPath = buildHashPath(hashPath, idx1, d)

		var newTable = new(fullTable)
		//var newTable = pool.Get().(*fullTable)
		newTable.hashPath = hashPath

		curTable.numEnts = 1
		curTable.nodes[idx1] = newTable

		curTable = newTable
	}
	// We either BREAK out of the loop,
	// OR we hit d > MAXDEPTH.
	if d > MAXDEPTH {
		// leaf1.hashcode() == leaf2.hashcode()
		var idx = index(leaf1.hashcode(), MAXDEPTH)
		leaf, _ := leaf1.put(leaf2.key, leaf2.val)
		curTable.set(idx, leaf)
	}

	return retTable
}

func upgradeToFullTable(hashPath uint32, tabEnts []tableEntry) tableI {
	var ft = new(fullTable)
	ft.hashPath = hashPath
	ft.numEnts = uint(len(tabEnts))

	for _, ent := range tabEnts {
		ft.nodes[ent.idx] = ent.node
	}

	return ft
}

// hashcode() is required for nodeI
func (t fullTable) hashcode() uint32 {
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
	return fmt.Sprintf("fullTable{hashPath:%s, nentries()=%d}", hash30String(t.hashPath), t.nentries())
}

func (t fullTable) toString(depth uint) string {
	return fmt.Sprintf("fullTable{hashPath:%s, nentries()=%d}", hashPathString(t.hashPath, depth), t.nentries())
}

// LongString() is required for tableI
func (t fullTable) LongString(indent string, depth uint) string {
	var strs = make([]string, 2+len(t.nodes))

	strs[0] = indent + fmt.Sprintf("fullTable{hashPath:%s, nentries()=%d", hashPathString(t.hashPath, depth), t.nentries())

	for i, n := range t.nodes {
		if t.nodes[i] == nil {
			strs[2+i] = indent + fmt.Sprintf("\tt.nodes[%d]: nil", i)
		} else {
			if t, ok := t.nodes[i].(tableI); ok {
				strs[2+i] = indent + fmt.Sprintf("\tt.nodes[%d]:\n%s", i, t.LongString(indent+"\t", depth+1))
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
	return t.numEnts
}

// This function MUST return the slice of tableEntry structs from lowest
// tableEntry.idx to highest tableEntry.idx .
func (t fullTable) entries() []tableEntry {
	var n = t.nentries()
	var ents = make([]tableEntry, n)
	for i, j := uint(0), 0; i < TABLE_CAPACITY; i++ {
		if t.nodes[i] != nil {
			//The difference with compressedTable is t.nodes[i] vs. t.nodes[j]
			ents[j] = tableEntry{i, t.nodes[i]}
			j++
		}
	}
	return ents
}

// get(uint32) is required for tableI
func (t fullTable) get(idx uint) nodeI {
	return t.nodes[idx]
}

// set(uint32, nodeI) is required for tableI
func (t fullTable) set(idx uint, nn nodeI) tableI {
	var nt = t.copy()

	var occupied = false
	if nt.nodes[idx] != nil {
		occupied = true
	}

	if nn != nil {
		nt.nodes[idx] = nn
		if !occupied {
			nt.numEnts += 1
		}
	} else /* if nn == nil */ {
		nt.nodes[idx] = nn

		if occupied {
			nt.numEnts -= 1
		}

		if nt.numEnts == 0 {
			return nil
		}

		if GRADE_TABLES && nt.numEnts < TABLE_CAPACITY/2 {
			return downgradeToCompressedTable(nt.hashPath, nt.entries())
		}
	}

	return nt
}
