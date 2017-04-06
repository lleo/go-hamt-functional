package hamt64

//POPCNT Implementation
// copied from https://github.com/jddixon/xlUtil_go/blob/master/popCount.go
//  was MIT License

const (
	hexiFives  = uint64(0x5555555555555555)
	hexiThrees = uint64(0x3333333333333333)
	hexiOnes   = uint64(0x0101010101010101)
	hexiFs     = uint64(0x0f0f0f0f0f0f0f0f)
)

// The bitCount64() function is a software based implementation of the POPCNT
// instruction. It returns the number of bits set in a uint64 word.
//
// This is copied from https://github.com/jddixon/xlUtil_go/blob/master/popCount.go
func bitCount64(n uint64) uint {
	n = n - ((n >> 1) & hexiFives)
	n = (n & hexiThrees) + ((n >> 2) & hexiThrees)
	return uint((((n + (n >> 4)) & hexiFs) * hexiOnes) >> 56)
}
