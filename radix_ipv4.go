// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "net"

func network_to_key(network *net.IPNet)([]byte, int) {
	var length int

	/* Get the network width. width of 0 id prohibited */
	length, _ = network.Mask.Size()
	return []byte(network.IP.To4()), length
}

func (r *Radix)IPv4LookupLonguest(network *net.IPNet)(*Node) {
	var length int
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = network_to_key(network)
	if length == 0 {
		return nil
	}

	/* Perform lookup */
	return lookup_longuest_last_match(r, &key, length)
}

func (r *Radix)IPv4LookupLonguestPath(network *net.IPNet)([]*Node) {
	var length int
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = network_to_key(network)
	if length == 0 {
		return make([]*Node, 0)
	}

	/* Perform lookup */
	return lookup_longuest_all_match(r, &key, length)
}

func (r *Radix)IPv4Get(network *net.IPNet)(*Node) {
	var length int
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = network_to_key(network)
	if length == 0 {
		return nil
	}

	/* Perform lookup */
	return lookup_longuest_exact_match(r, &key, length)
}

func (r *Radix)IPv4Insert(network *net.IPNet, data interface{})(interface{}) {
	var length int
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = network_to_key(network)
	if length == 0 {
		return nil
	}

	/* Perform insert */
	return insert(r, &key, length, data)
}

func (r *Radix)IPv4DeleteNetwork(network *net.IPNet)() {
	var node *Node
	var length int
	var key []byte

	/* Get the network width. width of 0 id prohibited */
	key, length = network_to_key(network)
	if length == 0 {
		return
	}

	/* Perform lookup */
	node = lookup_longuest_exact_match(r, &key, length)
	if node == nil {
		return
	}

	/* Delete entry */
	del(r, node)
}

func (r *Radix)IPv4Delete(n *Node)() {
	del(r, n)
}

func (n *Node)IPv4GetNet()(* net.IPNet) {
	var network *net.IPNet

	network = &net.IPNet{}
	network.Mask = net.CIDRMask(n.End + 1, 32)
	network.IP = net.IP(n.Bytes).Mask(network.Mask)

	return network
}
