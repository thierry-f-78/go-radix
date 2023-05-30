// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "testing"


func Test_ref(t *testing.T) {
	var r *Radix
	var ref uint32
	var ref_back uint32
	var nref *Node
	var nref_back *Node

	r = NewRadix()
	r.growth()
	r.growth()
	r.growth()
	r.growth()
	r.growth()

	if r.capacity != 5 * 65536 {
		t.Fatalf("Expect capacity of %d, got %d", 5 * 65536, r.capacity)
	}

	ref = uint32(3 << 16 | 4343)
	ref_back = r.n2r(r.r2n(ref))
	if ref != ref_back {
		t.Fatalf("Expect reference %x, got %x", ref, ref_back)
	}

	nref = &r.pool[3].nodes[16332]
	nref_back = r.r2n(r.n2r(nref))
	if nref != nref_back {
		t.Fatalf("Expect reference %p, got %p", nref, nref_back)
	}
}
