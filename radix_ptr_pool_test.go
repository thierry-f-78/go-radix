// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "reflect"
import "testing"

func Test_ptr_range(t *testing.T) {
	var r *Radix
	var want []ptr_range

	r = &Radix{}

	r.add_range(12, 13, 1, 0)
	want = []ptr_range{
		ptr_range{12, 13, 1, 0},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(14, 15, 2, 0)
	want = []ptr_range{
		ptr_range{12, 13, 1, 0},
		ptr_range{14, 15, 2, 0},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(1, 2, 3, 0)
	want = []ptr_range{
		ptr_range{1, 2, 3, 0},
		ptr_range{12, 13, 1, 0},
		ptr_range{14, 15, 2, 0},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(6, 9, 4, 0)
	want = []ptr_range{
		ptr_range{1, 2, 3, 0},
		ptr_range{6, 9, 4, 0},
		ptr_range{12, 13, 1, 0},
		ptr_range{14, 15, 2, 0},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	/*
	 * the following error test generate panic
	 */

	func() {
		defer func() { if recover() == nil { t.Errorf("Expect panic") } }()
		r.add_range(7, 8, 5, 0)
	}()

	func() {
		defer func() { if recover() == nil { t.Errorf("Expect panic") } }()
		r.add_range(6, 7, 5, 0)
	} ()

	func() {
		defer func() { if recover() == nil { t.Errorf("Expect panic") } }()
		r.add_range(7, 8, 5, 0)
	} ()

	func() {
		defer func() { if recover() == nil { t.Errorf("Expect panic") } }()
		r.add_range(6, 9, 5, 0)
	} ()

	func() {
		defer func() { if recover() == nil { t.Errorf("Expect panic") } }()
		r.add_range(5, 8, 5, 0)
	} ()

	func() {
		defer func() { if recover() == nil { t.Errorf("Expect panic") } }()
		r.add_range(8, 10, 5, 0)
	} ()

	func() {
		defer func() { if recover() == nil { t.Errorf("Expect panic") } }()
		r.add_range(5, 10, 5, 0)
	} ()
}
