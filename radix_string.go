// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

func string_to_key(str string)([]byte, int16) {
	return []byte(str), int16(len(str)) * 8
}

// StringLookupLonguest get a string as prefix and return the leaf which match the
// longest part of the prefix. Return nil if none match.
func (r *Radix)StringLookupLonguest(str string)(*Node) {
	var length int16
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = string_to_key(str)
	if length == 0 {
		return nil
	}

	/* Perform lookup */
	return r.LookupLonguest(&key, length)
}

// StringLookupLonguestPath take the radix tree and a string as prefix, return the list
// of all leaf matching the prefix. If none match, return nil
func (r *Radix)StringLookupLonguestPath(str string)([]*Node) {
	var length int16
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = string_to_key(str)
	if length == 0 {
		return make([]*Node, 0)
	}

	/* Perform lookup */
	return r.LookupLonguestPath(&key, length)
}

// Get gets a string as prefix and return exact match of the prefix. Exact match
// is a node wich match the prefix bit and the length.
func (r *Radix)StringGet(str string)(*Node) {
	var length int16
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = string_to_key(str)
	if length == 0 {
		return nil
	}

	/* Perform lookup */
	return r.Get(&key, length)
}

// StringInsert string as prefix in the tree. The tree accept only unique value, if
// the prefix already exists in the tree, return existing leaf,
// otherwaise return nil.
func (r *Radix)StringInsert(str string, data interface{})(*Node, bool) {
	var length int16
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = string_to_key(str)
	if length == 0 {
		return nil, false
	}

	/* Perform insert */
	return r.Insert(&key, length, data)
}

// StringDelete lookup string and remove it. does nothing
// if the string not exists.
func (r *Radix)StringDelete(str string)() {
	var node *Node
	var length int16
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = string_to_key(str)
	if length == 0 {
		return
	}

	/* Perform lookup */
	node = r.Get(&key, length)
	if node == nil {
		return
	}

	/* Delete entry */
	r.Delete(node)
}

// StringNewIter return struct Iter for browsing all nodes there children
// match the string prefix.
func (r *Radix)StringNewIter(str string)(*Iter) {
	var length int16
	var key []byte

	key, length = string_to_key(str)
	return r.NewIter(&key, length)
}

// StringGetKey convert node key/length prefix to string
func (n *Node)StringGetKey()(string) {
	return string(n.node.Bytes)
}
