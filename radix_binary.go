// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package radix

import "bytes"

var beg_mask = [8]byte{ 0xff, 0x7f, 0x3f, 0x1f, 0x0f, 0x07, 0x03, 0x01 }
var end_mask = [8]byte{ 0x80, 0xc0, 0xe0, 0xf0, 0xf8, 0xfc, 0xfe, 0xff }
var mix_mask = [8][8]byte{
	{ 0x80, 0xc0, 0xe0, 0xf0, 0xf8, 0xfc, 0xfe, 0xff },
	{ 0x00, 0x40, 0x60, 0x70, 0x78, 0x7c, 0x7e, 0x7f },
	{ 0x00, 0x00, 0x20, 0x30, 0x38, 0x3c, 0x3e, 0x3f },
	{ 0x00, 0x00, 0x00, 0x10, 0x18, 0x1c, 0x1e, 0x1f },
	{ 0x00, 0x00, 0x00, 0x00, 0x08, 0x0c, 0x0e, 0x0f },
	{ 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x06, 0x07 },
	{ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x03 },
	{ 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01 },
}

func bitcmp(a *[]byte, b *[]byte, start int, end int)(bool) {
	var first_byte int
	var last_byte int

	/* If first byte and last byte are the same
	 * XOR set 0 if bit are equal, 1 in other case.
	 * AND remove useless bit and replace by 0
	 * If the comparison math, the result is 0
	 *
	 * a     ? ? 1 ?  |  ? ? 1 ?
	 * b     ? ? 0 ?  |  ? ? 1 ?
	 * XOR = ? ? 1 ?  |  ? ? 0 ?
	 * mask  0 0 1 0  |  0 0 1 0
	 * AND = 0 0 1 0  |  0 0 0 0
	 *       no match |  match
	 */

	/* we compare "start" and "end" like this
	 *
	 *    start / 8 == end / 8
	 *
	 * so "/ 8" is equivalent to ">> 3" in other words
	 * "/ 8" keep all bit except 3 LSB.
	 * XOR return 0 if the bits are equal, so only the
	 * 3 lsb bit can be set, the result never exceed
	 * 4+2+1 = 7.
	 */
	if (start ^ end) < 8 {
		return ((*a)[start >> 3] ^ (*b)[start >> 3]) & mix_mask[start & 0x07][end & 0x07] == 0
	}

	/* Get position in byte array and get bit position in byte */
	first_byte = start >> 3 /* first_byte = start / 8 */
	last_byte = end >> 3    /* last_byte = end / 8 */
	return !((((*a)[first_byte] ^ (*b)[first_byte]) & beg_mask[start & 0x07] != 0) || // compare first byte
	         (((*a)[last_byte] ^ (*b)[last_byte]) & end_mask[end & 0x07] != 0)  || // compare last byte
	         !bytes.Equal((*a)[first_byte+1:last_byte], (*b)[first_byte+1:last_byte])) // compare remains bytes
}

func bitget(a *[]byte, bitno int)(byte) {
	return ((*a)[bitno / 8] >> (7 - (bitno % 8))) & 0x01
}

/* MSB is bit 0
 * LSB is bit 7
 */
func firstbitset(b byte)(int) {
	var bit int

	bit = 0

	if b & 0xf0 == 0 { /* 1111 0000 => if no match, first bit set have a minimum weight of +4 */
		bit += 4
	} else {
		b >>= 4
	}

	if b & 0x0c == 0 { /* 0000 1100 => if no match, first bit set have a minimum weight of +2 */
		bit += 2
	} else {
		b >>= 2
	}

	if b & 0x02 == 0 { /* 0000 0010 => if no match, first bit set have a minimum weight of +1 */
		bit += 1
	}

	return bit
}

func bitlonguestmatch(a *[]byte, b *[]byte, start int, end int)(int) {
	var first_byte int
	var first_shift int
	var first_byte_mask byte
	var last_byte int
	var last_shift int
	var last_byte_mask byte
	var cmp byte

	if end == -1 {
		return 0
	}

	/* Get position in byte array and get bit position in byte */
	first_byte = start >> 3    /* first_byte = start / 8 */
	first_shift = start & 0x07 /* first_byte = start % 8 */
	last_byte = end >> 3       /* last_byte = end / 8 */
	last_shift = end & 0x07    /* last_byte = end % 8 */

	/* First byte compare mask */
	first_byte_mask = 0xff >> first_shift

	/* Last byte compare mask */
	last_byte_mask = 0xff << (7 - last_shift)

	/* Special case only one byte */
	if first_byte == last_byte {
		cmp = ((*a)[first_byte] ^ (*b)[first_byte]) & first_byte_mask & last_byte_mask
		if cmp == 0 {
			return -1
		}
		return (first_byte * 8) + firstbitset(cmp)
	}

	/* Check difference in the first byte */
	cmp = ((*a)[first_byte] ^ (*b)[first_byte]) & first_byte_mask
	if cmp != 0 {
		return (first_byte * 8) + firstbitset(cmp)
	}

	/* Compare other bytes */
	for first_byte++; first_byte < last_byte; first_byte++ {
		if (*a)[first_byte] != (*b)[first_byte] {
			cmp = (*a)[first_byte] ^ (*b)[first_byte]
			return (first_byte * 8) + firstbitset(cmp)
		}
	}

	/* Check difference in the last byte */
	cmp = ((*a)[last_byte] ^ (*b)[last_byte]) & last_byte_mask
	if cmp != 0 {
		return (last_byte * 8) + firstbitset(cmp)
	}

	/* Prefix are equal */
	return -1
}

/* return true if b is a prefix of a */
func is_children_of(a *[]byte, b *[]byte, al int, bl int)(bool) {
	if bl > al {
		return false
	}
	return bitcmp(a, b, 0, bl)
}
