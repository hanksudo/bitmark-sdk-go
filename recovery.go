package bitmarksdk

import "fmt"

// 0...10 bit masks
var masks = []int{0, 1, 3, 7, 15, 31, 63, 127, 255, 511, 1023}

// convert a binary of 33 bytes to a phrase of 24 worhs
func bytesToPhrase(input [33]byte) []string {
	phrase := make([]string, 0, 24)
	accumulator := 0
	bits := 0
	n := 0
	for i := 0; i < len(input); i += 1 {
		accumulator = accumulator<<8 + int(input[i])
		bits += 8
		if bits >= 11 {
			bits -= 11 // [ 11 bits] [offset bits]

			n += 1
			index := accumulator >> uint(bits)
			accumulator &= masks[bits]
			word := wordlist[index]
			phrase = append(phrase, word)
		}
	}

	return phrase
}

// array of words to 33 bytes
func phraseToBytes(words []string) ([33]byte, error) {
	var seed [33]byte

	if 24 != len(words) {
		return seed, fmt.Errorf("number of words: %d expected: 24", len(words))
	}

	databytes := make([]byte, 0, 33)

	remainder := 0
	bits := 0
	for _, word := range words {
		n := -1
	loop:
		for i, bip := range wordlist {
			if word == bip {
				n = i
				break loop
			}
		}
		if n < 0 {
			return seed, fmt.Errorf("invalid word: %q", word)
		}
		remainder = remainder<<11 + n
		for bits += 11; bits >= 8; bits -= 8 {
			a := 0xff & (remainder >> uint(bits-8))
			databytes = append(databytes, byte(a))

		}
		remainder &= masks[bits]

	}
	if 33 != len(databytes) {
		return seed, fmt.Errorf("only converted: %d bytes expected: 33", len(databytes))
	}

	copy(seed[:], databytes)
	return seed, nil
}
