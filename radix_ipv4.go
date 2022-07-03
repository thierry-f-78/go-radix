// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "net"

func (r *Radix)IPv4LookupLonguest(network *net.IPNet)([]*interface{}) {
	var node *Node
	var length int
	var message []byte

	/* Get the network width. width of 0 id prohibited */
	length, _ = network.Mask.Size()
	if length == 0 {
		return nil
	}

	/* Get IP and convert it to byte array */
	message = []byte(network.IP)

	/* Perform lookup */
	node = lookup_longuest_last_match(r, &message, length)
	if node == nil {
		return nil
	}
	return node.Data
}

func (r *Radix)IPv4LookupLonguestPath(network *net.IPNet)([][]*interface{}) {
	var nodes []*Node
	var data [][]*interface{}
	var index int
	var length int
	var message []byte

	/* Get the network width. width of 0 id prohibited */
	length, _ = network.Mask.Size()
	if length == 0 {
		return make([][]*interface{}, 0)
	}

	/* Get IP and convert it to byte array */
	message = []byte(network.IP)

	/* Perform lookup */
	nodes = lookup_longuest_all_match(r, &message, length)
	for index, _ = range nodes {
		data = append(data, nodes[index].Data)
	}

	return data
}

func (r *Radix)IPv4Insert(network *net.IPNet, data *interface{})(*interface{}) {
	var length int
	var message []byte

	/* Get the network width. width of 0 id prohibited */
	length, _ = network.Mask.Size()
	if length == 0 {
		return nil
	}

	/* Get IP and convert it to byte array */
	message = []byte(network.IP)

	return insert(r, &message, length, data)
}
