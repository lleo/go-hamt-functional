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
	octo_fives  = uint32(0x55555555)
	octo_threes = uint32(0x33333333)
	octo_ones   = uint32(0x01010101)
	octo_fs     = uint32(0x0f0f0f0f)
)

// The bitCount32() function is a software based implementation of the POPCNT
// instruction. It returns the number of bits set in a uint32 word.
//
// This is copied from https://github.com/jddixon/xlUtil_go/blob/master/popCount.go
func bitCount32(n uint32) uint {
	n = n - ((n >> 1) & octo_fives)
	n = (n & octo_threes) + ((n >> 2) & octo_threes)
	return uint((((n + (n >> 4)) & octo_fs) * octo_ones) >> 24)
}
