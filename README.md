
SYNOPSIS
========

This library is a Go language implementation of a functional HAMT, Hash Array
Mapped Trie.

There is a lot of Jargon in there to be unpacked. This library can be imported
into a Go language program with the following golang import statements:

	import hamt "github.com/lleo/go-hamt-functional"

The term, "persistent", means that any modification to the datastructure does
not change the original datastructure. Rather, a new top level data structure
is created which only contains the differences from the original, yet shares
all the unmodified parts of the datastructure. Say you have a Tree where you
add a leaf. The new leaf is added to a modified internal modified tree node.
That internal modified tree node is a copy of the original tree node. And each
tree node parent upto the root tree node is copied and modified.
									
	        root tree node   root tree node'
	            /    \         /   \   	 	
	           /  +-- \----- +      \ 	   	   	
              /  /     \             \
	   tree node 1   tree node 2  tree node 2'
	  	  /  \          /  \        /   \
	     /    \        / +--\------+     \
        /  	   \   	  /	/  	 \ 	   	   	  \
	Leaf 1	Leaf 2  Leaf 3  Leaf 4     Leaf 4'
												   
A Hash Array Mapped Trie is Trie where each node is represented by an Array.
The pointer to the next node down in the Trie is selected by and index drived
from the Hash. For a 32 bit hash value we use Arrays 32 entries deep and the
hash value is chopped into 6 groups of 5 bits each. For a 64 bit hash value,
we use Arrays 64 entries deep and the hash value is chopped into 10 groups of
6 bits each. Each bit group is the index into the next table; for 32 bit hash
values 2^5 == 32; for 64 bit hash values 2^6 == 64.


