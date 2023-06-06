// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "fmt"
import "reflect"
import "unsafe"

/* This is a tree node. */
type node struct {
	/* 16 */ Bytes string /* slice of bytes for key */
	/*  4 */ Parent uint32
	/*  4 */ Left uint32
	/*  4 */ Right uint32
	/*  2 */ Start int16 /* the first representative bit in this node */
	/*  2 */ End int16 /* the first non-representative bit in this node */
	/* 32 */
}

// Node is a struct which describe leaf of the tree.
type Node struct {
	/* 32 */ node node // It is absolutely necessary this member was the first
	/* 16 */ Data interface{} // Contains interface matching the node
	/* 48 */
}

func n2N(n *node)(*Node) {
	return reflect.NewAt(reflect.TypeOf(Node{}), unsafe.Pointer(n)).Interface().(*Node)
}

const null = uint32(0x00000000)
const node_sz = uint32(unsafe.Sizeof(node{}))
const leaf_sz = uint32(unsafe.Sizeof(Node{}))

// Radix is the struct which contains the tree root.
type Radix struct {
	Node uint32
	length int
	node node_pool
	leaf leaf_pool
	ptr_range []ptr_range
}

// NewRadix return initialized *Radix tree.
func NewRadix()(*Radix) {
	var radix *Radix

	radix = &Radix{}
	radix.length = 0

	return radix
}

// Len return the number of leaf in the tree.
func (r *Radix)Len()(int) {
	return r.length
}

// Node_counters describe tree node/leaf counters
type Node_counters struct {
	Capacity int // the total nodes/leaf capacity
	Free int // the number of free nodes/leaf
	Size int // the size of a node/leaf in bytes
}

// Node_counters describe tree counters
type Counters struct {
	Length int // Number of leaf used in the tree
	Node Node_counters // Counters relative to nodes
	Leaf Node_counters // counters relative to leaf.
}

// Counters return counters useful to monitor the radix tree
// usage.
func (r *Radix)Counters()(*Counters) {
	return &Counters{
		Length: r.length,
		Node: Node_counters{
			Capacity: r.node.capacity,
			Free: r.node.free,
			Size: int(node_sz),
		},
		Leaf: Node_counters{
			Capacity: r.leaf.capacity,
			Free: r.leaf.free,
			Size: int(leaf_sz),
		},
	}
}

// Equal return true if nodes are equal. Node are equal if there are the same
// prefix length and bytes.
func Equal(n1 *Node, n2 *Node)(bool) {
	return equal(&n1.node, &n2.node)
}
func equal(n1 *node, n2 *node)(bool) {
	if n1.End != n2.End {
		return false
	}
	return bitcmp([]byte(n1.Bytes), []byte(n2.Bytes), 0, n1.End)
}

/* Print node value */
func (r *Radix)get_string(n *node)(string) {
	var out string
	var b byte
	var mode string
	var ref uint32

	if n == nil {
		return "nil"
	}

	ref = r.n2r(n)

	if is_leaf(ref) {
		mode = "LEAF"
	} else {
		mode = "NODE"
	}

	for _, b = range []byte(n.Bytes) {
		out = fmt.Sprintf("%s0x%02x ", out, b)
	}
	return fmt.Sprintf("%s/ %d, mode=%s, Left=[%d][%d], Right=[%d][%d], Parent=[%d][%d]",
	                   out, n.End + 1, mode,
	                   n.Left >> 16, n.Left & 0xffff,
	                   n.Right >> 16, n.Right & 0xffff,
	                   n.Parent >> 16, n.Parent & 0xffff)
}

/* Return true if n is a children of a */
func (n *node)isChildrenOf(p *node)(bool) {
	return is_children_of([]byte(n.Bytes), []byte(p.Bytes), n.End, p.End)
}

func (n *node)isAlignedChildrenOf(p *node)(bool) {
	if !is_children_of([]byte(n.Bytes), []byte(p.Bytes), n.End, p.End) {
		return false
	}
	if p.End == n.End {
		return true
	}
	return are_zero([]byte(n.Bytes), int(p.End) + 1, int(n.End))
}

