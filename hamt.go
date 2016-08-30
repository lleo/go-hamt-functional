/*
Package hamt is the front-end for a 32bit and 64bit implementations of a
functional Hash Array Mapped Trie (HAMT) datastructure.

In this case, functional means immutable and persistent. "immutable" means
that the datastructure is is never changed after construction. Where
"persistent" means that when a new HAMT structure is built, based on a
previous datastructure, that new datastructure shares all un-modified parts
of the "parent" datastructure. The changes in a "persistent" datastructure
are the leaf and up each interior node along the path up to the root.

Given how wide a HAMT node is (either 32 or 64 nodes wide) HAMT datastructures
not very deep; either 6, for 32bit, or 10, for 64bit implementations, nodes
deep. This neans HAMTs are effectively O(1) for Search, Insertions, and
Deletions.

Both 32 and 64 bit implementations of HAMTs are of fixed depth is because they
are [Tries](https://en.wikipedia.org/wiki/Trie). The key of a Trie is split
into n-number smaller indecies and each node from the root uses each successive
index. For example a Trie with a string key would be split into chanracters and
each node from root would be indexed by the next character in the string key.

In the case of a this HAMT implementation the key is hashed into a 30 or 60 bit
number. In the case of the string_key we take the []byte slice of the string
and feed it to hash.fnv.New32() or New64() hash generator. Since these
generate 32 and 64 bit hash values respectively and we need 30 and 60 bit
values, we use the [xor-fold technique](http://www.isthe.com/chongo/tech/comp/fnv/index.html#xor-fold)
to "fold" the high 2 or 4 bits of the 32 and 64 bit hash values into 30 and
60 bit values for our needs.

We want 30 and 60 bit values because they split nicely into six 5bit and ten
6bit values respectively. Each of these 5 and 6 bit values become the indexies
of our Trie nodes with a maximum depth of 6 or 10 respectively. Further 5 bits
indexes into a 32 entry table nodes for 32 bit HAMTs and 6 bit index into 64
entry table nodes for 64 bit HAMTs; isn't that symmetrical :).

For a regular HAMT, when key, value pair must be created, deleted, or changed
the key is hashed into a 30 or 60 bit value (described above) and that hash30
or hash60 value represents a path of 5 or 6 bit values to place a leaf
containing the key, value pair. For a Get() or Del() operation we lookup the
the deepest node along that pate that is not-nil. For a Put() operation we
lookup the deepest location that is nil and not beyond the lenth of the path.

For a "functional" HAMT, the Get() operation is unchanged. For the Put() we
lookup a parent node with a nil entry along the given path, copy the parent,
insert the new node in the appropriate location in the copy then turn round
and copy the grandparent inserting the new parent into the grandparent along
the path repeatedly up to the root. A simmilar opeation happens for Del().
For both Put() and Del() the last operation of the newly copied `Hamt struct`
is to update the `nentries` entry.
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

// NewHamt32 returns the hamt32.EMPTY value. Given that Hamt
// structs are immutable a single hamt32.EMPTY value can be used.
func NewHamt32() hamt32.Hamt {
	return hamt32.EMPTY
}

// NewHamt64 returns the hamt64.EMPTY value. Given that Hamt
// structs are immutable a single hamt64.EMPTY value can be uses.
func NewHamt64() hamt64.Hamt {
	return hamt64.EMPTY
}
