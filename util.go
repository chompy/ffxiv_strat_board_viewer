package main

import "encoding/binary"

var FORWARD_TRANSLATION_TABLE = map[rune]rune{'+': 'N', '-': 'P', '0': 'x', '1': 'g', '2': '0', '3': 'K', '4': '8', '5': 'S', '6': 'J', '7': '2', '8': 's', '9': 'Z', 'A': 'D', 'B': 'F', 'C': 't', 'D': 'T', 'E': '6', 'F': 'E', 'G': 'a', 'H': 'V', 'I': 'c', 'J': 'p', 'K': 'L', 'L': 'M', 'M': 'm', 'N': 'e', 'O': 'j', 'P': '9', 'Q': 'X', 'R': 'B', 'S': '4', 'T': 'R', 'U': 'Y', 'V': '7', 'W': '_', 'X': 'n', 'Y': 'O', 'Z': 'b', 'a': 'i', 'b': '-', 'c': 'v', 'd': 'H', 'e': 'C', 'f': 'A', 'g': 'r', 'h': 'W', 'i': 'o', 'j': 'd', 'k': 'I', 'l': 'q', 'm': 'h', 'n': 'U', 'o': 'l', 'p': 'k', 'q': '3', 'r': 'f', 's': 'y', 't': '5', 'u': 'G', 'v': 'w', 'w': '1', 'x': 'u', 'y': 'z', 'z': 'Q'}

func translateRune(c rune, translationTable map[rune]rune) rune {
	if translationTable[c] != 0 {
		return translationTable[c]
	}
	return c
}

func forwardTranslateRune(c rune) rune {
	return translateRune(c, FORWARD_TRANSLATION_TABLE)
}

func translateString(input string, translationTable map[rune]rune) string {
	var output = ""
	for _, c := range input {
		if translationTable[c] != 0 {
			output += string(translationTable[c])
		} else {
			output += string(c)
		}
	}
	return output
}

func forwardTranslate(input string) string {
	return translateString(input, FORWARD_TRANSLATION_TABLE)
}

func mapIn(c rune) int {
	if c >= 'A' && c <= 'Z' {
		return int(c) - 65
	}
	if c >= 'a' && c <= 'z' {
		return int(c) - 71
	}
	if c >= '0' && c <= '9' {
		return int(c) + 4
	}
	if c == '-' || c == '>' {
		return 62
	}
	if c == '_' || c == '?' {
		return 63
	}
	return 0
}

func mapOut(n int) rune {
	if n < 26 {
		return rune(n + 65)
	}
	if n < 52 {
		return rune(n + 71)
	}
	if n < 62 {
		return rune(n - 4)
	}
	if n == 62 {
		return '-'
	}
	return '_'
}

func readByte(data []byte, pos *int) byte {
	out := data[*pos]
	*pos += 1
	return out
}

func readInt16(data []byte, pos *int) int {
	out := int16(binary.LittleEndian.Uint16(data[*pos:]))
	*pos += 2
	return int(out)
}

func readUint16(data []byte, pos *int) int {
	out := binary.LittleEndian.Uint16(data[*pos:])
	*pos += 2
	return int(out)
}

func readUint32(data []byte, pos *int) int {
	out := binary.LittleEndian.Uint32(data[*pos:])
	*pos += 2
	return int(out)
}

func readString(data []byte, pos *int) string {
	length := int(binary.LittleEndian.Uint16(data[*pos:]))
	*pos += 2
	out := string(data[*pos : *pos+length])
	*pos += length
	return out
}
