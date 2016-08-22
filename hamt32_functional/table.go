package hamt32_functional

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

type tableEntry struct {
	idx  uint
	node nodeI
}

//POPCNT Implementation
// copied from https://github.com/jddixon/xlUtil_go/blob/master/popCount.go
//  was MIT License

const (
	OCTO_FIVES  = uint32(0x55555555)
	OCTO_THREES = uint32(0x33333333)
	OCTO_ONES   = uint32(0x01010101)
	OCTO_FS     = uint32(0x0f0f0f0f)
)

func BitCount32(n uint32) uint {
	n = n - ((n >> 1) & OCTO_FIVES)
	n = (n & OCTO_THREES) + ((n >> 2) & OCTO_THREES)
	return uint((((n + (n >> 4)) & OCTO_FS) * OCTO_ONES) >> 24)
}
