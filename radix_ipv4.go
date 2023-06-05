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

func (n *Node)IPv4GetNet()(* net.IPNet) {
	var network *net.IPNet

	network = &net.IPNet{}
	network.Mask = net.CIDRMask(int(n.node.End) + 1, 32)
	network.IP = net.IP(n.node.Bytes).Mask(network.Mask)

	return network
}

func (r *Radix)IPv4NewIter(network *net.IPNet)(*Iter) {
	var length int16
	var key []byte

	key, length = network_to_key(network)
	return r.NewIter(&key, length)
}
