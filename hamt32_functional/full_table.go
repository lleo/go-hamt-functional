package hamt32_functional

import (
	"fmt"
	"strings"
)

type fullTable struct {
	hashPath uint32 // depth*NBITS32 of hash to get to this location in the Trie
	nodeMap  uint32
	nodes    [TABLE_CAPACITY]nodeI
}

func UpgradeToFullTable(hashPath uint32, tabEnts []tableEntry) tableI {
	var ft = new(fullTable)
	ft.hashPath = hashPath
	//ft.nodeMap = 0 //unnecessary

	for _, ent := range tabEnts {
		var nodeBit = uint32(1 << ent.idx)
		ft.nodeMap |= nodeBit
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
	nt.nodeMap = t.nodeMap
	for i := 0; i < len(t.nodes); i++ {
		nt.nodes[i] = t.nodes[i]
	}
	return nt
}

// String() is required for nodeI
func (t fullTable) String() string {
	// fullTable{hashPath:/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d, nentries:%d,}
	return fmt.Sprintf("fullTable{hashPath:%s, nentries()=%d}", Hash30String(t.hashPath), t.nentries())
}

func (t fullTable) toString(depth uint) string {
	return fmt.Sprintf("fullTable{hashPath:%s, nentries()=%d}", hashPathString(t.hashPath, depth), t.nentries())
}

// LongString() is required for tableI
func (t fullTable) LongString(indent string, depth uint) string {
	var strs = make([]string, 3+len(t.nodes))

	strs[0] = indent + fmt.Sprintf("fullTable{hashPath:%s, nentries()=%d", hashPathString(t.hashPath, depth), t.nentries())

	strs[1] = indent + "\tnodeMap=" + nodeMapString(t.nodeMap) + ","

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
	return BitCount32(t.nodeMap)
}

// This function MUST return the slice of tableEntry structs from lowest
// tableEntry.idx to highest tableEntry.idx .
func (t fullTable) entries() []tableEntry {
	var n = t.nentries()
	var ents = make([]tableEntry, n)
	for i, j := uint(0), 0; i < TABLE_CAPACITY; i++ {
		var nodeBit = uint32(1 << i)
		if (t.nodeMap & nodeBit) > 0 {
			//The difference with compressedTable is t.nodes[i] vs. t.nodes[j]
			ents[j] = tableEntry{i, t.nodes[i]}
			j++
		}
	}
	return ents
}

// get(uint32) is required for tableI
func (t fullTable) get(idx uint) nodeI {
	var nodeBit = uint32(1 << idx)

	if (t.nodeMap & nodeBit) == 0 {
		return nil
	}

	var node = t.nodes[idx]

	return node
}

// set(uint32, nodeI) is required for tableI
func (t fullTable) set(idx uint, nn nodeI) tableI {
	var nt = t.copy()

	var nodeBit = uint32(1 << idx)

	if nn != nil {
		nt.nodeMap |= nodeBit
		nt.nodes[idx] = nn
	} else /* if nn == nil */ {
		nt.nodeMap &^= nodeBit
		nt.nodes[idx] = nn

		if nt.nodeMap == 0 {
			return nil
		}

		if BitCount32(nt.nodeMap) < TABLE_CAPACITY/2 {
			return DowngradeToCompressedTable(nt.hashPath, nt.entries())
		}

	}

	return nt
}
