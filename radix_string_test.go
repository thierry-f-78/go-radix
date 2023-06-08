// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "testing"

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
	} else {
		if n.StringGetKey() != "aaaa" {
			t.Errorf("aaaa should be found")
		}
		s, _ = n.Data.(string)
		if s != "key aaaa" {
			t.Errorf("\"key aaaa\" should be found")
		}
	}

	/* Lookup exact */
	n = r.StringGet("aa")
	if n == nil {
		t.Errorf("aa should be found")
	} else {
		if n.StringGetKey() != "aa" {
			t.Errorf("aa should be found")
		}
		s, _ = n.Data.(string)
		if s != "key aa" {
			t.Errorf("\"key aa\" should be found")
		}
	}

	/* lookup longest prefix */
	n = r.StringLookupLonguest("aaaa stayin alive")
	if n == nil {
		t.Errorf("aaaa should be found")
	} else {
		if n.StringGetKey() != "aaaa" {
			t.Errorf("aaaa should be found")
		}
		s, _ = n.Data.(string)
		if s != "key aaaa" {
			t.Errorf("\"key aaaa\" should be found")
		}
	}

	/* lookup longest prefix */
	n = r.StringLookupLonguest("aa stayin alive")
	if n == nil {
		t.Errorf("aa should be found")
	} else {
		if n.StringGetKey() != "aa" {
			t.Errorf("aa should be found")
		}
		s, _ = n.Data.(string)
		if s != "key aa" {
			t.Errorf("\"key aa\" should be found")
		}
	}

	/* lookup longest prefix */
	n = r.StringLookupLonguest("ar stayin alive")
	if n != nil {
		t.Errorf("lookup should be nil")
	}

	/* Lookup longest path */
	ns = r.StringLookupLonguestPath("aaaa")
	if len(ns) != 3 {
		t.Errorf("Expect 3 entries")
	}
}
