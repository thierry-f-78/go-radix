// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "fmt"
import "testing"

func TestBitcmp(t *testing.T) {
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
}

func TestBitget(t *testing.T) {
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
}

func TestFirstbitset(t *testing.T) {
	var res int16

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
}

func TestBitlonguestmatch(t *testing.T) {
	var res int16

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

func TestIs_children_of(t *testing.T) {
	/* test is_children_of */
	if !is_children_of(&[]byte{0xff, 0xff, 0x00, 0x00}, &[]byte{0xff, 0xff, 0x00, 0x00}, 15, 15) {
		t.Errorf("Should be true")
	}
	if is_children_of(&[]byte{0xff, 0x00, 0x00, 0x00}, &[]byte{0xff, 0xff, 0x00, 0x00}, 15, 15) {
		t.Errorf("Should be false")
	}
	if !is_children_of(&[]byte{0xff, 0xff, 0xff, 0x00}, &[]byte{0xff, 0xff, 0x00, 0x00}, 23, 15) {
		t.Errorf("Should be true")
	}
	if is_children_of(&[]byte{0xff, 0x00, 0xff, 0x00}, &[]byte{0xff, 0xff, 0x00, 0x00}, 23, 15) {
		t.Errorf("Should be false")
	}
}

func TestArezero(t *testing.T) {
	/* test are_zero */
	if (!are_zero(&[]byte{0,0,0,0}, 0, 31)) {
		t.Errorf("Should match")
	}
	if (are_zero(&[]byte{0,0,1,0}, 0, 31)) {
		t.Errorf("Should not match")
	}
	if (!are_zero(&[]byte{1,0,0,0}, 8, 31)) {
		t.Errorf("Should match")
	}
	if (are_zero(&[]byte{1,0,1,0}, 8, 31)) {
		t.Errorf("Should not match")
	}
	if (!are_zero(&[]byte{1,0,0,1}, 8, 23)) {
		t.Errorf("Should match")
	}
	if (are_zero(&[]byte{1,0,1,1}, 8, 23)) {
		t.Errorf("Should not match")
	}

	/* Partial bytes */
	if (!are_zero(&[]byte{0xf0,0x00,0x00,0x00}, 4, 31)) {
		t.Errorf("Should match")
	}
	if (are_zero(&[]byte{0xff,0x00,0x00,0x00}, 4, 31)) {
		t.Errorf("Should not match")
	}

	/* One bit at byte start */
	if (!are_zero(&[]byte{0xff,0x7f,0xff,0xff}, 8, 8)) {
		t.Errorf("Should match")
	}

	/* One bit at byte start */
	if (!are_zero(&[]byte{0x7f,0xff,0xff,0xff}, 0, 0)) {
		t.Errorf("Should match")
	}

	/* One bit at byte end */
	if (!are_zero(&[]byte{0xff,0b11111110,0xff,0xff}, 15, 15)) {
		t.Errorf("Should match")
	}

	/* One bit at byte anywhere */
	if (!are_zero(&[]byte{0xff,0b11110111,0xff,0xff}, 12, 12)) {
		t.Errorf("Should match")
	}

	/* One bit at byte end */
	if (!are_zero(&[]byte{0xff,0xff,0xff,0b11111110}, 31, 31)) {
		t.Errorf("Should match")
	}
}

func TestGenMasks(t *testing.T) {
	var i int
	var j int
	var beg_mask [8]byte
	var end_mask [8]byte
	var mix_mask [8][8]byte

	/* Build mask value for byte keep at end of byte */
	for i = 0; i < 8; i++ {
		beg_mask[i] = 0xff >> i
	}

	/* Build values for byte keepv at start of byte */
	for i = 0; i < 8; i++ {
		end_mask[i] = 0xff << (7 - i)
	}

	/* Build values for mixed mask */
	for  i = 0; i < 8; i++ {
		for j = 0; j < 8; j++ {
			mix_mask[i][j] = beg_mask[i] & end_mask[j]
		}
	}

	/* Display results */
	fmt.Printf("var beg_mask = [8]byte{ 0x%02x", beg_mask[0])
	for i = 1; i < 8; i++ {
		fmt.Printf(", 0x%02x", beg_mask[i])
	}
	fmt.Printf(" }\n")

	fmt.Printf("var end_mask = [8]byte{ 0x%02x", end_mask[0])
	for i = 1; i < 8; i++ {
		fmt.Printf(", 0x%02x", end_mask[i])
	}
	fmt.Printf(" }\n")

	fmt.Printf("var mix_mask = [8][8]byte{\n");
	for i = 0; i < 8; i++ {
		fmt.Printf("\t{ 0x%02x", mix_mask[i][0])
		for j = 1; j < 8; j++ {
			fmt.Printf(", 0x%02x", mix_mask[i][j])
		}
		fmt.Printf(" },\n")
	}
	fmt.Printf(" }\n")
}
