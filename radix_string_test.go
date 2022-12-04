// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "fmt"
import "strings"
import "testing"

func display_node_string(n *Node, level int, branch string) {
	var typ string

	if n.Data != nil {
		typ = "LEAF"
	} else {
		typ = "NODE"
	}

	fmt.Printf("%s%s: %p/%s start=%d end=%d key=%q/%d\n", strings.Repeat("   ", level), branch, n, typ, n.Start, n.End, string(n.Bytes), n.End + 1)
	if n.Left != nil {
		display_node_string(n.Left, level+1, "L")
	}
	if n.Right != nil {
		display_node_string(n.Right, level+1, "R")
	}

}

func display_radix_string(r *Radix) {

	if r.Node == nil {
		fmt.Printf("root pointer nil\n")
		return
	}

	display_node_string(r.Node, 0, "-")
}

func Test_string(t *testing.T) {
	var r *Radix
	var n *Node
	var ns []*Node
	var s string

	/* Init DB */
	r = NewRadix()

	/* Insert string */
	r.StringInsert("aaaa", "key aaaa")
	r.StringInsert("aaa", "key aaa")
	r.StringInsert("aa", "key aa")

	/* Lookup exact */
	n = r.StringGet("aaaa")
	if n == nil {
		t.Errorf("aaaa should be found")
	}
	if n.StringGetKey() != "aaaa" {
		t.Errorf("aaaa should be found")
	}
	s, _ = n.Data.(string)
	if s != "key aaaa" {
		t.Errorf("\"key aaaa\" should be found")
	}

	/* Lookup exact */
	n = r.StringGet("aa")
	if n == nil {
		t.Errorf("aa should be found")
	}
	if n.StringGetKey() != "aa" {
		t.Errorf("aa should be found")
	}
	s, _ = n.Data.(string)
	if s != "key aa" {
		t.Errorf("\"key aa\" should be found")
	}

	/* lookup longest prefix */
	n = r.StringLookupLonguest("aaaa stayin alive")
	if n == nil {
		t.Errorf("aaaa should be found")
	}
	if n.StringGetKey() != "aaaa" {
		t.Errorf("aaaa should be found")
	}
	s, _ = n.Data.(string)
	if s != "key aaaa" {
		t.Errorf("\"key aaaa\" should be found")
	}

	/* lookup longest prefix */
	n = r.StringLookupLonguest("aa stayin alive")
	if n == nil {
		t.Errorf("aa should be found")
	}
	if n.StringGetKey() != "aa" {
		t.Errorf("aa should be found")
	}
	s, _ = n.Data.(string)
	if s != "key aa" {
		t.Errorf("\"key aa\" should be found")
	}

	/* Lookup longest path */
	ns = r.StringLookupLonguestPath("aaaa")
	if len(ns) != 3 {
		t.Errorf("Expect 3 entries")
	}
}
