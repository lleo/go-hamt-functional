package hamt32_functional

import "github.com/lleo/go-hamt/hamt_key"

type leafI interface {
	nodeI
	get(key hamt_key.Key) (interface{}, bool)
	put(key hamt_key.Key, val interface{}) (leafI, bool) //bool == replace? val
	del(key hamt_key.Key) (leafI, interface{}, bool)     //bool == deleted? key
	keyVals() []keyVal
}

//type keyVal struct {
//	key hamt_key.Key
//	val interface{}
//}
//
//func (kv keyVal) String() string {
//	return fmt.Sprintf("{hamt_key.Key(\"%s\"), val:%v}", kv.key, kv.val)
//}
