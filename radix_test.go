// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "compress/gzip"
import "encoding/binary"
import "bufio"
import "fmt"
import "math/rand"
import "net"
import "os"
import "strconv"
import "strings"
import "testing"
import "time"

func TestRadix(t *testing.T) {
	/*
	var r *Radix
	var b []byte
	var n []*Node

	r = NewRadix(true)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00010101}
	r.Insert(b, 32, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00010011}
	r.Insert(b, 32, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00010000}
	r.Insert(b, 29, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00000101}
	r.Insert(b, 32, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00000101}
	r.Insert(b, 32, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00001100}
	r.Insert(b, 30, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00001100}
	r.Insert(b, 32, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00001000}
	r.Insert(b, 29, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00010101}
	n = lookup_longuest(r, b, 32, true)
	fmt.Printf("%+v\n", n)
	*/
}

type ent struct {
	b []byte
	l int16
}

func Benchmark_Radix(t *testing.B) {
	var r *Radix
	var file *os.File
	var scanner *bufio.Scanner
	var err error
	var tokens []string
	var ip uint32
	var bytes []byte
	var int_dec int64
	var now time.Time
	var step time.Time
	var count int
	var list []ent
	var i int
	var ent ent
	var rounds = 1000000
	var node *Node
	var ip3 net.IPNet
	var ip2 net.IPNet
	var b []byte
	var hit int
	var miss int
	var zr *gzip.Reader

	r = NewRadix()

	/* Load file data/ip.db.gz */

	file, err = os.Open("data/ip.db.gz")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	zr, err = gzip.NewReader(file)
	if err != nil {
		panic(err)
	}
	count = 0
	now = time.Now()
	scanner = bufio.NewScanner(zr)
	for scanner.Scan() {

		tokens = strings.Split(scanner.Text(), "|")

		int_dec, err = strconv.ParseInt(tokens[0], 10, 64)
		if err != nil {
			panic(err)
		}
		ip = uint32(int_dec)
		ent.b = make([]byte, 4)
		binary.BigEndian.PutUint32(ent.b, ip)

		int_dec, err = strconv.ParseInt(tokens[1], 10, 8)
		if err != nil {
			panic(err)
		}
		ent.l = int16(int_dec)

		list = append(list, ent)

		count++
	}
	step = time.Now()
	fmt.Printf("Load %d entries in %fs\n", count, step.Sub((now)).Seconds())

	/* populate radix with file loaded */

	now = time.Now()
	for _, ent = range list {
		r.Insert(&ent.b, ent.l, "")
	}
	step = time.Now()
	fmt.Printf("Index %d entries in %fs\n", count, step.Sub((now)).Seconds())

	/* Perform random generation to have a reference */

	bytes = make([]byte, 4)
	now = time.Now()
	for i = 0; i < rounds; i++ {
		binary.BigEndian.PutUint32(bytes, rand.Uint32())
	}
	step = time.Now()
	fmt.Printf("Generate %d random numbers in %fs\n", rounds, step.Sub((now)).Seconds())

	/* perform random lookup to bench algo */

	now = time.Now()
	hit = 0
	miss = 0
	for i = 0; i < rounds; i++ {
		binary.BigEndian.PutUint32(bytes, rand.Uint32())
		node = r.LookupLonguest(&bytes, 32)
		if node != nil {
			hit++
		} else {
			miss++
		}
	}
	step = time.Now()
	d := step.Sub((now)).Seconds()
	fmt.Printf("Generate %d lookup in %fs with %d hit, %d miss\n", rounds, d, hit, miss)
	fmt.Printf("Mean time %f / 1000000 = %fns\n", d, d * 1000.0)

	/* perform full scan */

	now = time.Now()
	node = r.First()
	for {
		node = r.Next(node)
		if node == nil {
			break
		}
		b = []byte(node.node.Bytes)
		for len(b) < 4 {
			b = append([]byte{0x00}, b...)
		}
		ip2.IP = net.IP(b)
		ip2.Mask = net.CIDRMask(int(node.node.End) + 1, 32)
//		fmt.Printf("%s\n", ip2.String())
	}
	step = time.Now()
	fmt.Printf("Dump all data in %fs\n", step.Sub((now)).Seconds())

	/* Return first entry */

	node = r.First()
	if node == nil {
		panic("first cannot be null")
	}
	b = []byte(node.node.Bytes)
	for len(b) < 4 {
		b = append([]byte{0x00}, b...)
	}
	ip2.IP = net.IP(b)
	ip2.Mask = net.CIDRMask(int(node.node.End) + 1, 32)
	fmt.Printf("first = %s\n", ip2.String())

	/* Return last entry */

	node = r.Last()
	if node == nil {
		panic("first cannot be null")
	}
	b = []byte(node.node.Bytes)
	for len(b) < 4 {
		b = append([]byte{0x00}, b...)
	}
	ip2.IP = net.IP(b)
	ip2.Mask = net.CIDRMask(int(node.node.End) + 1, 32)
	fmt.Printf("last = %s\n", ip2.String())

	/* Returrn all cildrens of key 255.255.224.0/20 */

	var it *Iter
	var key []byte
	var ml int16

//	key = []byte{0xff, 0xff, 0xe0, 0x00}
	key = []byte{0xff, 0xff, 0x80, 0x00}
//	key = []byte{0xd9, 0x14, 0x74, 0x88}
	ml = 18

	ip3.IP = net.IP(key)
	ip3.Mask = net.CIDRMask(int(ml), 32)
	it = r.NewIter(&key, ml)
	for it.Next() {
		node = it.Get()
		b = []byte(node.node.Bytes)
		for len(b) < 4 {
			b = append([]byte{0x00}, b...)
		}
		ip2.IP = net.IP(b)
		ip2.Mask = net.CIDRMask(int(node.node.End) + 1, 32)
		fmt.Printf("%s contains %s\n", ip3.String(), ip2.String())
	}

	count = 0
	key = []byte{}
	ml = 0
	now = time.Now()
	it = r.NewIter(&key, ml)
	for it.Next() {
		node = it.Get()
		r.Delete(node)
		count++
	}
	step = time.Now()
	fmt.Printf("Delete %d data in %fs\n", count, step.Sub((now)).Seconds())
}