// LookupLonguestPath take the radix tree and a key/length prefix, return the list
// of all leaf matching the prefix. If none match, return nil
func (r *Radix)LookupLonguestPath(data *[]byte, length int16)([]*Node) {
	var node *node
	var path_node []*Node
	var end int16
	var ref uint32

	/* Browse tree */
	length-- /* convert length to index of last bit */
	path_node = make([]*Node, 0)
	ref = r.Node
	node = r.r2n(r.Node)
	for {

		/* We reach end of tree */
		if node == nil {
			return path_node
		}

		/* Cannot match because the key length is shorter than the node length */
		if length < node.End {
			return path_node
		}

		/* Match node. Perform bitcmp only if the input length is greater than current node length */
		if node.End != -1 && !bitcmp([]byte(node.Bytes), *data, node.Start, node.End) {
			return path_node
		}
		if is_leaf(ref) {
			path_node = append(path_node, n2N(node))
		}

		/* If the node no match or we reach end of browsing, return data */
		if length <= node.End {
			return path_node
		}

		/* Continue browsing: get the value of next bit.  */
		end = node.End + 1
		if (*data)[end / 8] & (0x80 >> (end % 8)) != 0 {
			ref = node.Right
			node = r.r2n(node.Right)
		} else {
			ref = node.Left
			node = r.r2n(node.Left)
		}
	}
}

// LookupLonguest get a key/length prefix and return the leaf which match the
// longest part of the prefix. Return nil if none match.
func (r *Radix)LookupLonguest(data *[]byte, length int16)(*Node) {
	var node *node
	var last_node *Node
	var end int16
	var ref uint32

	/* Browse tree */
	length-- /* convert length to index of last bit */
	ref = r.Node
	node = r.r2n(r.Node)
	for {

		/* Check if processed node is nil */
		if node == nil {
			return last_node
		}

		/* Can't match because the inpout key length is less than node,
		 * Otherwise, check the match.
		 */
		end = node.End
		if length < end || (end != -1 && !bitcmp([]byte(node.Bytes), *data, node.Start, end)) {
			return last_node
		}

		/* store node according with match_only
		 * if the node match the entry, always add node
		 * also add node if match_only is not required
		 */
		if is_leaf(ref) {
			last_node = n2N(node)
		}

		/* We reach the end */
		if length <= end {
			return last_node
		}

		/* Continue browsing: get the value of next bit.  */
		end++
		if (*data)[end / 8] & (0x80 >> (end % 8)) != 0 {
			ref = node.Right
			node = r.r2n(node.Right)
		} else {
			ref = node.Left
			node = r.r2n(node.Left)
		}
	}
}

// Get gets a key/length prefix and return exact match of the prefix. Exact match
// is a node wich match the prefix bit and the length.
func (r *Radix)Get(data *[]byte, length int16)(*Node) {
	var n *Node
	n = r.LookupLonguest(data, length)
	if n == nil {
		return nil
	}
	if n.node.End + 1 != length {
		return nil
	}
	return n
}

func lookup_longuest_last_node(r *Radix, data []byte, length int16)(*node, uint32) {
	var node *node
	var end int16
	var ref uint32

	/* Browse tree */
	length-- /* convert length to index of last bit */
	ref = r.Node
	node = r.r2n(r.Node)
	for {

		/* If node is nil, return nil, otherwise, return the node
		 * because the tree match at least the first bit
		 */
		if node == nil || length <= node.End {
			return node, ref
		}

		/* Perform bitcmp only if the input length is greater than current node length
		 * If the node match, continue browsing, otherwise return node.
		 */
		if node.End != -1 && !bitcmp([]byte(node.Bytes), data, node.Start, node.End) {
			return node, ref
		}

		/* Continue browsing: get the value of next bit.  */
		end = node.End + 1
		if data[end / 8] & (0x80 >> (end % 8)) != 0 {
			if node.Right == null {
				return node, ref
			}
			ref = node.Right
			node = r.r2n(node.Right)
		} else {
			if node.Left == null {
				return node, ref
			}
			ref = node.Left
			node = r.r2n(node.Left)
		}
	}
}

