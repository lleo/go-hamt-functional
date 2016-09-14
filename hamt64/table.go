package hamt64

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

func newTable(depth uint, hashPath uint64, leaf leafI) tableI {
	if !GRADE_TABLES {
		return newFullTable(depth, hashPath, leaf)
	}
	return newCompressedTable(depth, hashPath, leaf)
}

func newTable2(depth uint, hashPath uint64, leaf1 leafI, leaf2 flatLeaf) tableI {
	if !GRADE_TABLES {
		return newFullTable2(depth, hashPath, leaf1, leaf2)
	}
	return newCompressedTable2(depth, hashPath, leaf1, leaf2)
}

type tableEntry struct {
	idx  uint
	node nodeI
}

//POPCNT Implementation
// copied from https://github.com/jddixon/xlUtil_go/blob/master/popCount.go
//  was MIT License

const (
	hexi_fives  = uint64(0x5555555555555555)
	hexi_threes = uint64(0x3333333333333333)
	hexi_ones   = uint64(0x0101010101010101)
	hexi_fs     = uint64(0x0f0f0f0f0f0f0f0f)
)

func bitCount64(n uint64) uint {
	n = n - ((n >> 1) & hexi_fives)
	n = (n & hexi_threes) + ((n >> 2) & hexi_threes)
	return uint((((n + (n >> 4)) & hexi_fs) * hexi_ones) >> 56)
}
