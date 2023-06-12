[![GoDoc](https://pkg.go.dev/badge/github.com/thierry-f-78/go-radix)](https://pkg.go.dev/github.com/thierry-f-78/go-radix)

# Radix package provide radix algorithm.

_https://en.wikipedia.org/wiki/Radix_tree_

This implementation allow indexing key with a bit precision. Note, the wikipedia reference show a radix tree with a byte precision. It seems natural with string. This radix tree is designed to manipulate networks.

On goal of this implementation is using little amount of memory, in order to deal wiht huge trees.

- Each tree accepts ~ ![formula](https://render.githubusercontent.com/render/math?math=2\times10^{12}) nodes.

- Keys are 16bit encoded, so could have 64k bits of length or 8kB.

- Like most tree algothm, keys are stored sorted, accessing the little or greater value is very fast.

- Lookup algortithm complexity is O(log(n)), tree depth is log(n).

- The package provide facility to use string, uint64 and network as key.

## Benchmark

go-radix is written to be fast and save memory. Below basic benchmark using CPU `Intel(R) Core(TM) i7-1068NG7 CPU @ 2.30GHz`. Using IPv4 networks as key. The benchmark code is provided in radix_test.go file with a reference dataset.

- Index 2 649 172 entries in 1.22s, so 2.17M insert/s

- Lookup 1 000 000 random data in 0.78s with 146 190 hit, 853 810 miss, so 1.28M lookup/s

- Browse 2 649 172 entries in 0.47s, so 5.64M read/s

- Delete 2 649 172 sequential data in 0.74s, so 3.58M delete/s

Nothing about memory consumation in this test beacause it hard to measure relevant data, because the garbage collector.

## Example

```go
package radix

import "net"

import "github.com/thierry-f-78/go-radix"

func Example_ipv4() {

	// Create new tree root
	r := radix.NewRadix()

	// Insert first network
	_, n1, _ := net.ParseCIDR("10.0.0.0/16") 
	r.IPv4Insert(n1, "This is the first network inserted")

	// Lookup the network
	_, n2, _ := net.ParseCIDR("10.0.0.33/32") 
	node1 := r.IPv4LookupLonguest(n2)
	if node1 != nil {
		println("network", n2.String(), "is contained in network", node1.IPv4GetNet().String())
		println("network", node1.IPv4GetNet().String(), "is associated with data", node1.Data.(string))
	}

	// Lookup too large network
	_, n3, _ := net.ParseCIDR("10.0.0.0/8") 
	node2 := r.IPv4LookupLonguest(n3)
	if node2 == nil {
		println("network", n3.String(), "has no entries in the tree")
	}
}

func Example_string() {
	// Create new tree root
	r := radix.NewRadix()

	// insert string
	r.StringInsert("home", "This is a prefix")

	// lookup word
	node1 := r.StringLookupLonguest("homemade")
	if node1 != nil {
		println("homemade has prefix", node1.StringGetKey(), "in the tree, with data", node1.Data.(string))
	}
}

func main() {
	Example_ipv4()
	Example_string()
}
```

## Internals

The tree use two kind of nodes. Internal node `type node struct` were are not exported and leaf node `type Node struct`. These name are ambiguous, this is due to a necessary compatibility with existing programs. Node are just an interface{} associated with a
`node struct`.

### Saving memory

Because this king of tree could contains millions of entries, it is important to use the mimnimum of memory. Member structs are obviously sorted in order to respect alignment packing data without hole.

In other way, I used three means to compact memory:

#### 1. Saving memory : use string in place of []byte

The most simpler is storing []byte with string type. The string uses 16 byte, the []byte uses 24 bytes.

#### 2. Saving memory : use 32bit reference in place of 64bit pointers

More tricky, how use 32bit reference in place of 64bit pointers for chaining nodes. I use memory pool and 64k array of 64k array of nodes. The following explanation is not complete, but the code comment contains more information. There just an introduction.

Usage of memory pool implies while the radix tree is in use, allocated node are not garbage colected.

```
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
```

memory_pool growth by block of 64k nodes. Each time there are no free node, memory pool has one more slot of 64k node.

memory_backref is dichotomic sorted list of pointer start and stop which reference the memory_pool index which contains pointer. `ptr_start = &[0]node, ptr_stop = &[65535]node`.

memory_free is the list of avalaible nodes.

node pointer from reference = &memory_pool[x][y], where x and y and 16bit values

node reference from pointer:

- x = dichotomic search of memory_pool_index in memory_backref array using &node pointer reference
- y = (&node - ptr_start) / node_size

In reality, there are two distinct pool per tree using this concept. The pool of node and the pool of leaf. the memory_backref contains also a boolean value to distinct these two
kind of pool.

In reality x is not used until 64k but limited to 32k and the msb is used to differenciate the two pools.

#### 3. Saving memory : split node types according with their kind to avoid interface{} pointer

Some node are used for internal chainaing, other nodes are used to display data. data is stored in an interface which uses 16bytes. It is garbage to let these 16bytes for each nodes. there the approximatively the same numbers of internal nodes than
exposed nodes, so for 1M data indexed, 1M of internal nodes.

```
 type node struct {
    ... // chaining things
 }
 type Node struct {
    node node
    Data interface{}
 }
````

I use a property which is &Node adress = &Node.node adress. I manipulate only node to chain each data. If I need to insert a leaf, I use Node.node. When I browse the tree I need to kwnown if I encounter leaf or node. I just chack the msb of the reference.
