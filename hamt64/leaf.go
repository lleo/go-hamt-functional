package hamt64

import "github.com/lleo/go-hamt/key"

type leafI interface {
	nodeI
	get(key key.Key) (interface{}, bool)
	put(key key.Key, val interface{}) (leafI, bool) //bool == replace? val
	del(key key.Key) (leafI, interface{}, bool)     //bool == deleted? key
	keyVals() []keyVal
}

//type keyVal struct {
//	key key.Key
//	val interface{}
//}
//
//func (kv keyVal) String() string {
//	return fmt.Sprintf("{key.Key(\"%s\"), val:%v}", kv.key, kv.val)
//}
