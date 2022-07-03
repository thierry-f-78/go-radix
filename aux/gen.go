// Copyright (C) 2022 Thierry Fournier <tfournier@arpalert.org>

package main

import "fmt"

func main() {
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

