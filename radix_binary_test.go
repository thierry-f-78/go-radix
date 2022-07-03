// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "testing"

func TestBitcmp(t *testing.T) {
	var res int

	/* full bytes */
	if (!bitcmp(&[]byte{0,0,0,0}, &[]byte{0,0,0,0}, 0, 31)) {
		t.Errorf("Should match")
	}
	if (bitcmp(&[]byte{0,0,1,0}, &[]byte{0,0,0,0}, 0, 31)) {
		t.Errorf("Should not match")
	}
	if (!bitcmp(&[]byte{1,0,0,0}, &[]byte{0,0,0,0}, 8, 31)) {
		t.Errorf("Should match")
	}
	if (bitcmp(&[]byte{1,0,1,0}, &[]byte{0,0,0,0}, 8, 31)) {
		t.Errorf("Should not match")
	}
	if (!bitcmp(&[]byte{1,0,0,1}, &[]byte{0,0,0,0}, 8, 23)) {
		t.Errorf("Should match")
	}
	if (bitcmp(&[]byte{1,0,1,1}, &[]byte{0,0,0,0}, 8, 23)) {
		t.Errorf("Should not match")
	}

	/* Partial bytes */
	if (!bitcmp(&[]byte{0xff,0xff,0xff,0xff}, &[]byte{0x0f,0xff,0xff,0xff}, 4, 31)) {
		t.Errorf("Should match")
	}
	if (bitcmp(&[]byte{0xff,0xff,0xff,0xff}, &[]byte{0x0f,0xff,0xf8,0xff}, 4, 31)) {
		t.Errorf("Should not match")
	}
	if (!bitcmp(&[]byte{0xff,0xff,0xff,0xff}, &[]byte{0x0f,0xff,0xff,0xf0}, 4, 27)) {
		t.Errorf("Should match")
	}
	if (bitcmp(&[]byte{0xff,0xff,0xff,0xff}, &[]byte{0x0f,0xff,0xf8,0xf0}, 4, 27)) {
		t.Errorf("Should not match")
	}
	if (!bitcmp(&[]byte{0xff,0xff,0xff,0xff}, &[]byte{0x00,0x0f,0xff,0xf0}, 12, 27)) {
		t.Errorf("Should match")
	}
	if (bitcmp(&[]byte{0xff,0xff,0xff,0xff}, &[]byte{0x00,0x0f,0x8f,0xf0}, 12, 27)) {
		t.Errorf("Should not match")
	}
	if (!bitcmp(&[]byte{0xff,0xff,0xff,0xff}, &[]byte{0x00,0x0f,0xf0,0x00}, 12, 19)) {
		t.Errorf("Should match")
	}
	if (bitcmp(&[]byte{0xff,0xff,0xff,0xff}, &[]byte{0x00,0x0f,0x80,0x00}, 12, 19)) {
		t.Errorf("Should not match")
	}
	if (!bitcmp(&[]byte{0xff,0xff,0xff,0xff}, &[]byte{0x00,0x0e,0x00,0x00}, 12, 14)) {
		t.Errorf("Should match")
	}
	if (bitcmp(&[]byte{0xff,0xff,0xff,0xff}, &[]byte{0x00,0x0c,0x00,0x00}, 12, 14)) {
		t.Errorf("Should not match")
	}

	/* Test bit get */
	if bitget(&[]byte{0x00,0x01,0x00,0x00}, 15) != 1 {
		t.Errorf("Bit 13 should be 1")
	}
	if bitget(&[]byte{0x00,0x80,0x00,0x00}, 8) != 1 {
		t.Errorf("Bit 8 should be 1")
	}
	if bitget(&[]byte{0x00,0x20,0x00,0x00}, 10) != 1 {
		t.Errorf("Bit 10 should be 1")
	}
	if bitget(&[]byte{0xff,0xfe,0xff,0xff}, 15) != 0 {
		t.Errorf("Bit 13 should be 0")
	}
	if bitget(&[]byte{0xff,0x7f,0xff,0xff}, 8) != 0 {
		t.Errorf("Bit 8 should be 0")
	}
	if bitget(&[]byte{0xff,0xdf,0xff,0xff}, 10) != 0 {
		t.Errorf("Bit 10 should be 0")
	}

	/* Test first bit set */
	res = firstbitset(0x80)
	if res != 0 {
		t.Errorf("First bit set should be 0, got %d", res)
	}
	res = firstbitset(0x55)
	if res != 1 {
		t.Errorf("First bit set should be 1, got %d", res)
	}
	res = firstbitset(0x05)
	if res != 5 {
		t.Errorf("First bit set should be 5, got %d", res)
	}
	res = firstbitset(0x01)
	if res != 7 {
		t.Errorf("First bit set should be 7, got %d", res)
	}

	/* test bitlonguestmatch */
	res = bitlonguestmatch(&[]byte{0xff, 0xff, 0xff, 0xff}, &[]byte{0xff, 0xff, 0xff, 0xff}, 0, 31)
	if res != -1 {
		t.Errorf("Expect -1, got %d", res)
	}
	res = bitlonguestmatch(&[]byte{0xff, 0xff, 0xff, 0xff}, &[]byte{0xff, 0xff, 0xfe, 0xff}, 0, 31)
	if res != 23 {
		t.Errorf("Expect 23, got %d", res)
	}
	res = bitlonguestmatch(&[]byte{0xff, 0xff, 0xff, 0xff}, &[]byte{0x00, 0xff, 0xfe, 0xff}, 8, 31)
	if res != 23 {
		t.Errorf("Expect 23, got %d", res)
	}
	res = bitlonguestmatch(&[]byte{0xff, 0xff, 0xff, 0xff}, &[]byte{0xff, 0xff, 0xff, 0xff}, 0, 23)
	if res != -1 {
		t.Errorf("Expect -1, got %d", res)
	}
	res = bitlonguestmatch(&[]byte{0xff, 0xff, 0xff, 0xff}, &[]byte{0xff, 0xff, 0xfd, 0xff}, 0, 22)
	if res != 22 {
		t.Errorf("Expect 22, got %d", res)
	}
	res = bitlonguestmatch(&[]byte{0xff, 0xff, 0xff, 0xff}, &[]byte{0xff, 0xff, 0xfe, 0xff}, 0, 23)
	if res != 23 {
		t.Errorf("Expect 23, got %d", res)
	}
	res = bitlonguestmatch(&[]byte{0xff, 0xff, 0xff, 0xff}, &[]byte{0x00, 0xff, 0xfe, 0xff}, 8, 23)
	if res != 23 {
		t.Errorf("Expect 23, got %d", res)
	}
}
