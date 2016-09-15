package hamt32

// nodeI is the interface for every entry in a table; so table entries are
// either a leaf or a table or nil.
//
// The nodeI interface can be for compressedTable, fullTable, flatLeaf, or
// collisionLeaf.
//
// The tableI interface is for compressedTable and fullTable.
//
// The hashcode() method for leaf structs is the 30 most significant bits of
// the keys hash.
//
// The hashcode() method for table structs is the depth*NBITS32 of the hash path
// that leads to the table's position in the Trie.
//
// For leafs hashcode() is the 30 bits returned by hash30(key).
// For collisionLeafs this is the definition of what a collision is.
//
type nodeI interface {
	hashcode() uint32
	String() string
}

// Every tableI is a nodeI.
//
type tableI interface {
	nodeI

	toString(depth uint) string
	LongString(indent string, depth uint) string

	nentries() uint // get the number of node entries

	// Get an Ordered list of index and node pairs. This slice MUST BE Ordered
	// from lowest index to highest.
	entries() []tableEntry

	get(idx uint) nodeI               // get an entry
	set(idx uint, entry nodeI) tableI // set and entry
}

func newTable(depth uint, hashPath uint32, leaf leafI) tableI {
	if !GRADE_TABLES && TABLE_TYPE == "full" {
		return newFullTable(depth, hashPath, leaf)
	}
	return newCompressedTable(depth, hashPath, leaf)
}

func newTable2(depth uint, hashPath uint32, leaf1 leafI, leaf2 flatLeaf) tableI {
	if !GRADE_TABLES && TABLE_TYPE == "full" {
		return newFullTable2(depth, hashPath, leaf1, leaf2)
	}
	return newCompressedTable2(depth, hashPath, leaf1, leaf2)
}

type tableEntry struct {
	idx  uint
	node nodeI
}