func (r *Radix)replace(o *node, n *node) {
	var replace_node *node

	*n = *o

	if n.Parent != null {
		replace_node = r.r2n(n.Parent)
		if replace_node.Left == r.n2r(o) {
			replace_node.Left = r.n2r(n)
		} else {
			replace_node.Right = r.n2r(n)
		}
	}

	if n.Left != null {
		replace_node = r.r2n(n.Left)
		replace_node.Parent = r.n2r(n)
	}

	if n.Right != null {
		replace_node = r.r2n(n.Right)
		replace_node.Parent = r.n2r(n)
	}
}

// Insert key/length prefix in the tree. The tree accept only unique value, if
// the prefix already exists in the tree, return existing leaf,
// otherwaise return nil.
func (r *Radix)Insert(key *[]byte, length int16, data interface{})(*Node, bool) {
	var leaf *Node
	var lookup_node *node
	var newnode *node
	var bitno int16
	var l int16
	var ref uint32

	if length == 0 {
		return nil, false
	}

	/* Browse tree and return the closest node */
	lookup_node, ref = lookup_longuest_last_node(r, *key, length)

	/* Create leaf node */
	leaf = r.leaf_alloc()
	leaf.node.Bytes = string(*key)
	leaf.node.Start = 0
	leaf.node.End = length - 1
	leaf.node.Parent = null
	leaf.node.Left = null
	leaf.node.Right = null
	leaf.Data = data

	/* CASE #1
	 *
	 * Special case, tree is empty, create node
	 */
	if lookup_node == nil {
		r.Node = r.n2r(&leaf.node)
		r.length++
		return leaf, true
	}

	/* The last node exact match the new entry */
	if length > lookup_node.End && bitcmp(*key, []byte(lookup_node.Bytes), lookup_node.Start, lookup_node.End) {

		/* CASE #2
		 *
		 * First, if we have a perfect match, just modify
		 * existing node.
		 *
		 * INSERT-KEY 0101 / 4
		 * STOP-NODE  0101 / 4
		 *
		 */
		if lookup_node.End == length - 1 {

			/* Unique mode is active and the data is set, return stored data */
			if is_leaf(ref) {
				return n2N(lookup_node), false
			}

			/* replace original not leaf node by leaf Node, and
			 * release memory of the non-leaf node
			 */
			r.replace(lookup_node, &leaf.node)
			r.free(lookup_node)

			r.length++
			return leaf, true
		}

		/* CASE #3
		 *
		 * The appended network is greater than the last node
		 * but the lookup stops on this node. So next node is
		 * not set.
		 * Determine the first bit after the last significant
		 * bit of matched node, and choose append left or right
		 *
		 * INSERT-KEY 010111 / 6
		 * STOP-NODE  0101 / 4
		 */
		leaf.node.Start = lookup_node.End + 1
		leaf.node.Parent = r.n2r(lookup_node)
		if bitget(*key, lookup_node.End + 1) == 1 {
			lookup_node.Right = r.n2r(&leaf.node)
		} else {
			lookup_node.Left = r.n2r(&leaf.node)
		}
		r.length++
		return leaf, true
	}

	/* Match the longuest part in the key */
	if leaf.node.End < lookup_node.End {
		l = leaf.node.End
	} else {
		l = lookup_node.End
	}
	bitno = bitlonguestmatch(*key, []byte(lookup_node.Bytes), lookup_node.Start, l)
	if bitno == -1 {

		/* CASE #4
		 *
		 * if the new key match exactly current key, but it have
		 * less length, just insert node between current node and
		 * its parent.
		 *
		 * INSERT-KEY 0101 / 4
		 * STOP-NODE  010111 / 6
		 */
		if lookup_node.Parent != null {
			leaf.node.Start = r.r2n(lookup_node.Parent).End + 1
		}
		leaf.node.Parent = lookup_node.Parent
		lookup_node.Parent = r.n2r(&leaf.node)
		lookup_node.Start = leaf.node.End + 1

		/* Append existing nodes */
		if bitget([]byte(lookup_node.Bytes), lookup_node.Start) == 1 {
			leaf.node.Right = r.n2r(lookup_node)
			leaf.node.Left = null
		} else {
			leaf.node.Right = null
			leaf.node.Left = r.n2r(lookup_node)
		}

		/* Update original parent */
		if leaf.node.Parent == null {
			r.Node = r.n2r(&leaf.node)
		} else if r.r2n(leaf.node.Parent).Left == r.n2r(lookup_node) {
			r.r2n(leaf.node.Parent).Left = r.n2r(&leaf.node)
		} else {
			r.r2n(leaf.node.Parent).Right = r.n2r(&leaf.node)
		}

		r.length++
		return leaf, true
	}

	/* CASE #5
	 *
	 * The node key partially match (at least the first byte)
	 * of the input network. We determine length of common
	 * path, we insert common node. Add the new node at left or
	 * right. Adjust the current node and put it at left or right.
	 *
	 * INSERT-KEY 010101 / 6
	 * STOP-NODE  010111 / 6
	 */

	/* create new node */
	newnode = r.node_alloc()
	newnode.Bytes = string(*key)
	newnode.Start = lookup_node.Start
	newnode.End = bitno - 1
	newnode.Parent = lookup_node.Parent

	/* Update existing node */
	lookup_node.Start = bitno
	lookup_node.Parent = r.n2r(newnode)

	/* Update leaf */
	leaf.node.Start = bitno
	leaf.node.Parent = r.n2r(newnode)

	/* Append existing nodes */
	if bitget(*key, bitno) == 1 {
		newnode.Right = r.n2r(&leaf.node)
		newnode.Left = r.n2r(lookup_node)
	} else {
		newnode.Right = r.n2r(lookup_node)
		newnode.Left = r.n2r(&leaf.node)
	}

	/* Update original parent */
	if newnode.Parent == null {
		r.Node = r.n2r(newnode)
	} else if r.r2n(newnode.Parent).Left == r.n2r(lookup_node) {
		r.r2n(newnode.Parent).Left = r.n2r(newnode)
	} else {
		r.r2n(newnode.Parent).Right = r.n2r(newnode)
	}

	r.length++
	return leaf, true
}

