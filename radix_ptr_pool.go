// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "unsafe"

const kind_node = 0

type chunk struct {
	nodes [65536]Node
	ptr uintptr
}

type node_pool struct {
	free int
	capacity int
	pool []*chunk
	next uint32
}

type ptr_range struct {
	start uintptr
	end uintptr
	index int
	kind int
}

func (r *Radix)node_alloc()(*Node) {
	var n *Node

	if r.node.free == 0 {
		r.growth()
	}
	r.node.free--
	n = r.r2n(r.node.next)
	r.node.next = n.Left
	return n
}

func (r *Radix)node_free(n *Node)() {
	n.Data = nil
	n.Bytes = ""
	n.Parent = null
	n.Left = r.node.next
	n.Right = null
	r.node.free++
	r.node.next = r.n2r(n)
}

/* reference to node */
func (r *Radix)r2n(v uint32)(*Node) {
	if v == null {
		return nil
	}
	return &r.node.pool[v >> 16].nodes[v & 0xffff]
}

func (r *Radix)growth() {
	var c *chunk
	var i int

	if len(r.node.pool) >= 32768 {
		panic("reach the maximum number of node pools allowed")
	}
	c = &chunk{}
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

/* if insert a range which overlap existing range, it panic */
func (r *Radix)add_range(start uintptr, end uintptr, index int, kind int) {
	var left int
	var right int
	var pivot int
	var insert_data []ptr_range

	if start >= end {
		return
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
			return
		}
		if left > right {
			return
		}
	}
}

/* Give a node and return its reference. If the node is not
 * from a local pool return 0.
 */
func (r *Radix)n2r(n *Node)(uint32) {
	var left int
	var right int
	var pivot int
	var index int
	var p uintptr
	var c *chunk
	var k uint32
	var kind int

	p = uintptr(unsafe.Pointer(n))
	right = len(r.ptr_range)
	if right == 0 {
		panic("e")
		return 0
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
		panic("g")
			return 0
		}
	}
	c = r.node.pool[index]
	if kind == kind_node {
		k = 0x00000000
	} else {
		k = 0x80000000
	}
	return k | (uint32(index) << 16) | (uint32(p - c.ptr) / node_sz)
}