/* This function browse the tree and validate its integrity */
func browse(t *testing.T, r *Radix) {
	var n *Node
	var s string

	for n = r.First(); n != nil; n = r.Next(n) {
		s = n.Data.(string)
		if s == "" {
			t.Errorf("expect non-empty string")
		}
	}
}

func create_radix_test()(r *Radix) {
	var ipn *net.IPNet

	r = NewRadix()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("192.168.0.0")
	ipn.Mask = net.CIDRMask(16, 32)
	r.IPv4Insert(ipn, "Network 192.168.0.0/16")
	r.check_lvl1_and_die_on_error()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.0.0.0")
	ipn.Mask = net.CIDRMask(8, 32)
	r.IPv4Insert(ipn, "Network 10.0.0.0/8 - 2")
	r.check_lvl1_and_die_on_error()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.0.0.0")
	ipn.Mask = net.CIDRMask(10, 32)
	r.IPv4Insert(ipn, "Network 10.0.0.0/10")
	r.check_lvl1_and_die_on_error()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.64.0.0")
	ipn.Mask = net.CIDRMask(10, 32)
	r.IPv4Insert(ipn, "Network 10.64.0.0/10")
	r.check_lvl1_and_die_on_error()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.64.0.0")
	ipn.Mask = net.CIDRMask(9, 32)
	r.IPv4Insert(ipn, "Network 10.64.0.0/9")
	r.check_lvl1_and_die_on_error()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.64.0.0")
	ipn.Mask = net.CIDRMask(11, 32)
	r.IPv4Insert(ipn, "Network 10.64.0.0/11")
	r.check_lvl1_and_die_on_error()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.96.0.0")
	ipn.Mask = net.CIDRMask(11, 32)
	r.IPv4Insert(ipn, "Network 10.96.0.0/11")
	r.check_lvl1_and_die_on_error()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.0.0.0")
	ipn.Mask = net.CIDRMask(24, 32)
	r.IPv4Insert(ipn, "Network 100.0.0.0/24")
	r.check_lvl1_and_die_on_error()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.0.0.0")
	ipn.Mask = net.CIDRMask(15, 32)
	r.IPv4Insert(ipn, "Network 100.0.0.0/15")
	r.check_lvl1_and_die_on_error()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.7.0.0")
	ipn.Mask = net.CIDRMask(24, 32)
	r.IPv4Insert(ipn, "Network 100.7.0.0/24")
	r.check_lvl1_and_die_on_error()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.7.0.0")
	ipn.Mask = net.CIDRMask(15, 32)
	r.IPv4Insert(ipn, "Network 100.7.0.0/15")
	r.check_lvl1_and_die_on_error()

	return r
}

