// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "fmt"

/* This is a tree node. */
type Node struct {
	Bytes []byte /* slice of bytes for IPv4 */
	Start int /* the first representative bit in this node */
	End int /* the first non-representative bit in this node */
	Parent *Node
	Left *Node
	Right *Node
	Data []interface{} /* Conatins the list of interface matching the node */
}

type Radix struct {
	Node *Node
	Unique bool
}

func NewRadix(unique bool)(*Radix) {
	var radix *Radix

	radix = &Radix{}
	radix.Unique = unique

	return radix
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

	for _, b = range n.Bytes {
		out = fmt.Sprintf("%s0x%02x ", out, b)
	}
	return fmt.Sprintf("%s/ %d, mode=%s, Left=%p, Right=%p, Parent=%p",
	                   out, n.End + 1, mode, n.Left, n.Right, n.Parent)
}

/* Take the radix tree and a network
 * return true if leaf match and the list of nodes go throught
 */
func lookup_longuest_all_match(r *Radix, data *[]byte, length int)([]*Node) {
	var node *Node
	var path_node []*Node
	var end int

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
		if node.End != -1 && !bitcmp(&node.Bytes, data, node.Start, node.End) {
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
			node = node.Right
		} else {
			node = node.Left
		}
	}
}

func lookup_longuest_last_match(r *Radix, data *[]byte, length int)(*Node) {
	var node *Node
	var last_node *Node
	var end int

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
		if length < end || (end != -1 && !bitcmp(&node.Bytes, data, node.Start, end)) {
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
			node = node.Right
		} else {
			node = node.Left
		}
	}
}

func lookup_longuest_exact_match(r *Radix, data *[]byte, length int)(*Node) {
	var n *Node
	n = lookup_longuest_last_match(r, data, length)
	if n == nil {
		return nil
	}
	if n.End + 1 != length {
		return nil
	}
	return n
}

func lookup_longuest_last_node(r *Radix, data *[]byte, length int)(*Node) {
	var node *Node
	var end int

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
		if node.End != -1 && !bitcmp(&node.Bytes, data, node.Start, node.End) {
			return node
		}

		/* Continue browsing: get the value of next bit.  */
		end = node.End + 1
		if (*data)[end / 8] & (0x80 >> (end % 8)) != 0 {
			node = node.Right
		} else {
			node = node.Left
		}
	}
}

func insert(r *Radix, message *[]byte, length int, data interface{})(interface{}) {
	var leaf *Node
	var node *Node
	var newnode *Node
	var bitno int
	var l int

	if length == 0 {
		return nil
	}

	/* Browse tree and return the closest node */
	node = lookup_longuest_last_node(r, message, length)

	/* Create leaf node */
	leaf = &Node{}
	leaf.Bytes = make([]byte, len(*message))
	copy(leaf.Bytes, *message)
	leaf.Start = 0
	leaf.End = length - 1
	leaf.Parent = nil
	leaf.Left = nil
	leaf.Right = nil
	leaf.Data = append(leaf.Data, data)

	/* CASE #1
	 *
	 * Special case, tree is empty, create node
	 */
	if node == nil {
		r.Node = leaf
		return data
	}

	/* The last node exact match the new entry */
	if bitcmp(message, &node.Bytes, node.Start, node.End) {

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

			/* Not a leaf, convert it */
			if node.Data != nil {
				node.Data = append(node.Data, data)
				return data
			}

			/* Unique mode is active and the data is set, return stored data */
			if r.Unique && node.Data != nil && len(node.Data) == 1 {
				return node.Data[0]
			}

			/* append data */
			node.Data = append(node.Data, data)
			return data
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
		leaf.Parent = node
		if bitget(message, node.End + 1) == 1 {
			node.Right = leaf
		} else {
			node.Left = leaf
		}
		return data
	}

	/* Match the longuest part in the key */
	if leaf.End < node.End {
		l = leaf.End
	} else {
		l = node.End
	}
	bitno = bitlonguestmatch(message, &node.Bytes, node.Start, l)
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
		if node.Parent != nil {
			leaf.Start = node.Parent.End + 1
		}
		leaf.Parent = node.Parent
		node.Parent = leaf
		node.Start = leaf.End + 1

		/* Append existing nodes */
		if bitget(&node.Bytes, bitno) == 1 {
			leaf.Right = node
			leaf.Left = nil
		} else {
			leaf.Right = nil
			leaf.Left = node
		}

		/* Update original parent */
		if leaf.Parent == nil {
			r.Node = leaf
		} else if leaf.Parent.Left == node {
			leaf.Parent.Left = leaf
		} else {
			leaf.Parent.Right = leaf
		}

		return data
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
	newnode = &Node{}
	newnode.Bytes = make([]byte, len(*message))
	copy(newnode.Bytes, *message)
	newnode.Start = node.Start
	newnode.End = bitno - 1
	newnode.Parent = node.Parent
	newnode.Data = nil

	/* Update existing node */
	node.Start = bitno
	node.Parent = newnode

	/* Update leaf */
	leaf.Start = bitno
	leaf.Parent = newnode

	/* Append existing nodes */
	if bitget(message, bitno) == 1 {
		newnode.Right = leaf
		newnode.Left = node
	} else {
		newnode.Right = node
		newnode.Left = leaf
	}

	/* Update original parent */
	if newnode.Parent == nil {
		r.Node = newnode
	} else if newnode.Parent.Left == node {
		newnode.Parent.Left = newnode
	} else {
		newnode.Parent.Right = newnode
	}

	return data
}