// Delete remove Node from the tree.
func (r *Radix)Delete(n *Node) {
	r.del(&n.node)
	r.length--
}

func (r *Radix)del(n *node) {
	var p *node
	var c *node
	var ref uint32

	/* If the node has two childs and it is a Leaf, replace it by
	 * a node. Otherwise, do nothing
	 */
	if n.Left != null && n.Right != null {
		ref = r.n2r(n)
		if is_leaf(ref) {
			p = r.node_alloc()
			r.replace(n, p)
			r.free(n)
		}
		return
	}

	/* If node has one child. Remove the node, and
	 * Link the child to the parent. Change child bits
	 */
	if (n.Left == null) != (n.Right == null) {

		if n.Left != null {
			c = r.r2n(n.Left)
		} else {
			c = r.r2n(n.Right)
		}
		c.Start = n.Start
		c.Parent = n.Parent
		if n.Parent == null {
			r.Node = r.n2r(c)
			r.free(n)
			return
		}
		if r.r2n(n.Parent).Left == r.n2r(n) {
			r.r2n(n.Parent).Left = r.n2r(c)
		} else {
			r.r2n(n.Parent).Right = r.n2r(c)
		}
		r.free(n)
		return
	}

	/* If the node has no childs, just remove it. */
	if n.Left == null && n.Right == null {

		/* we reach root */
		if n.Parent == null {
			r.Node = null
			return
		}

		/* Remove my branch on the parent node */
		p = r.r2n(n.Parent)
		if p.Left == r.n2r(n) {
			p.Left = null
		} else if p.Right == r.n2r(n) {
			p.Right = null
		}

		/* if the parent node is a leaf, do not remove */
		if !is_leaf(n.Parent) {
			r.free(n)
			return
		}

		/* Remove the parent node */
		r.del(p)
	}
}

/*
 *  ()        ()
 *    \      /
 *    L \  / R
 *       ()        ()
 *       P \      /
 *           \  /
 *            ()
 */

// Next return next Node in browsing order. Return nil
// if we reach end of tree.
func (r *Radix)Next(n *Node)(*Node) {
	return r.next(&n.node)
}

