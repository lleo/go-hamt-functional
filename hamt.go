/*
Package hamt is the front-end for a 32bit and 64bit implementations of a
functional Hash Array Mapped Trie (HAMT) datastructure.

In this case, functional means immutable and persistent. The term "immutable"
means that the datastructure is is never changed after construction. Where
"persistent" means that when that immutable datastructure is modified, is is
based on the previous datastructure. In other words the new datastructure is
not a copy with the modification applied. In stead, that new datastructure
shares all un-modified parts of the previouse datastructure and only the
changed parts are copied and modified; unchanged parts of the datastructure
are shared between the old and new version.

A HAMT structure is a tree with a fixed & wide branching factor. Trees make
and excellent datastructure to be immutable and persistent. Our HAMT starts
with a root branch. Branches are called tables, because they are represented
as tables with the "branching factor" number of entries. These entries may
be one of three types of nodes: further branches (aka tables) or key/val
entries (aka leafs) or emtpy (aka nil).

A HAMT is a key/val indexing datastructure. Rather than indexing the
datastructure on the key directly, which could result in a rather deep tree
datastructure. We generate a hash value of the key, and split the hash value
into a fixed number of indexes into fixed size arrays. This results in a
tree with a maximum depth and a wide branching factor.

For example, we can use a key type of a string. Hash that string into a 32 bit
hash value. Coerce that 32 bit value into a 30 bit value. Then split that
30 bit hash value into six 5 bit values. Those 5 bit values will index perfectly
into tree nodes with 32 wide branching factor. Now we have tree with a string
for the key that is AT MOST six levels deep, in other words O(1) lookup and
modification operations.

Lets call the number of hash bits H (for hash value). The number of parts the
hash value can be split into we'll call D (for depth). The width of each table
(aka branching factor) is 2^B; I think of B as "bits per level". The
relationship of H:D:B is given by H/B = D. I've implemented in hamt32 (H=30,
D=6, B=5) and hamt64 is (H=60, D=10, B=6). We could call the branching factor
W for "width" of each tree node. However W is superfluous, because it can be
derived from B (aka W=2^B).

The number in hamt32 is the branching factor W=2^B=32; from H=30,D=6,B=5 .
The number in hamt64 is the branching factor W=2^B=64; from H=60,D=10,B=6 .

HAMTs are [Tries](https://en.wikipedia.org/wiki/Trie), because when we are
trying to find a location to Get, Put, or Delete a key/value pair we mearly
have to walk the "hash path" till we find a non-branching node. The HashPath
is the H bit hash value, split into a ordered sequence of B bit integer values
that is, at most, D entry tries long.

Lets start with a concrete example of a hamt32 (aka H=30,D=6,B=5). Given the
string "ewyx" the Hash30() HashVal30 is 0x11a01c5e. Converted into six descreet
5 bit values (from lowest bit to highest) you get 30, 2, 7, 0 26, and 8. This
library prints them out from HashVal30.String() as "/30/02/07/00/26/08"; The
hash path from lowest to highest bit. That string, "/30/02/07/00/26/08", is the
"hash path". Looking up where to find this entry we look up the 30th index of
the root of the tree, if that entry is another branch we look up the 2nd index
of that next branch. We continue (7th, 0th et al) until we find a non-branch
entry. The non-branch entry can be a leaf or empty.

Just to be pedantic the go-hamt-key API calculates the indexes by depth as
follows:

    my k = stringkey.New("ewyx")
    var h30 = key.Hash30() //=> 295705694 or 0x11a01c5e
    h30.BitString()        //=> "00 01000 11010 00000 00111 00010 11110"
    h30.Index(0)           //=> 11110 or 30
    h30.Index(1)           //=> 00010 or 2
    h30.Index(2)           //=> 00111 or 7
    h30.Index(3)           //=> 00000 or 0
    h30.Index(4)           //=> 11010 or 26
    h30.Index(5)           //=> 01000 or 8
    h30.String()           //=> "/30/02/07/00/26/08"

Now we know how to find the candidate location or entry for our operation. That
operation can be either a straight lookup, called with h.Get(k); or it can be
an insertion of a key/value pair, called with h.Put(k,v); or lastly it can be a
deletion operation, called with h.Del(k).

For either hamt32.Hamt or hamt64.Hamt value we have three primary operations:
h.Get(), h.Put(), and h.Del().

Only h.Put() and h.Del() modify the HAMT. When they modify a table, first the
table is copied, then the modification is made to the copy. Next the parent
table must be copied so that the new table's entry in the copied parent may be
modified. This is continued to the root table and the HAMT structure itself is
copied. This is the h.persist() call. Hence, h.Put() and h.Del() return the new
HAMT structure as well as any other return values specific to h.Put() or
h.Del().

Given that Get() makes no modification of the HAMT structure, it only returns
a boolean indicating the key was found in the HAMT and the key's value.

Put() returns a copy of the HAMT and a boolean indicating whether a new entry
was added (true) or a current entry was updated (false).

Del() returns a boolean value indicating if the key was found, and if true what
the value of the deleted key was, and the new HAMT structure. If the Del()
didn't find the key (a false return value) key's value data is nil and the HAMT
value is the current HAMT.

*/
package hamt

import (
	"github.com/lleo/go-hamt-functional/hamt32"
	"github.com/lleo/go-hamt-functional/hamt64"
)

// NewHamt32 returns an empty hamt32.Hamt value. Given that Hamt
// structs are immutable we return the Hamt structure by value.
func NewHamt32() hamt32.Hamt {
	return hamt32.Hamt{}
}

// NewHamt64 returns an empty hamt64 value. Given that Hamt
// structs are immutable we return the Hamt structure by value.
func NewHamt64() hamt64.Hamt {
	return hamt64.Hamt{}
}
