package hamt

import "fmt"

type leafI interface {
	nodeI
	get(key []byte) (interface{}, bool)
	put(key []byte, val interface{}) (leafI, bool) //bool == replace? val
	del(key []byte) (leafI, interface{}, bool)     //bool == deleted? key
}

type keyVal struct {
	key []byte
	val interface{}
}

func (kv keyVal) String() string {
	return fmt.Sprintf("{[]byte(\"%s\"), val:%v}", kv.key, kv.val)
}

//byteSlicesEqual function
//
func byteSlicesEqual(k1 []byte, k2 []byte) bool {
	if len(k1) != len(k2) {
		return false
	}
	if len(k1) == 0 {
		return true
	}
	for i := 0; i < len(k1); i++ {
		if k1[i] != k2[i] {
			return false
		}
	}
	return true
}
