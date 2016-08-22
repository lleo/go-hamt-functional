package hamt_functional

import (
	"fmt"

	hamt32 "github.com/lleo/go-hamt-functional/hamt32_functional"
	hamt64 "github.com/lleo/go-hamt-functional/hamt64_functional"
	"github.com/lleo/go-hamt/hamt_key"
)

//type HamtFunctional interface {
//	Get(hamt_key.Key) (interface{}, bool)
//	Put(hamt_key.Key, interface{}) (HamtFunctional, bool)
//	Del(hamt_key.Key) (HamtFunctional, interface{}, bool)
//	IsEmpty() bool
//	String() string
//	LongString(indent string) string
//}

type keyVal struct {
	key hamt_key.Key
	val interface{}
}

func (kv keyVal) String() string {
	return fmt.Sprintf("keyVal{%s, %v}", kv.key, kv.val)
}

func NewHamt64() hamt64.Hamt {
	return hamt64.EMPTY
}

func NewHamt32() hamt32.Hamt {
	return hamt32.EMPTY
}
