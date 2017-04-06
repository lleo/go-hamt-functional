package hamt32

//POPCNT Implementation
// copied from https://github.com/jddixon/xlUtil_go/blob/master/popCount.go
//  was MIT License

const (
	octoFives  = uint32(0x55555555)
	octoThrees = uint32(0x33333333)
	octoOnes   = uint32(0x01010101)
	octoFs     = uint32(0x0f0f0f0f)
)

// The bitCount32() function is a software based implementation of the POPCNT
// instruction. It returns the number of bits set in a uint32 word.
//
// This is copied from https://github.com/jddixon/xlUtil_go/blob/master/popCount.go
func bitCount32(n uint32) uint {
	n = n - ((n >> 1) & octoFives)
	n = (n & octoThrees) + ((n >> 2) & octoThrees)
	return uint((((n + (n >> 4)) & octoFs) * octoOnes) >> 24)
}
