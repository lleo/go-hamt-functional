
SYNOPSIS
========

This library is a Go language implementation of a functional HAMT, Hash Array
Mapped Trie.

There is a lot of Jargon in there to be unpacked. Functional mean immutable
and persistent. The term "immutable" means that the datastructure is is never
changed after construction. Where "persistent" means that when that immutable
datastructure is modified, is is based on the previous datastructure. In other
words the new datastructure is not a copy with the modification applied.
Instead, that new datastructure shares all unmodified parts of the previouse
datastructure and only the changed parts are copied and modified; unchanged
parts of the datastructure are shared between the old and new version.

Imagine a hypothetical tree structure with four leaves, two interior nodes and
a root node. If you change the fourth leaf node, then a new fourth leaf node
is created, as well as it's parent interior node, and a new root node.

            root tree node   root tree node'
                /    \         /   \
               /  +---\----- +      \
              /  /     \             \
         tree node 1   tree node 2  tree node 2'
            /  \          /  \        /   \
           /    \        / +--\------+     \
          /      \      / /    \            \
      Leaf 1   Leaf 2 Leaf 3  Leaf 4      Leaf 4'

Given this approach to changing a tree, a tree with a wide branching factor
would be relatively shallow. So the path from root to leaf would be short and
the amount of shared content would be substantial.

A Hash Array Mapped Trie is Trie where each node is represented by an Array.
The pointer to the next node down in the Trie is selected by and index drived
from Hash of the Key. For node arrays 32 entries wide we choose to use a 30 bit
hash value. The hash value is chopped into 6 groups of 5 bits each. That works
out nicely because 5 bits codifies an unsigned integer from 0 to 31. So each of
the six 5 bit values index into each of the tree node's 32 entry array.

Lets calculate the 30bit hash value for the key "a". 

	mask30 := uint(1<<30) - 1
    h := fnv.New32()
	h.Write([]byte("a"))
	h32 := h.Sum32() //uint32
	h30 := (h32 >> 30) ^ (h32 & mask30)
	fmt.Printf("%d 0x%08x 0b%032b\n", h30, h30, h30)

    84696446 0x050c5d7e 0b00000101000011000101110101111110

We can take the lower 30bits of "00000101000011000101110101111110" and split it
into six 5 bit values betweeen 0-31.

    00010 10000 11000 10111 01011 11110
	2, 16, 24, 23, 11, 30

If we format it from lowest bit to highest and zero-pad the numbers and
join those numbers with a "/" and precede it with the same "/" we get a string:

    /30/11/23/24/16/02

That is what this library calls a hash path. This allows you to address a node
in a tree with 32 entry branches.

However we are working with a Trie. That means in order to look up a key we
only have to walk down enough elements of the hash path to find a node that is
not a branch but rather a leaf.

In order to insert into a Trie, we walk down the Trie to find a leaf or an
empty slot. If it is an empty slot we simply insert a new leaf. If we find a
preexisting leaf, it gets a more complicated. We insert a new node array into
that slot and we try to insert the new leaf and the old leaf into that new
array using the hash path of each leaf one level deeper. The complication
occurs when the next entry in the hash paths of both leafs are the same. That
results in another collision. So we have to create another array and try to
insert both leaves at the next deeper index in their hash paths. We continue
that till the maximum depth is achieved. In the end we just start stacking
all the leaves that collide at the maximum depth into a fat ""collision leaf".
But truthfully we won't have many leaves colliding with identical 30 bit hash
paths.

Deleting from the HAMT is simpler that insertion. We find the leaf, if it
exists, along the leaf's hash path and remove it. The only complicated part is
when there is only one leaf left in an array level. When that happends we
replace the array level's position in it's parent with the remaining leaf. This
allows us to collapse the HAMT upon deletion.

The last part of this HAMT implementation to deal with is the "immutable" and
"persistent" properties. When we change a node array at the by adding or
deleting a leaf, we first have to copy that array and change the copy. Then
we have to copy and update the parent to point to the changed node array. We
continue this copy and update recursivly down to the root node array. When
the roo node array is copied and updated, a new HAMT structure is created to
represent the modified HAMT value. However, you should take note that most of
the previous HAMT is preserved as only the node arrays of the hash path to the
modification were copied and updated.


USAGE
=====

It is simplest to import the version of the HAMT library you want to use. The
32 way branching HAMT (hamt32) or the 64 way branching HAMT (hamt64). I
strongly recommend the hamt32 it is faster and wastes the least memory on
unused table entries.

To import the hamt32 library:

    import "github.com/lleo/go-hamt-functional/hamt32"

And you can refer to it as "hamt32" from then on.

Since this is an "immutable" implementation of HAMT you do not have to
constructed a new HAMT structure via 'new()'. Just use hamt32.Hamt as a
value.

    import (
        "github.com/lleo/go-hamt-functional/hamt32"
        "github.com/lleo/go-hamt-key/stringkey"
    )
    
    var h hamt32.Hamt
    for i, s := range []string{"foo", "bar", "baz"} {
        var k = stringkey.New(s)
        var added bool
        h, added = h.Put(k, i)
            if !added {
                log.Panicf("replaced, didn't add, value %d, for key %s", i, k)
            }
    }

Notice that we are keeping the current value of the Hamt in the 'h' variable
and forgetting the previous value. That allows some of the previous Hamt's
internal datastructures to be garbage collected as needed.



