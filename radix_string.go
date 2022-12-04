// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

func string_to_key(str string)([]byte, int) {
	return []byte(str), len(str) * 8
}

func (r *Radix)StringLookupLonguest(str string)(*Node) {
	var length int
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = string_to_key(str)
	if length == 0 {
		return nil
	}

	/* Perform lookup */
	return r.LookupLonguest(&key, length)
}

func (r *Radix)StringLookupLonguestPath(str string)([]*Node) {
	var length int
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = string_to_key(str)
	if length == 0 {
		return make([]*Node, 0)
	}

	/* Perform lookup */
	return r.LookupLonguestPath(&key, length)
}

func (r *Radix)StringGet(str string)(*Node) {
	var length int
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = string_to_key(str)
	if length == 0 {
		return nil
	}

	/* Perform lookup */
	return r.Get(&key, length)
}

func (r *Radix)StringInsert(str string, data interface{})(*Node, bool) {
	var length int
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = string_to_key(str)
	if length == 0 {
		return nil, false
	}

	/* Perform insert */
	return r.Insert(&key, length, data)
}

func (r *Radix)StringDelete(str string)() {
	var node *Node
	var length int
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

func (r *Radix)StringNewIter(str string)(*Iter) {
	var length int
	var key []byte

	key, length = string_to_key(str)
	return r.NewIter(&key, length)
}

func (n *Node)StringGetKey()(string) {
	return string(n.Bytes)
}
