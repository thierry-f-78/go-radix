// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

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

func display_node(n *Node, level int, branch string) {
	var typ string
	var ip net.IPNet
	var b []byte

	if n.Data != nil {
		typ = "LEAF"
	} else {
		typ = "NODE"
	}

	b = n.Bytes
	for len(b) < 4 {
		b = append([]byte{0x00}, b...)
	}
	ip.IP = net.IP(b)
	ip.Mask = net.CIDRMask(n.End + 1, 32)

	fmt.Printf("%s%s: %p/%s start=%d end=%d ip=%s\n", strings.Repeat("   ", level), branch, n, typ, n.Start, n.End, ip.String())
	if n.Left != nil {
		display_node(n.Left, level+1, "L")
	}
	if n.Right != nil {
		display_node(n.Right, level+1, "R")
	}

}

func display_radix(r *Radix) {

	if r.Node == nil {
		fmt.Printf("root pointer nil\n")
		return
	}

	display_node(r.Node, 0, "-")
}

func TestRadix(t *testing.T) {
	/*
	var r *Radix
	var b []byte
	var n []*Node

	r = NewRadix(true)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00010101}
	insert(r, b, 32, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00010011}
	insert(r, b, 32, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00010000}
	insert(r, b, 29, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00000101}
	insert(r, b, 32, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00000101}
	insert(r, b, 32, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00001100}
	insert(r, b, 30, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00001100}
	insert(r, b, 32, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00001000}
	insert(r, b, 29, nil)
	display_radix(r)

	println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	b = []byte{0x00, 0x00, 0x00, 0b00010101}
	n = lookup_longuest(r, b, 32, true)
	fmt.Printf("%+v\n", n)
	*/
}

type ent struct {
	b []byte
	l int
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

	r = NewRadix(true)

	/* Load file data/ip.db */

	file, err = os.Open("data/ip.db")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	count = 0
	now = time.Now()
	scanner = bufio.NewScanner(file)
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
		ent.l = int(int_dec)

		list = append(list, ent)

		count++
	}
	step = time.Now()
	fmt.Printf("Load %d entries in %fs\n", count, step.Sub((now)).Seconds())

	/* populate radix with file loaded */

	now = time.Now()
	for _, ent = range list {
		insert(r, &ent.b, ent.l, nil)
	}
	step = time.Now()
	fmt.Printf("Index %d entries in %fs\n", count, step.Sub((now)).Seconds())

	/* Perform random generation to have a reference */

	bytes = make([]byte, 4)
	now = time.Now()
	for i = 0; i < rounds; i++ {
		binary.LittleEndian.PutUint32(bytes, rand.Uint32())
	}
	step = time.Now()
	fmt.Printf("Generate %d random numbers in %fs\n", rounds, step.Sub((now)).Seconds())

	/* perform random lookup to bench algo */

	now = time.Now()
	hit = 0
	miss = 0
	for i = 0; i < rounds; i++ {
		binary.BigEndian.PutUint32(bytes, rand.Uint32())
		node = lookup_longuest_last_match(r, &bytes, 32)
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
	node = r.Node
	for {
		node = next(node)
		if node == nil {
			break
		}
		b = node.Bytes
		for len(b) < 4 {
			b = append([]byte{0x00}, b...)
		}
		ip2.IP = net.IP(b)
		ip2.Mask = net.CIDRMask(node.End + 1, 32)
//		fmt.Printf("%s\n", ip2.String())
	}
	step = time.Now()
	fmt.Printf("Dump all data in %fs\n", step.Sub((now)).Seconds())

	/* Perform revser scan */

	now = time.Now()
	node = r.Node
	for {
		node = prev(node)
		if node == nil {
			break
		}
		b = node.Bytes
		for len(b) < 4 {
			b = append([]byte{0x00}, b...)
		}
		ip2.IP = net.IP(b)
		ip2.Mask = net.CIDRMask(node.End + 1, 32)
//		fmt.Printf("%s\n", ip2.String())
	}
	step = time.Now()
	fmt.Printf("Dump all data in %fs\n", step.Sub((now)).Seconds())

	/* Return first entry */

	node = first(r)
	if node == nil {
		panic("first cannot be null")
	}
	b = node.Bytes
	for len(b) < 4 {
		b = append([]byte{0x00}, b...)
	}
	ip2.IP = net.IP(b)
	ip2.Mask = net.CIDRMask(node.End + 1, 32)
	fmt.Printf("first = %s\n", ip2.String())

	/* Return last entry */

	node = last(r)
	if node == nil {
		panic("first cannot be null")
	}
	b = node.Bytes
	for len(b) < 4 {
		b = append([]byte{0x00}, b...)
	}
	ip2.IP = net.IP(b)
	ip2.Mask = net.CIDRMask(node.End + 1, 32)
	fmt.Printf("last = %s\n", ip2.String())

	/* Returrn all cildrens of key 255.255.224.0/20 */

	var it *iter
	var key []byte
	var ml int

//	key = []byte{0xff, 0xff, 0xe0, 0x00}
	key = []byte{0xff, 0xff, 0x80, 0x00}
//	key = []byte{0xd9, 0x14, 0x74, 0x88}
	ml = 18

	ip3.IP = net.IP(key)
	ip3.Mask = net.CIDRMask(ml, 32)
	it = new_iter(r, &key, ml, true)
	for it.next() {
		node = it.get()
		b = node.Bytes
		for len(b) < 4 {
			b = append([]byte{0x00}, b...)
		}
		ip2.IP = net.IP(b)
		ip2.Mask = net.CIDRMask(node.End + 1, 32)
		fmt.Printf("%s contains %s\n", ip3.String(), ip2.String())
	}

	count = 0
	key = []byte{}
	ml = 0
	now = time.Now()
	it = new_iter(r, &key, ml, true)
	for it.next() {
		node = it.get()
		del(r, node)
		count++
	}
	step = time.Now()
	fmt.Printf("Delete %d data in %fs\n", count, step.Sub((now)).Seconds())
}
