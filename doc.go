// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>
//
// This library is free software; you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation, version 3
// exclusively.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public
// License along with this library; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301  USA

/*
Radix package provide radix algorithm.

https://en.wikipedia.org/wiki/Radix_tree

Properties

• This implementation allow indexing key with a bit precision. The wikipedia
reference show a radix tree with a byte precision. It seems natural with string.
This radix tree is designed to manipulate networks.

• Each tree could accept ~2x10^12 nodes.

• Keys are 16bit encoded, so could have 64k bytes of length.

• Like most tree algothm, keys are stored sorted, accessing the little or greater
value is very fast.

• Lookup algortithm complexity is O(log(n)), tree depth is log(n).

• The package provide facility to use string, uint64 and network as key.

Benchmark

It is written to be fast and save memory. Below basic benchmark using
CPU "Intel(R) Core(TM) i7-1068NG7 CPU @ 2.30GHz". Using IPv4 networks
as key. The benchmark code is provided in radix_test.go file with a
reference dataset.

• Index 2 649 172 entries in 1.22s, so 2.17M insert/s

• Lookup 1 000 000 random data in 0.78s with 146 190 hit, 853 810 miss, so 1.28M lookup/s

• Browse 2 649 172 entries in 0.47s, so 5.64M read/s

• Delete 2 649 172 sequential data in 0.74s, so 3.58M delete/s

Nothing about memory consumation in this test beacause it hard to measure relevant
data, because the garbage collector.

Internals

The tree use two kind of nodes. Internal node `type node struct` were are not exported and
leaf node `type Node struct`. These name are ambiguous, this is due to a necessary
compatibility with existing programs. Node are just an interface{} associated with a
`node struct`.

Saving memory

Because this king of tree could contains millions of entries, it is important to use
the mimnimum of memory. Member structs are obviously sorted in order to respect alignment
packing data without hole.

In other way, I used three means to compact memory:

1. Saving memory : use string in place of []byte

The most simpler is storing []byte with string type. The string uses 16 byte, the []byte uses
24 bytes.

2. Saving memory : use 32bit reference in place of 64bit pointers

More tricky, how use 32bit reference in place of 64bit pointers for chaining nodes. I use
memory pool and 64k array of 64k array of nodes. The following explanation is not complete,
but the code comment contains more information. There just an introduction.

Usage of memory pool implies while the radix tree is in use, allocated node are not garbage
colected.

 type node struct {
    ...
    next uint32
 }
 const node_size = unsafe.Sizeof(node{})
 var memory_pool [][65536]node
 var memory_backref []struct {
    ptr_start uintptr
    ptr_stop uintptr
    memory_pool_index int
 }
 var memory_free uint32

memory_pool growth by block of 64k nodes. Each time there are no free node, memory pool
has one more slot of 64k node.

memory_backref is dichotomic sorted list of pointer start and stop which reference the
memory_pool index which contains pointer. ptr_start = &[0]node, ptr_stop = &[65535]node.

memory_free is the list of avalaible nodes.

node pointer from reference = &memory_pool[x][y], where x and y and 16bit values

node reference from pointer:

• x = dichotomic search of memory_pool_index in memory_backref array using &node pointer reference

• y = (&node - ptr_start) / node_size

In reality, there are two distinct pool per tree using this concept. The pool of node
and the pool of leaf. the memory_backref contains also a boolean value to distinct these two
kind of pool.

In reality x is not used until 64k but limited to 32k and the msb is used to differenciate
the two pools.

3. Saving memory : split node types according with their kind to avoid interface{} pointer

Some node are used for internal chainaing, other nodes are used to display data.
data is stored in an interface which uses 16bytes. It is garbage to let these 16bytes
for each nodes. there the approximatively the same numbers of internal nodes than
exposed nodes, so for 1M data indexed, 1M of internal nodes.

 type node struct {
    ... // chaining things
 }
 type Node struct {
    node node
    Data interface{}
 }

I use a property which is &Node adress = &Node.node adress. I manipulate only node to chain
each data. If I need to insert a leaf, I use Node.node. When I browse the tree I need to
kwnown if I encounter leaf or node. I just chack the msb of the reference.

*/
package radix