func Test_Radix(t *testing.T) {
	var r *Radix
	var ipn *net.IPNet
	var n *Node

	/* Check error case */

	r = NewRadix()

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("0.0.0.0")
	ipn.Mask = net.CIDRMask(0, 32)
	r.IPv4Insert(ipn, "Network 0.0.0.0/0")
	r.check_lvl1_and_die_on_error()
	if r.length != 0 {
		t.Errorf("Network 0.0.0.0/0 should not be inserted")
	}
	browse(t, r)

	/* CASE #1
	 */
	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("192.168.0.0")
	ipn.Mask = net.CIDRMask(16, 32)
	r.IPv4Insert(ipn, "Network 192.168.0.0/16")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	/* CASE #2 
	 *
	 * Simple case, just detect existing node, and do nothing.
	 */

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.0.0.0")
	ipn.Mask = net.CIDRMask(8, 32)
	r.IPv4Insert(ipn, "Network 10.0.0.0/8 - 2")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	/* CASE #2 - end
	 *
	 * Add 10.0.0.0/10
	 * Add 10.0.0.0/10
	 * Add 10.64.0.0/10 -> create intermediate node 10.64.0.0/9
	 */

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.0.0.0")
	ipn.Mask = net.CIDRMask(10, 32)
	r.IPv4Insert(ipn, "Network 10.0.0.0/10")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.64.0.0")
	ipn.Mask = net.CIDRMask(10, 32)
	r.IPv4Insert(ipn, "Network 10.64.0.0/10")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.64.0.0")
	ipn.Mask = net.CIDRMask(9, 32)
	r.IPv4Insert(ipn, "Network 10.64.0.0/9")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	/* CASE #3
	 *
	 * - left branch
	 * - right branch
	 */

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.64.0.0")
	ipn.Mask = net.CIDRMask(11, 32)
	r.IPv4Insert(ipn, "Network 10.64.0.0/11")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.96.0.0")
	ipn.Mask = net.CIDRMask(11, 32)
	r.IPv4Insert(ipn, "Network 10.96.0.0/11")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	/* CASE #4
	 *
	 */

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.0.0.0")
	ipn.Mask = net.CIDRMask(24, 32)
	r.IPv4Insert(ipn, "Network 100.0.0.0/24")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.0.0.0")
	ipn.Mask = net.CIDRMask(15, 32)
	r.IPv4Insert(ipn, "Network 100.0.0.0/15")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.7.0.0")
	ipn.Mask = net.CIDRMask(24, 32)
	r.IPv4Insert(ipn, "Network 100.7.0.0/24")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.7.0.0")
	ipn.Mask = net.CIDRMask(15, 32)
	r.IPv4Insert(ipn, "Network 100.7.0.0/15")
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	/* Test delete node */

	r = NewRadix()
	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("192.168.0.0")
	ipn.Mask = net.CIDRMask(16, 32)
	r.IPv4Insert(ipn, "Network 192.168.0.0/16")
	r.check_lvl1_and_die_on_error()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	r = create_radix_test()
	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("192.168.0.0")
	ipn.Mask = net.CIDRMask(16, 32)
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.0.0.0")
	ipn.Mask = net.CIDRMask(8, 32)
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("192.168.0.0")
	ipn.Mask = net.CIDRMask(16, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.0.0.0")
	ipn.Mask = net.CIDRMask(8, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.0.0.0")
	ipn.Mask = net.CIDRMask(10, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.64.0.0")
	ipn.Mask = net.CIDRMask(10, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.64.0.0")
	ipn.Mask = net.CIDRMask(9, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.64.0.0")
	ipn.Mask = net.CIDRMask(11, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("10.96.0.0")
	ipn.Mask = net.CIDRMask(11, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.0.0.0")
	ipn.Mask = net.CIDRMask(24, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.0.0.0")
	ipn.Mask = net.CIDRMask(15, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.7.0.0")
	ipn.Mask = net.CIDRMask(24, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

	ipn = &net.IPNet{}
	ipn.IP = net.ParseIP("100.7.0.0")
	ipn.Mask = net.CIDRMask(15, 32)
	r = create_radix_test()
	n = r.IPv4Get(ipn)
	r.Delete(n)
	r.check_lvl1_and_die_on_error()
	browse(t, r)

}

func Test_Radix_dead_sequence_01(t *testing.T) {
	var r *Radix
	var n *net.IPNet

	r = NewRadix()

	_, n, _ = net.ParseCIDR("34.74.12.152/32")
	fmt.Printf("\nInsert %s\n", n.String())
	r.IPv4Insert(n, true)
	r.check_lvl1_and_die_on_error()
	r.DebugStdout()

	_, n, _ = net.ParseCIDR("34.74.12.153/32")
	fmt.Printf("\nInsert %s\n", n.String())
	r.IPv4Insert(n, true)
	r.check_lvl1_and_die_on_error()
	r.DebugStdout()

	_, n, _ = net.ParseCIDR("34.74.12.152/31")
	fmt.Printf("\nInsert %s\n", n.String())
	r.IPv4Insert(n, true)
	r.check_lvl1_and_die_on_error()
	r.DebugStdout()
}

func Test_Radix_random(t *testing.T) {
	var r *Radix
	var k []byte = make([]byte, 2)
	var n *Node
	var i int

	r = NewRadix()

	for i = 0; i < 10000000; i++ {
		k[0] = byte(rand.Intn(256))
		k[1] = byte(rand.Intn(4)) << 6
		if rand.Intn(1) == 0 {
			r.Insert(&k, int16(rand.Intn(11)), "test")
		} else {
			n = r.LookupLonguest(&k, int16(rand.Intn(11)))
			if n != nil {
				r.Delete(n)
			}
		}
	}
}

func Test_Equal(t *testing.T) {
	var n1 node
	var n2 node
	var k []byte

	/* test equal */
	n1.Bytes = string([]byte{0,0,0,0})
	n1.Start = 0
	n1.End = 31

	n2.Bytes = string([]byte{0,0,0,0})
	n2.Start = 0
	n2.End = 31

	if !equal(&n1, &n2) {
		t.Errorf("Should be equal")
	}

	n2.End = 30
	if equal(&n1, &n2) {
		t.Errorf("Should be different")
	}

	k = []byte(n2.Bytes)
	k[2] = 1
	n2.Bytes = string(k)
	n2.End = 31
	if equal(&n1, &n2) {
		t.Errorf("Should be different")
	}
}

func TestIsAlignedChildrenOf(t *testing.T) {
	var n1 node
	var n2 node

	/* full align / child: extact match */

	n1.Bytes = string([]byte{0,0,0,0})
	n1.Start = 0
	n1.End = 31

	n2.Bytes = string([]byte{0,0,0,0})
	n2.Start = 0
	n2.End = 31

	if !n1.isChildrenOf(&n2) {
		t.Errorf("Should match")
	}

	if !n1.isAlignedChildrenOf(&n2) {
		t.Errorf("Should match")
	}

	/* not aligned / not children: parent is smaller than child */

	n1.Bytes = string([]byte{0,0,0,0})
	n1.Start = 0
	n1.End = 30

	n2.Bytes = string([]byte{0,0,0,0})
	n2.Start = 0
	n2.End = 31

	if n1.isChildrenOf(&n2) {
		t.Errorf("Should not match")
	}

	if n1.isAlignedChildrenOf(&n2) {
		t.Errorf("Should not match")
	}

	/* aligned / children : parent is greatest than child */

	n1.Bytes = string([]byte{0,0,0,0})
	n1.Start = 0
	n1.End = 31

	n2.Bytes = string([]byte{0,0,0,0})
	n2.Start = 0
	n2.End = 23

	if !n1.isChildrenOf(&n2) {
		t.Errorf("Should match")
	}

	if !n1.isAlignedChildrenOf(&n2) {
		t.Errorf("Should match")
	}

	/* not aligned / children */

	n1.Bytes = string([]byte{0,0,0,4})
	n1.Start = 0
	n1.End = 31

	n2.Bytes = string([]byte{0,0,0,0})
	n2.Start = 0
	n2.End = 23

	if !n1.isChildrenOf(&n2) {
		t.Errorf("Should match")
	}

	if n1.isAlignedChildrenOf(&n2) {
		t.Errorf("Should not match")
	}

	/* aligned children - just one bit change */

	/* n2: 3.34.0.0/15 -> 00000011.0010001 0.00000000.00000000
	 * n1: 3.34.0.0/16 -> 00000011.00100010. 00000000.00000000
	 */

	n1.Bytes = string([]byte{3,34,0,0})
	n1.Start = 0
	n1.End = 15

	n2.Bytes = string([]byte{3,34,0,0})
	n2.Start = 0
	n2.End = 14

	if !n1.isChildrenOf(&n2) {
		t.Errorf("Should match")
	}

	if !n1.isAlignedChildrenOf(&n2) {
		t.Errorf("Should not match")
	}
}
