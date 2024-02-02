// Copyright (C) 2024 Thierry Fournier <tfournier@arpalert.org>

package radix

import "encoding/binary"
import "time"

const time_length = 64

func time_to_key(value time.Time)([]byte) {
	var bytes [8]byte

	binary.BigEndian.PutUint64(bytes[:], uint64(value.UnixMicro()))
	return bytes[:]
}

// TimeGet gets a time.Time prefix and return exact match of the prefix. Exact match
// is a node wich match the prefix bit and the length. Note the tree precision is microsecond
func (r *Radix)TimeGet(value time.Time)(*Node) {
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key = time_to_key(value)

	/* Perform lookup */
	return r.Get(&key, time_length)
}

// TimeLookupLonguest get the bigger time close from the key Return nil if none match.
func (r *Radix)TimeLookupLonguestGe(value time.Time)(*Node) {
	var key []byte
	var n *Node
	var date time.Time

	/* Get the network width. width of 0 id prohibited */
	key = time_to_key(value)

	/* Perform lookup */
	n = r.LookupLonguest(&key, time_length)
	if n == nil {
		return nil
	}
	date = n.TimeGetValue()
	if date.Equal(value) || date.After(value) {
		return n
	}
	return r.Next(n)
}

// TimeInsert time.Time prefix in the tree. The tree accept only unique value, if
// the prefix already exists in the tree, return existing leaf,
// otherwise return nil. Note the tree precision is microsecond
func (r *Radix)TimeInsert(value time.Time, data interface{})(*Node, bool) {
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key = time_to_key(value)

	/* Perform insert */
	return r.Insert(&key, time_length, data)
}

// TimeDelete lookup time.Time and remove it. does nothing
// if the network not exists. Note the tree precision is microsecond
func (r *Radix)TimeDelete(value time.Time)() {
	var node *Node
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key = time_to_key(value)

	/* Perform lookup */
	node = r.Get(&key, time_length)
	if node == nil {
		return
	}

	/* Delete entry */
	r.Delete(node)
}

// TimeGetValue convert node key/length prefix to time.Time data. Note the
// tree precision is microsecond
func (n *Node)TimeGetValue()(time.Time) {
	if len(n.node.Bytes) != 8 {
		return time.Time{}
	}
	return time.UnixMicro(int64(binary.BigEndian.Uint64([]byte(n.node.Bytes))))
}

// UInt64NewIter return struct Iter for browsing all nodes there children
// match the key/length prefix. Note the tree precision is microsecond
func (r *Radix)TimeNewIter(value time.Time)(*Iter) {
	var key []byte

	key = time_to_key(value)
	return r.NewIter(&key, time_length)
}
