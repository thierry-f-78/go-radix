// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "net"

func network_to_key(network *net.IPNet)([]byte, int16) {
	var l int
	var length int16

	/* Get the network width. width of 0 id prohibited */
	l, _ = network.Mask.Size()
	length = int16(l)
	return []byte(network.IP.To4()), length
}

// IPv4LookupLonguest get a ipv4 network and return the leaf which match the
// longest part of the prefix. Return nil if none match.
func (r *Radix)IPv4LookupLonguest(network *net.IPNet)(*Node) {
	var length int16
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = network_to_key(network)
	if length == 0 {
		return nil
	}

	/* Perform lookup */
	return r.LookupLonguest(&key, length)
}

// IPv4LookupLonguestPath take the radix tree and a ipv4 network, return the list
// of all leaf matching the prefix. If none match, return nil
func (r *Radix)IPv4LookupLonguestPath(network *net.IPNet)([]*Node) {
	var length int16
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = network_to_key(network)
	if length == 0 {
		return make([]*Node, 0)
	}

	/* Perform lookup */
	return r.LookupLonguestPath(&key, length)
}

// IPv4Get gets a ipv4 network and return exact match of the prefix. Exact match
// is a node wich match the prefix bit and the length.
func (r *Radix)IPv4Get(network *net.IPNet)(*Node) {
	var length int16
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = network_to_key(network)
	if length == 0 {
		return nil
	}

	/* Perform lookup */
	return r.Get(&key, length)
}

// IPv4Insert ipv4 network in the tree. The tree accept only unique value, if
// the prefix already exists in the tree, return existing leaf,
// otherwaise return nil.
func (r *Radix)IPv4Insert(network *net.IPNet, data interface{})(*Node, bool) {
	var length int16
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = network_to_key(network)
	if length == 0 {
		return nil, false
	}

	/* Perform insert */
	return r.Insert(&key, length, data)
}

// IPv4DeleteNetwork lookup network and remove it. does nothing
// if the network not exists.
func (r *Radix)IPv4DeleteNetwork(network *net.IPNet)() {
	var node *Node
	var length int16
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = network_to_key(network)
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

// IPv4GetNet convert node key/length prefix to IPv4 network data
func (n *Node)IPv4GetNet()(* net.IPNet) {
	var network *net.IPNet

	network = &net.IPNet{}
	network.Mask = net.CIDRMask(int(n.node.End) + 1, 32)
	network.IP = net.IP(n.node.Bytes).Mask(network.Mask)

	return network
}

// IPv4NewIter return struct Iter for browsing all nodes there children
// match the ipv4 network
func (r *Radix)IPv4NewIter(network *net.IPNet)(*Iter) {
	var length int16
	var key []byte

	key, length = network_to_key(network)
	return r.NewIter(&key, length)
}