func (r *Radix)Delete(n *Node) {
	var p *Node
	var c *Node

	/* If the node has two childs, just cleanup data */
	if n.Left != nil && n.Right != nil {
		n.Data = nil
		return
	}

	/* If node has one child. Remove the node, and
	 * Link the child to the parent. Change child bits
	 */
	if (n.Left == nil) != (n.Right == nil) {
		if n.Left != nil {
			c = n.Left
		} else {
			c = n.Right
		}
		c.Start = n.Start
		c.Parent = n.Parent
		if n.Parent == nil {
			r.Node = c
			return
		}
		if n.Parent.Left == n {
			n.Parent.Left = c
		} else {
			n.Parent.Right = c
		}
		return
	}

	/* If the node has no childs, just remove it. */
	if n.Left == nil && n.Right == nil {

		/* we reach root */
		if n.Parent == nil {
			r.Node = nil
			return
		}

		/* Remove my branch on the parent node */
		p = n.Parent
		if p.Left == n {
			p.Left = nil
		} else if p.Right == n {
			p.Right = nil
		}

		/* if the parent node is a leaf, do not remove */
		if p.Data != nil {
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
func (n *Node)Next()(*Node) {
	var prev *Node
	for {
		if prev == n.Parent || prev == nil {
			/* we come from parent, go left, right and them parent */
			prev = n
			if n.Left != nil {
				n = n.Left
			} else if n.Right != nil {
				n = n.Right
			} else if n.Parent != nil {
				n = n.Parent
			}
		} else if prev == n.Left {
			/* we come from left branch, go right or go back */
			prev = n
			if n.Right != nil {
				n = n.Right
			} else if n.Parent != nil {
				n = n.Parent
			}
		} else if prev == n.Right {
			/* we come from right branch, we go back */
			prev = n
			if n.Parent != nil {
				n = n.Parent
			}
		}

		/* None match, this is the end */
		if n == prev {
			return nil
		}

		/* If we reach leaf, return node */
		if n.Data != nil {
			return n
		}

		/* Otherwise continue browsing */
	}
}

func (n *Node)Prev()(*Node) {
	var prev *Node
	for {
		if prev == n.Parent || prev == nil {
			/* we come from parent, go right, go left or go back */
			prev = n
			if n.Right != nil {
				n = n.Right
			} else if n.Left != nil {
				n = n.Left
			} else if n.Parent != nil {
				n = n.Parent
			}
		} else if prev == n.Right {
			/* we come from left branch, go left or go back */
			prev = n
			if n.Left != nil {
				n = n.Left
			} else if n.Parent != nil {
				n = n.Parent
			}
		} else if prev == n.Left {
			/* we come from right branch, we go back */
			prev = n
			if n.Parent != nil {
				n = n.Parent
			}
		}

		/* None match, this is the end */
		if n == prev {
			return nil
		}

		/* If we reach leaf, return node */
		if n.Data != nil {
			return n
		}

		/* Otherwise continue browsing */
	}
}

func (r *Radix)First()(*Node) {
	var n *Node

	if r.Node == nil {
		return nil
	}

	n = r.Node.Next()
	if n == nil {
		if r.Node.Data != nil {
			return r.Node
		}
		return nil
	}
	return n
}

func (r *Radix)Last()(*Node) {
	var n *Node

	if r.Node == nil {
		return nil
	}

	n = r.Node.Prev()
	if n == nil {
		if r.Node.Data != nil {
			return r.Node
		}
		return nil
	}
	return n
}

type Iter struct {
	node *Node
	next_node *Node
	key *[]byte
	length int
	forward bool
}

func (radix *Radix)NewIter(key *[]byte, length int, forward bool)(*Iter) {
	var i *Iter

	i = &Iter{}
	i.key = key
	i.length = length
	i.forward = forward

	/* Lookup next node */
	if length == 0 {
		i.next_node = radix.Node
	} else {
		i.next_node = lookup_longuest_last_node(radix, key, length)
		if i.next_node != nil && !is_children_of(&i.next_node.Bytes, i.key, i.next_node.End, i.length - 1) {
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
	if i.forward {
		i.next_node = i.next_node.Next()
	} else {
		i.next_node = i.next_node.Prev()
	}
	if i.next_node == nil {
		return
	}
	if i.length > 0 && !is_children_of(&i.next_node.Bytes, i.key, i.next_node.End, i.length - 1) {
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
