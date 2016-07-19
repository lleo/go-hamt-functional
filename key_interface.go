package hamt

type HamtKey interface {
	Equals(HamtKey) bool
	Hash64() uint64
	String() string
}
