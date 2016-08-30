package hamt32

import (
	"fmt"
	"strings"
)

type fullTable struct {
	hashPath uint32 // depth*NBITS of hash to get to this location in the Trie
	nodes    [TABLE_CAPACITY]nodeI
	numEnts  uint
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
	for i := 0; i < len(t.nodes); i++ {
		nt.nodes[i] = t.nodes[i]
	}
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
	var node = t.nodes[idx]

	return node
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

		if nt.numEnts < TABLE_CAPACITY/2 {
			return downgradeToCompressedTable(nt.hashPath, nt.entries())
		}

	}

	return nt
}
