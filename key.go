package hamt_functional

type HamtKey interface {
	Equals(HamtKey) bool
	Hash64() uint64
	String() string
}
