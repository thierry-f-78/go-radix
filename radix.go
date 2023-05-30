// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "fmt"
import "unsafe"

/* This is a tree node. */
type Node struct {
	/* 16 */ Data interface{} /* Contains the list of interface matching the node */
	/* 16 */ Bytes string /* slice of bytes for key */
	/*  4 */ Parent uint32
	/*  4 */ Left uint32
	/*  4 */ Right uint32
	/*  2 */ Start int16 /* the first representative bit in this node */
	/*  2 */ End int16 /* the first non-representative bit in this node */
	/*  2 */ pool uint16
	/*  6    6 bytes reserved but unused for alignment */
	/* 56 */
}

var null uint32 = 0x80000000
var node_sz uint32 = uint32(unsafe.Sizeof(Node{}))

type chunk struct {
	nodes [65536]Node
	ptr uintptr
}

type Radix struct {
	Node *Node
	length int
	free int
	capacity int
	pool []*chunk
	next uint32
}

func (r *Radix)node_alloc()(*Node) {
	var n *Node

	if r.free == 0 {
		r.growth()
	}
	r.free--
	n = r.r2n(r.next)
	r.next = n.Left
	return n
}

func (r *Radix)node_free(n *Node)() {
	n.Data = nil
	n.Bytes = ""
	n.Parent = null
	n.Left = r.next
	n.Right = null
	r.free++
	r.next = r.n2r(n)
}

/* reference to node */
func (r *Radix)r2n(v uint32)(*Node) {
	if v == null {
		return nil
	}
	return &r.pool[v >> 16].nodes[v & 0xffff]
}

/* node to reference */
func (r *Radix)n2r(n *Node)(uint32) {
	var c *chunk
	var p uintptr

	p = uintptr(unsafe.Pointer(n))
	c = r.pool[n.pool]
	return (uint32(n.pool) << 16) | (uint32(p - c.ptr) / node_sz)
}

func (r *Radix)growth() {
	var c *chunk
	var i int

	if len(r.pool) >= 32768 {
		panic("reach the maximum number of node pools allowed")
	}
	c = &chunk{}
	c.ptr = (uintptr)(unsafe.Pointer(&c.nodes[0]))
	r.pool = append(r.pool, c)
	r.free += 65536
	r.capacity += 65536
	for i, _ = range c.nodes {
		c.nodes[i].pool = uint16(len(r.pool)) - 1
		c.nodes[i].Left = r.next
		r.next = r.n2r(&c.nodes[i])
	}
}

func NewRadix()(*Radix) {
	var radix *Radix

	radix = &Radix{}
	radix.length = 0

	return radix
}

func (r *Radix)Len()(int) {
	return r.length
}

/* Return true if nodes are equal */
func Equal(n1 *Node, n2 *Node)(bool) {
	if n1.End != n2.End {
		return false
	}
	return bitcmp([]byte(n1.Bytes), []byte(n2.Bytes), 0, n1.End)
}

/* Print node value */
func (n *Node)String()(string) {
	var out string
	var b byte
	var mode string

	if n == nil {
		return "nil"
	}

	if n.Data != nil {
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
func (n *Node)IsChildrenOf(p *Node)(bool) {
	return is_children_of([]byte(n.Bytes), []byte(p.Bytes), n.End, p.End)
}

func (n *Node)IsAlignedChildrenOf(p *Node)(bool) {
	if !is_children_of([]byte(n.Bytes), []byte(p.Bytes), n.End, p.End) {
		return false
	}
	if p.End == n.End {
		return true
	}
	return are_zero([]byte(n.Bytes), int(p.End) + 1, int(n.End))
}

/* Take the radix tree and a network
 * return true if leaf match and the list of nodes go throught
 */
func (r *Radix)LookupLonguestPath(data *[]byte, length int16)([]*Node) {
	var node *Node
	var path_node []*Node
	var end int16

	/* Browse tree */
	length-- /* convert length to index of last bit */
	path_node = make([]*Node, 0)
	node = r.Node
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
		if node.Data != nil {
			path_node = append(path_node, node)
		}

		/* If the node no match or we reach end of browsing, return data */
		if length <= node.End {
			return path_node
		}

		/* Continue browsing: get the value of next bit.  */
		end = node.End + 1
		if (*data)[end / 8] & (0x80 >> (end % 8)) != 0 {
			node = r.r2n(node.Right)
		} else {
			node = r.r2n(node.Left)
		}
	}
}

func (r *Radix)LookupLonguest(data *[]byte, length int16)(*Node) {
	var node *Node
	var last_node *Node
	var end int16

	/* Browse tree */
	length-- /* convert length to index of last bit */
	node = r.Node
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
		if node.Data != nil {
			last_node = node
		}

		/* We reach the end */
		if length <= end {
			return last_node
		}

		/* Continue browsing: get the value of next bit.  */
		end++
		if (*data)[end / 8] & (0x80 >> (end % 8)) != 0 {
			node = r.r2n(node.Right)
		} else {
			node = r.r2n(node.Left)
		}
	}
}

