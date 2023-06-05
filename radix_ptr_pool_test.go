// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "reflect"
import "testing"

func Test_ptr_range(t *testing.T) {
	var r *Radix
	var want []ptr_range

	r = &Radix{}

	r.add_range(12, 13, 1)
	want = []ptr_range{
		ptr_range{12, 13, 1},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(14, 15, 2)
	want = []ptr_range{
		ptr_range{12, 13, 1},
		ptr_range{14, 15, 2},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(1, 2, 3)
	want = []ptr_range{
		ptr_range{1, 2, 3},
		ptr_range{12, 13, 1},
		ptr_range{14, 15, 2},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(6, 9, 4)
	want = []ptr_range{
		ptr_range{1, 2, 3},
		ptr_range{6, 9, 4},
		ptr_range{12, 13, 1},
		ptr_range{14, 15, 2},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(7, 8, 5)
	want = []ptr_range{
		ptr_range{1, 2, 3},
		ptr_range{6, 9, 4},
		ptr_range{12, 13, 1},
		ptr_range{14, 15, 2},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(6, 7, 5)
	want = []ptr_range{
		ptr_range{1, 2, 3},
		ptr_range{6, 9, 4},
		ptr_range{12, 13, 1},
		ptr_range{14, 15, 2},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(7, 8, 5)
	want = []ptr_range{
		ptr_range{1, 2, 3},
		ptr_range{6, 9, 4},
		ptr_range{12, 13, 1},
		ptr_range{14, 15, 2},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(6, 9, 5)
	want = []ptr_range{
		ptr_range{1, 2, 3},
		ptr_range{6, 9, 4},
		ptr_range{12, 13, 1},
		ptr_range{14, 15, 2},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(5, 8, 5)
	want = []ptr_range{
		ptr_range{1, 2, 3},
		ptr_range{6, 9, 4},
		ptr_range{12, 13, 1},
		ptr_range{14, 15, 2},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(8, 10, 5)
	want = []ptr_range{
		ptr_range{1, 2, 3},
		ptr_range{6, 9, 4},
		ptr_range{12, 13, 1},
		ptr_range{14, 15, 2},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}

	r.add_range(5, 10, 5)
	want = []ptr_range{
		ptr_range{1, 2, 3},
		ptr_range{6, 9, 4},
		ptr_range{12, 13, 1},
		ptr_range{14, 15, 2},
	}
	if !reflect.DeepEqual(r.ptr_range, want) {
		t.Errorf("Unmatched:\n   got: %#v\nexpect: %#v\n", r.ptr_range, want)
	}
}
