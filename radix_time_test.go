// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "testing"
import "time"

func TestRadixTime(t *testing.T) {
	var nw1 time.Time
	var nw2 time.Time
	var r *Radix
	var n *Node
	var s string

	/* Init DB */
	r = NewRadix()

	/* Insert value */
	nw1 = time.Now()
	r.TimeInsert(nw1, "test - nw1")

	/* Lookup network */
	n = r.TimeGet(nw1)
	if n == nil {
		t.Errorf("Should match")
	} else {
		nw2 = n.TimeGetValue()
		if !nw2.Equal(nw1) {
			t.Errorf("Should match")
		}
		s = n.Data.(string)
		if s != "test - nw1" {
			t.Errorf("Should match")
		}
	}
}
