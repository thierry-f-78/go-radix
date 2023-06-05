// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "unsafe"

type ptr_range struct {
	start uintptr
	end uintptr
	pool_index int
}

/* if insert a range which overlap existing range, it panic */
func (r *Radix)add_range(start uintptr, end uintptr, pool_index int) {
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
		pool_index: pool_index,
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
	var pool_index int
	var p uintptr
	var c *chunk

	p = uintptr(unsafe.Pointer(n))
	right = len(r.ptr_range)
	if right == 0 {
		panic("e")
		return 0
	}
	for {
		if left == right {
			pool_index = r.ptr_range[left].pool_index
			break
		}
		pivot = (left + right) / 2
		if p < r.ptr_range[pivot].start {
			right = pivot
		} else if p > r.ptr_range[pivot].end {
			left = pivot + 1
		} else {
			pool_index = r.ptr_range[pivot].pool_index
			break
		}
		if left > right {
		panic("g")
			return 0
		}
	}
	c = r.pool[pool_index]
	return (uint32(pool_index) << 16) | (uint32(p - c.ptr) / node_sz)
}