func (r *Radix)next(n *node)(*Node) {
	var prev uint32
	var ref uint32

	prev = null
	ref = r.n2r(n)

	for {
		if prev == n.Parent || prev == null {
			/* we come from parent, go left, right and them parent */
			prev = ref
			if n.Left != null {
				ref = n.Left
				n = r.r2n(n.Left)
			} else if n.Right != null {
				ref = n.Right
				n = r.r2n(n.Right)
			} else if n.Parent != null {
				ref = n.Parent
				n = r.r2n(n.Parent)
			}
		} else if prev == n.Left {
			/* we come from left branch, go right or go back */
			prev = ref
			if n.Right != null {
				ref = n.Right
				n = r.r2n(n.Right)
			} else if n.Parent != null {
				ref = n.Parent
				n = r.r2n(n.Parent)
			}
		} else if prev == n.Right {
			/* we come from right branch, we go back */
			prev = ref
			if n.Parent != null {
				ref = n.Parent
				n = r.r2n(n.Parent)
			}
		}

		/* None match, this is the end */
		if ref == prev {
			return nil
		}

		/* If we reach leaf, and I'n not com from parent, return node */
		if is_leaf(ref) && prev == n.Parent {
			return n2N(n)
		}

		/* Otherwise continue browsing */
	}
}

// First return first node of the tree. Return nil if the
// tree is empty.
func (r *Radix)First()(*Node) {
	if r.Node == null {
		return nil
	}

	/* If entry node is a leaf, return it */
	if is_leaf(r.Node) {
		return n2N(r.r2n(r.Node))
	}

	/* Otherwise return next node */
	return r.next(r.r2n(r.Node))
}

// Last return the last node of the tree. Return nil
// if the tree is empty.
func (r *Radix)Last()(*Node) {
	var n *node

	if r.Node == null {
		return nil
	}

	/* Return previous leaf, if there are no previous
	 * node and the entry point is a leaf, return
	 * entry point.
	 */
	n = r.r2n(r.Node)
	for {
		if n.Right != null {
			n = r.r2n(n.Right)
		} else if n.Left != null {
			n = r.r2n(n.Left)
		} else {
			return n2N(n)
		}
	}
}

// Iter is a struct for managing iteration
type Iter struct {
	node *node
	next_node *node
	key *[]byte
	length int16
	r *Radix
}

// NewIter return struct Iter for browsing all nodes there children
// match the given key/length prefix.
func (r *Radix)NewIter(key *[]byte, length int16)(*Iter) {
	var i *Iter
	var ref uint32

	i = &Iter{}
	i.key = key
	i.length = length
	i.r = r

	/* Lookup next node */
	if length == 0 {
		ref = r.Node
		i.next_node = r.r2n(r.Node)
	} else {
		i.next_node, ref = lookup_longuest_last_node(r, *key, length)
		if i.next_node != nil && !is_children_of([]byte(i.next_node.Bytes), *i.key, i.next_node.End, i.length - 1) {
			i.next_node = nil
		}
	}

	/* No nodes found, next node is nil, abort iteration */
	if i.next_node == nil {
		return i
	}

	/* If the first node matching is a leaf, there is the entry point */
	if is_leaf(ref) {
		return i
	}

	/* Otherwise, lookup for next leaf */
	i.set_next()

	return i
}

func (i *Iter)set_next()() {
	var n *Node

	if i.next_node == nil {
		return
	}
	n = i.r.next(i.next_node)
	if n == nil {
		i.next_node = nil
		return
	} else {
		i.next_node = &n.node
	}
	if i.length > 0 && !is_children_of([]byte(i.next_node.Bytes), *i.key, i.next_node.End, i.length - 1) {
		i.next_node = nil
	}
}

// Next return true if there next node avalaible. This function
// also perform lookup for the next node.
func (i *Iter)Next()(bool) {
	i.node = i.next_node
	i.set_next()
	return i.node != nil
}

// Get return the node. Many calls on this function return  the same
// value.
func (i *Iter)Get()(*Node) {
	return n2N(i.node)
}
