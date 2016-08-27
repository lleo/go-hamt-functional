/*
 */
package hamt

import (
	"fmt"

	"github.com/lleo/go-hamt-functional/hamt32"
	"github.com/lleo/go-hamt-functional/hamt64"
	"github.com/lleo/go-hamt/key"
)

//type HamtFunctional interface {
//	Get(key.Key) (interface{}, bool)
//	Put(key.Key, interface{}) (HamtFunctional, bool)
//	Del(key.Key) (HamtFunctional, interface{}, bool)
//	IsEmpty() bool
//	String() string
//	LongString(indent string) string
//}

type keyVal struct {
	key key.Key
	val interface{}
}

func (kv keyVal) String() string {
	return fmt.Sprintf("keyVal{%s, %v}", kv.key, kv.val)
}

func NewHamt32() hamt32.Hamt {
	return hamt32.EMPTY
}

func NewHamt64() hamt64.Hamt {
	return hamt64.EMPTY
}