func (r *Radix)Get(data *[]byte, length int16)(*Node) {
	var n *Node
	n = r.LookupLonguest(data, length)
	if n == nil {
		return nil
	}
	if n.End + 1 != length {
		return nil
	}
	return n
}

func lookup_longuest_last_node(r *Radix, data []byte, length int16)(*Node) {
	var node *Node
	var end int16

	/* Browse tree */
	length-- /* convert length to index of last bit */
	node = r.Node
	for {

		/* If node is nil, return nil, otherwise, return the node
		 * because the tree match at least the first bit
		 */
		if node == nil || length <= node.End {
			return node
		}

		/* Perform bitcmp only if the input length is greater than current node length
		 * If the node match, continue browsing, otherwise return node.
		 */
		if node.End != -1 && !bitcmp([]byte(node.Bytes), data, node.Start, node.End) {
			return node
		}

		/* Continue browsing: get the value of next bit.  */
		end = node.End + 1
		if data[end / 8] & (0x80 >> (end % 8)) != 0 {
			if node.Right == null {
				return node
			}
			node = r.r2n(node.Right)
		} else {
			if node.Left == null {
				return node
			}
			node = r.r2n(node.Left)
		}
	}
}

/* Return nil is node is inserted, otherwise return existing node */
func (r *Radix)Insert(key *[]byte, length int16, data interface{})(*Node, bool) {
	var leaf *Node
	var node *Node
	var newnode *Node
	var bitno int16
	var l int16

	if length == 0 {
		return nil, false
	}

	/* Browse tree and return the closest node */
	node = lookup_longuest_last_node(r, *key, length)

	/* Create leaf node */
	leaf = r.node_alloc()
	leaf.Bytes = string(*key)
	leaf.Start = 0
	leaf.End = length - 1
	leaf.Parent = null
	leaf.Left = null
	leaf.Right = null
	leaf.Data = data

	/* CASE #1
	 *
	 * Special case, tree is empty, create node
	 */
	if node == nil {
		r.Node = leaf
		r.length++
		return leaf, true
	}

	/* The last node exact match the new entry */
	if length > node.End && bitcmp(*key, []byte(node.Bytes), node.Start, node.End) {

		/* CASE #2
		 *
		 * First, if we have a perfect match, just modify
		 * existing node.
		 *
		 * INSERT-KEY 0101 / 4
		 * STOP-NODE  0101 / 4
		 *
		 */
		if node.End == length - 1 {

			/* Unique mode is active and the data is set, return stored data */
			if node.Data != nil {
				return node, false
			}

			/* append data */
			node.Data = data
			r.length++
			return node, true
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
		leaf.Start = node.End + 1
		leaf.Parent = r.n2r(node)
		if bitget(*key, node.End + 1) == 1 {
			node.Right = r.n2r(leaf)
		} else {
			node.Left = r.n2r(leaf)
		}
		r.length++
		return leaf, true
	}

	/* Match the longuest part in the key */
	if leaf.End < node.End {
		l = leaf.End
	} else {
		l = node.End
	}
	bitno = bitlonguestmatch(*key, []byte(node.Bytes), node.Start, l)
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
		if node.Parent != null {
			leaf.Start = r.r2n(node.Parent).End + 1
		}
		leaf.Parent = node.Parent
		node.Parent = r.n2r(leaf)
		node.Start = leaf.End + 1

		/* Append existing nodes */
		if bitget([]byte(node.Bytes), node.Start) == 1 {
			leaf.Right = r.n2r(node)
			leaf.Left = null
		} else {
			leaf.Right = null
			leaf.Left = r.n2r(node)
		}

		/* Update original parent */
		if leaf.Parent == null {
			r.Node = leaf
		} else if r.r2n(leaf.Parent).Left == r.n2r(node) {
			r.r2n(leaf.Parent).Left = r.n2r(leaf)
		} else {
			r.r2n(leaf.Parent).Right = r.n2r(leaf)
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
	newnode.Start = node.Start
	newnode.End = bitno - 1
	newnode.Parent = node.Parent
	newnode.Data = nil

	/* Update existing node */
	node.Start = bitno
	node.Parent = r.n2r(newnode)

	/* Update leaf */
	leaf.Start = bitno
	leaf.Parent = r.n2r(newnode)

	/* Append existing nodes */
	if bitget(*key, bitno) == 1 {
		newnode.Right = r.n2r(leaf)
		newnode.Left = r.n2r(node)
	} else {
		newnode.Right = r.n2r(node)
		newnode.Left = r.n2r(leaf)
	}

	/* Update original parent */
	if newnode.Parent == null {
		r.Node = newnode
	} else if r.r2n(newnode.Parent).Left == r.n2r(node) {
		r.r2n(newnode.Parent).Left = r.n2r(newnode)
	} else {
		r.r2n(newnode.Parent).Right = r.n2r(newnode)
	}

	r.length++
	return leaf, true
}

func (r *Radix)Delete(n *Node) {
	var p *Node
	var c *Node

	/* Node WILL be delete, update accounting right now */
	if n.Data != nil {
		r.length--
	}

	/* If the node has two childs, just cleanup data */
	if n.Left != null && n.Right != null {
		n.Data = nil
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
			r.Node = c
			r.node_free(n)
			return
		}
		if r.r2n(n.Parent).Left == r.n2r(n) {
			r.r2n(n.Parent).Left = r.n2r(c)
		} else {
			r.r2n(n.Parent).Right = r.n2r(c)
		}
		r.node_free(n)
		return
	}

	/* If the node has no childs, just remove it. */
	if n.Left == null && n.Right == null {

		/* we reach root */
		if n.Parent == null {
			r.Node = nil
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
		if p.Data != nil {
			r.node_free(n)
			return
		}

		/* Remove the parent node */
		r.Delete(p)
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
func (r *Radix)Next(n *Node)(*Node) {
	var prev *Node
	for {
		if prev == r.r2n(n.Parent) || prev == nil {
			/* we come from parent, go left, right and them parent */
			prev = n
			if n.Left != null {
				n = r.r2n(n.Left)
			} else if n.Right != null {
				n = r.r2n(n.Right)
			} else if n.Parent != null {
				n = r.r2n(n.Parent)
			}
		} else if prev == r.r2n(n.Left) {
			/* we come from left branch, go right or go back */
			prev = n
			if n.Right != null {
				n = r.r2n(n.Right)
			} else if n.Parent != null {
				n = r.r2n(n.Parent)
			}
		} else if prev == r.r2n(n.Right) {
			/* we come from right branch, we go back */
			prev = n
			if n.Parent != null {
				n = r.r2n(n.Parent)
			}
		}

		/* None match, this is the end */
		if n == prev {
			return nil
		}

		/* If we reach leaf, and I'n not com from parent, return node */
		if n.Data != nil && prev == r.r2n(n.Parent) {
			return n
		}

		/* Otherwise continue browsing */
	}
}

func (r *Radix)First()(*Node) {
	if r.Node == nil {
		return nil
	}

	/* If entry node is a leaf, return it */
	if r.Node.Data != nil {
		return r.Node
	}

	/* Otherwise return next node */
	return r.Next(r.Node)
}

func (r *Radix)Last()(*Node) {
	var n *Node

	if r.Node == nil {
		return nil
	}

	/* Return previous leaf, if there are no previous
	 * node and the entry point is a leaf, return
	 * entry point.
	 */
	n = r.Node
	for {
		if n.Right != null {
			n = r.r2n(n.Right)
		} else if n.Left != null {
			n = r.r2n(n.Left)
		} else {
			return n
		}
	}
}

type Iter struct {
	node *Node
	next_node *Node
	key *[]byte
	length int16
	r *Radix
}

func (r *Radix)NewIter(key *[]byte, length int16)(*Iter) {
	var i *Iter

	i = &Iter{}
	i.key = key
	i.length = length
	i.r = r

	/* Lookup next node */
	if length == 0 {
		i.next_node = r.Node
	} else {
		i.next_node = lookup_longuest_last_node(r, *key, length)
		if i.next_node != nil && !is_children_of([]byte(i.next_node.Bytes), *i.key, i.next_node.End, i.length - 1) {
			i.next_node = nil
		}
	}

	/* No nodes found, next node is nil, abort iteration */
	if i.next_node == nil {
		return i
	}

	/* If the first node matching is a leaf, there is the entry point */
	if i.next_node.Data != nil {
		return i
	}

	/* Otherwise, lookup for next leaf */
	i.set_next()

	return i
}

func (i *Iter)set_next()() {
	if i.next_node == nil {
		return
	}
	i.next_node = i.r.Next(i.next_node)
	if i.next_node == nil {
		return
	}
	if i.length > 0 && !is_children_of([]byte(i.next_node.Bytes), *i.key, i.next_node.End, i.length - 1) {
		i.next_node = nil
	}
}

func (i *Iter)Next()(bool) {
	i.node = i.next_node
	i.set_next()
	return i.node != nil
}

func (i *Iter)Get()(*Node) {
	return i.node
}
