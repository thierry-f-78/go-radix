// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "net"
import "math/rand"
import "testing"
import "time"

func TestRadixIPv4Order(t *testing.T) {
	var r *Radix
	var reference []string
	var pfx []string
	var s string
	var n *net.IPNet
	var a *Node
	var index int

	/* This is a sorted reference of networks */
	reference = []string{
		"10.0.0.0/8",
		"10.0.0.0/9",
		"10.0.0.0/10",
		"10.0.0.0/16",
		"10.0.0.0/24",
		"10.0.0.0/32",
		"10.8.0.0/16",
		"10.8.0.0/24",
		"10.14.0.0/16",
		"10.127.3.0/24",
		"10.128.0.0/16",
	}

	copy(pfx, reference)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(pfx), func(i, j int) { pfx[i], pfx[j] = pfx[j], pfx[i] })

	r = NewRadix()
	for _, s = range pfx {
		_, n, _ = net.ParseCIDR(s)
		r.IPv4Insert(n, "")
	}

	index = 0
	for a = r.First(); a != nil; a = r.Next(a) {
		// println(a.IPv4GetNet().String())
		if a.IPv4GetNet().String() != reference[index] {
			t.Errorf("something is wrong in sort order at index %d, expect %q, got %q",
			         index, reference[index], a.IPv4GetNet().String())
		}
		index++
	}
}

func TestRadixIPv4(t *testing.T) {
	var nw1 *net.IPNet
	var nw2 *net.IPNet
	var nw3 *net.IPNet
	var r *Radix
	var n *Node
	var ns []*Node
	var s string

	/* Init DB */

	r = NewRadix()

	/* Insert network */

	nw1 = &net.IPNet{}
	nw1.IP = net.ParseIP("10.4.0.0")
	nw1.Mask = net.CIDRMask(16, 32)
	r.IPv4Insert(nw1, "test - 10.4.0.0/16")

	/* Lookup network */

	n = r.IPv4Get(nw1)
	if n == nil {
		t.Errorf("Should match")
	}
	nw2 = n.IPv4GetNet()
	if !nw2.IP.Equal(nw1.IP) {
		t.Errorf("Should match")
	}
	s = n.Data.(string)
	if s != "test - 10.4.0.0/16" {
		t.Errorf("Should match")
	}

	/* Lookup longest */

	nw2 = &net.IPNet{}
	nw2.IP = net.ParseIP("10.4.0.0")
	nw2.Mask = net.CIDRMask(32, 32)

	n = r.IPv4LookupLonguest(nw2)
	if n == nil {
		t.Errorf("Should match")
	}
	nw2 = n.IPv4GetNet()
	if !nw2.IP.Equal(nw1.IP) {
		t.Errorf("Should match")
	}
	s = n.Data.(string)
	if s != "test - 10.4.0.0/16" {
		t.Errorf("Should match")
	}

	/* Insert parent network */

	nw3 = &net.IPNet{}
	nw3.IP = net.ParseIP("10.0.0.0")
	nw3.Mask = net.CIDRMask(8, 32)
	r.IPv4Insert(nw3, "test - 10.0.0.0/8")

	/* Lookup longest */

	nw2 = &net.IPNet{}
	nw2.IP = net.ParseIP("10.4.0.0")
	nw2.Mask = net.CIDRMask(32, 32)

	n = r.IPv4LookupLonguest(nw2)
	if n == nil {
		t.Errorf("Should match")
	}
	nw2 = n.IPv4GetNet()
	if !nw2.IP.Equal(nw1.IP) {
		t.Errorf("Should match")
	}
	s = n.Data.(string)
	if s != "test - 10.4.0.0/16" {
		t.Errorf("Should match")
	}

	/* Lookup longest path */

	nw2 = &net.IPNet{}
	nw2.IP = net.ParseIP("10.4.0.0")
	nw2.Mask = net.CIDRMask(32, 32)

	ns = r.IPv4LookupLonguestPath(nw2)
	if len(ns) != 2 {
		t.Errorf("Should have length of 2, got %d", len(ns))
	}

	nw2 = ns[0].IPv4GetNet()
	if !nw2.IP.Equal(nw2.IP) {
		t.Errorf("Should match")
	}
	s = ns[0].Data.(string)
	if s != "test - 10.0.0.0/8" {
		t.Errorf("Should match")
	}

	nw2 = ns[1].IPv4GetNet()
	if !nw2.IP.Equal(nw2.IP) {
		t.Errorf("Should match")
	}
	s = ns[1].Data.(string)
	if s != "test - 10.4.0.0/16" {
		t.Errorf("Should match")
	}

	/* Delete */

	r.IPv4DeleteNetwork(nw1)

	/* Lookup deleted */

	n = r.IPv4Get(nw1)
	if n != nil {
		t.Errorf("Should not match")
	}

	/* insert network and browse */

	r = NewRadix()

	nw1 = &net.IPNet{}
	nw1.IP = net.ParseIP("10.4.0.0")
	nw1.Mask = net.CIDRMask(32, 32)
	r.IPv4Insert(nw1, "10.4.0.0/32")

	nw1 = &net.IPNet{}
	nw1.IP = net.ParseIP("10.4.0.0")
	nw1.Mask = net.CIDRMask(16, 32)
	r.IPv4Insert(nw1, "10.4.0.0/16")

	nw1 = &net.IPNet{}
	nw1.IP = net.ParseIP("10.4.0.0")
	nw1.Mask = net.CIDRMask(24, 32)
	r.IPv4Insert(nw1, "10.4.0.0/24")

	println("Browse forward")
	for n = r.First(); n != nil; n = r.Next(n) {
		println(n.IPv4GetNet().String())
	}

	println("fin")
}

func TestRadixIPv4DS(t *testing.T) {
	var s string
	var r *Radix
	var ipn *net.IPNet
	var err error
	var n *Node
	var entered int
	var count int

	var load_networks []string = []string{
		"1.0.0.0/24",
		"1.0.4.0/22",
		"1.0.16.0/24",
		"1.0.64.0/18",
		"1.0.128.0/17",
		"1.0.128.0/24", /* <- this insert caused error */
		"1.0.129.0/24",
		"1.0.130.0/24",
		"1.0.131.0/24",
		"1.0.132.0/22",
		"1.0.136.0/22",
		"1.0.141.0/24",
		"1.0.142.0/23",
		"1.0.144.0/20",
		"1.0.164.0/22",
		"1.0.168.0/21",
		"1.0.192.0/20",
		"1.0.208.0/22",
		"1.0.212.0/23",
		"1.0.214.0/24",
	}

	r = NewRadix()

	for _, s = range load_networks {

		_, ipn, err = net.ParseCIDR(s)
		if err != nil {
			panic(err)
		}
		r.IPv4Insert(ipn, s)

		entered++
		if entered != r.Len() {
			t.Errorf("entered %d != len %d", entered, r.Len())
		}

		count = 0
		for n = r.First(); n != nil; n = r.Next(n) {
			count++
		}
		if entered != count {
			t.Errorf("entered %d != count %d", entered, count)
		}
	}

	/* browse */
	count = 0
	for n = r.First(); n != nil; n = r.Next(n) {
		if n.IPv4GetNet().String() != load_networks[count] {
			t.Errorf("Expect network %s, founs %s", load_networks[count], n.IPv4GetNet().String())
		}
		count++;
	}
}
