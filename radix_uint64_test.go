// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "testing"

func TestRadixUInt64(t *testing.T) {
	var nw1 uint64
	var nw2 uint64
	var r *Radix
	var n *Node
	var s string

	/* Init DB */
	r = NewRadix()

	/* Insert value */
	nw1 = 432343254252
	r.UInt64Insert(nw1, "test - nw1")

	/* Lookup network */
	n = r.UInt64Get(nw1)
	if n == nil {
		t.Errorf("Should match")
	}
	nw2 = n.UInt64GetValue()
	if nw2 != nw1 {
		t.Errorf("Should match")
	}
	s = n.Data.(string)
	if s != "test - nw1" {
		t.Errorf("Should match")
	}
}
