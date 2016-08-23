/*
 */
package hamt_functional

import (
	"fmt"

	"github.com/lleo/go-hamt-functional/hamt32_functional"
	"github.com/lleo/go-hamt-functional/hamt64_functional"
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

func NewHamt32() hamt32_functional.Hamt {
	return hamt32_functional.EMPTY
}

func NewHamt64() hamt64_functional.Hamt {
	return hamt64_functional.EMPTY
}
