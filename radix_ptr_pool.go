// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "unsafe"

const kind_node = 0
const kind_leaf = 1

type node_chunk struct {
	nodes [65536]node
	ptr uintptr
}

type node_pool struct {
	free int
	capacity int
	pool []*node_chunk
	next uint32
}

type leaf_chunk struct {
	nodes [65536]Node
	ptr uintptr
}

type leaf_pool struct {
	free int
	capacity int
	pool []*leaf_chunk
	next uint32
}

type ptr_range struct {
	start uintptr
	end uintptr
	index int
	kind int
}

func is_leaf(ref uint32)(bool) {
	return (ref & 0x80000000) != 0
}

func (r *Radix)node_alloc()(*node) {
	var n *node

	if r.node.free == 0 {
		r.node_growth()
	}
	r.node.free--
	n = r.r2n(r.node.next)
	r.node.next = n.Left
	return n
}

func (r *Radix)leaf_alloc()(*Node) {
	var n *Node

	if r.leaf.free == 0 {
		r.leaf_growth()
	}
	r.leaf.free--
	n = n2N(r.r2n(r.leaf.next))
	r.leaf.next = n.node.Left
	return n
}

func (r *Radix)free(n *node)() {
	var leaf *Node

	if is_leaf(r.n2r(n)) {
		leaf = n2N(n)
		leaf.Data = nil
		leaf.node.Bytes = ""
		leaf.node.Parent = null
		leaf.node.Left = r.leaf.next
		leaf.node.Right = null
		r.leaf.free++
		r.leaf.next = r.n2r(&leaf.node)
	} else {
		n.Bytes = ""
		n.Parent = null
		n.Left = r.node.next
		n.Right = null
		r.node.free++
		r.node.next = r.n2r(n)
	}
}

/* reference to node */
func (r *Radix)r2n(v uint32)(*node) {
	if v == null {
		return nil
	}
	if is_leaf(v) {
		return &r.leaf.pool[(v >> 16) & 0x7fff].nodes[v & 0xffff].node
	} else {
		return &r.node.pool[v >> 16].nodes[v & 0xffff]
	}
}

func (r *Radix)node_growth() {
	var c *node_chunk
	var i int

	if len(r.node.pool) >= 32768 {
		panic("reach the maximum number of node pools allowed")
	}
	c = &node_chunk{}
	c.ptr = (uintptr)(unsafe.Pointer(&c.nodes[0]))
	r.node.pool = append(r.node.pool, c)
	r.node.free += 65536
	r.node.capacity += 65536
	r.add_range(c.ptr, (uintptr)(unsafe.Pointer(&c.nodes[65536 - 1])), len(r.node.pool) - 1, kind_node)
	for i, _ = range c.nodes {
		/* first node of the first list the NULL node, so it never be used.
		 * to make the code simpler, it is allocated, but it is never set
		 * in the free nodes list
		 */
		if len(r.node.pool) == 1 && i == 0 {
			r.node.free--
			r.node.capacity--
			continue
		}
		c.nodes[i].Left = r.node.next
		r.node.next = r.n2r(&c.nodes[i])
	}
}

func (r *Radix)leaf_growth() {
	var c *leaf_chunk
	var i int

	if len(r.leaf.pool) >= 32768 {
		panic("reach the maximum number of node pools allowed")
	}
	c = &leaf_chunk{}
	c.ptr = (uintptr)(unsafe.Pointer(&c.nodes[0]))
	r.leaf.pool = append(r.leaf.pool, c)
	r.leaf.free += 65536
	r.leaf.capacity += 65536
	r.add_range(c.ptr, (uintptr)(unsafe.Pointer(&c.nodes[65536 - 1])), len(r.leaf.pool) - 1, kind_leaf)
	for i, _ = range c.nodes {
		c.nodes[i].node.Left = r.leaf.next
		r.leaf.next = r.n2r(&c.nodes[i].node)
	}
}

/* if insert a range which overlap existing range, it panic */
func (r *Radix)add_range(start uintptr, end uintptr, index int, kind int) {
	var left int
	var right int
	var pivot int
	var insert_data []ptr_range

	if start >= end {
		panic("start >= end")
	}

	insert_data = []ptr_range{ptr_range{
		start: start,
		end: end,
		index: index,
		kind: kind,
	}}

	right = len(r.ptr_range)
	for {
		if left == right {
			r.ptr_range = append(r.ptr_range[0:left], append(insert_data, r.ptr_range[left:]...)...)
			return
		}
		pivot = (left + right) / 2
		if end < r.ptr_range[pivot].start {
			right = pivot
		} else if start > r.ptr_range[pivot].end {
			left = pivot + 1
		} else {
			panic("cannot insert range")
		}
		if left > right {
			panic("cannot insert range")
		}
	}
}

/* Give a node and return its reference. If the node is not
 * from a local pool return 0.
 */
func (r *Radix)n2r(n *node)(uint32) {
	var left int
	var right int
	var pivot int
	var index int
	var p uintptr
	var cn *node_chunk
	var cl *leaf_chunk
	var kind int

	p = uintptr(unsafe.Pointer(n))
	right = len(r.ptr_range)
	if right == 0 {
		panic("unknown ref")
	}
	for {
		if left == right {
			index = r.ptr_range[left].index
			kind = r.ptr_range[left].kind
			break
		}
		pivot = (left + right) / 2
		if p < r.ptr_range[pivot].start {
			right = pivot
		} else if p > r.ptr_range[pivot].end {
			left = pivot + 1
		} else {
			index = r.ptr_range[pivot].index
			kind = r.ptr_range[pivot].kind
			break
		}
		if left > right {
			panic("unknown ref")
		}
	}
	if kind == kind_node {
		cn = r.node.pool[index]
		return (uint32(index) << 16) | (uint32(p - cn.ptr) / node_sz)
	} else {
		cl = r.leaf.pool[index]
		return 0x80000000 | (uint32(index) << 16) | (uint32(p - cl.ptr) / leaf_sz)
	}
}
