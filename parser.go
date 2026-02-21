package strategy_board

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"image/color"
	"io"
	"log"
	"math"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

const logVerbose = false
const boardPrefix = "[stgy:a"
const boardSuffix = "]"

var forwardTranslationTable = map[rune]rune{'+': 'N', '-': 'P', '0': 'x', '1': 'g', '2': '0', '3': 'K', '4': '8', '5': 'S', '6': 'J', '7': '2', '8': 's', '9': 'Z', 'A': 'D', 'B': 'F', 'C': 't', 'D': 'T', 'E': '6', 'F': 'E', 'G': 'a', 'H': 'V', 'I': 'c', 'J': 'p', 'K': 'L', 'L': 'M', 'M': 'm', 'N': 'e', 'O': 'j', 'P': '9', 'Q': 'X', 'R': 'B', 'S': '4', 'T': 'R', 'U': 'Y', 'V': '7', 'W': '_', 'X': 'n', 'Y': 'O', 'Z': 'b', 'a': 'i', 'b': '-', 'c': 'v', 'd': 'H', 'e': 'C', 'f': 'A', 'g': 'r', 'h': 'W', 'i': 'o', 'j': 'd', 'k': 'I', 'l': 'q', 'm': 'h', 'n': 'U', 'o': 'l', 'p': 'k', 'q': '3', 'r': 'f', 's': 'y', 't': '5', 'u': 'G', 'v': 'w', 'w': '1', 'x': 'u', 'y': 'z', 'z': 'Q'}

type BoardObjectFlag uint8

const (
	Visible BoardObjectFlag = 1 << iota
	FlipHorizontal
	FlipVertical
	Locked
)

/* Unpack board share code to raw bytes. */
func Unpack(input string) ([]byte, error) {
	log.Println("Unpack strategy board")
	if !strings.HasPrefix(input, boardPrefix) || !strings.HasSuffix(input, boardSuffix) || len(input) < len(boardPrefix)+len(boardSuffix)+1 {
		return nil, ParseError
	}

	input = input[len(boardPrefix) : len(input)-len(boardSuffix)]

	inputRunes := []rune(input)
	seed := mapIn(forwardTranslateRune(inputRunes[0]))

	var buffer = make([]rune, len(inputRunes)-1)
	for i, c := range inputRunes[1:] {
		t := forwardTranslateRune(c)
		x := mapIn(t)
		y := (x - seed - i) & 0x3f
		buffer[i] = mapOut(y)
	}

	base64Str, err := charmap.Windows1252.NewDecoder().String(string(buffer))
	if err != nil {
		return nil, err
	}

	decoded, err := base64.RawURLEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, err
	}

	z, err := zlib.NewReader(bytes.NewReader(decoded[6:]))
	if err != nil {
		return nil, err
	}

	decompressed, err := io.ReadAll(z)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

/* Parse strategy board data */
func Parse(data []byte) (Board, error) {
	log.Printf("Parse %d byte strategy board", len(data))

	// skip first 24 bytes
	pos := 24

	// assert section 1
	if readUint16(data, &pos) != 1 {
		return Board{}, SectionParseError
	}

	// read board name
	name := readString(data, &pos)
	log.Printf("  - NAME: %s", name)

	// read objects and object text
	objects := make([]Object, 0)
	for {
		// object section is 2, if next uint16 isn't 2 then we're done parsing objects
		if readUint16(data, &pos) != 2 {
			pos -= 2
			break
		}
		// read object type id
		typeId := readUint16(data, &pos)
		// read object text, type id 100 is text object
		text := ""
		if typeId == 100 {
			// assert section 3
			if readUint16(data, &pos) != 3 {
				return Board{}, SectionParseError
			}
			text = readString(data, &pos)
		}
		objects = append(objects, Object{TypeID: int(typeId), Text: text})
	}
	log.Printf("  - %d objects read", len(objects))

	// read object flags
	log.Println("  - Parse object flags")
	if err := parseSectionHeader(4, data, &pos, objects); err != nil {
		return Board{}, err
	}
	for i := range objects {
		flags := BoardObjectFlag(readUint16(data, &pos))
		objects[i].Visible = Visible&flags != 0
		objects[i].FlipHorizontal = FlipHorizontal&flags != 0
		objects[i].FlipVertical = FlipVertical&flags != 0
		if logVerbose {
			log.Printf("    - OBJ %d Flags: %d", i+1, flags)
		}
	}

	// read object coordinates
	log.Println("  - Parse object coordinates")
	if err := parseSectionHeader(5, data, &pos, objects); err != nil {
		return Board{}, err
	}
	for i := range objects {
		objects[i].X = int(math.Round((float64(readUint16(data, &pos)) / 5120) * 1024))
		objects[i].Y = int(math.Round((float64(readUint16(data, &pos)) / 3840) * 768))
		if logVerbose {
			log.Printf("    - OBJ %d Coordinates: %d, %d", i+1, objects[i].X, objects[i].Y)
		}
	}

	// read object angle
	log.Println("  - Parse object angles")
	if err := parseSectionHeader(6, data, &pos, objects); err != nil {
		return Board{}, err
	}
	for i := range objects {
		objects[i].Angle = readInt16(data, &pos)
		if logVerbose {
			log.Printf("    - OBJ %d Angle: %d", i+1, objects[i].Angle)
		}
	}

	// read object scale
	log.Println("  - Parse object scales")
	if err := parseSectionHeader(7, data, &pos, objects); err != nil {
		return Board{}, err
	}
	for i := range objects {
		objects[i].Scale = int(readByte(data, &pos))
		if logVerbose {
			log.Printf("    - OBJ %d Scale: %d", i+1, objects[i].Scale)
		}
	}
	pos += len(objects) % 2

	// read object color
	if err := parseSectionHeader(8, data, &pos, objects); err != nil {
		return Board{}, err
	}
	for i := range objects {
		objects[i].Color = color.NRGBA{
			uint8(readByte(data, &pos)),
			uint8(readByte(data, &pos)),
			uint8(readByte(data, &pos)),
			uint8(math.Round(255.0 * (1.0 - float64(uint8(readByte(data, &pos)))/100.0))),
		}
		if logVerbose {
			log.Printf("    - OBJ %d Color: R%d G%d B%d A%d", i+1, objects[i].Color.R, objects[i].Color.G, objects[i].Color.B, objects[i].Color.A)
		}
	}

	// read object params
	for _, section := range []int{10, 11, 12} {
		if err := parseSectionHeader(section, data, &pos, objects); err != nil {
			return Board{}, err
		}
		for i := range objects {
			objects[i].Params = append(objects[i].Params, readInt16(data, &pos))

		}
	}
	if logVerbose {
		for i := range objects {
			log.Printf("    - OBJ %d Params: %d", i+1, objects[i].Params)
		}
	}

	if readUint16(data, &pos) != 3 {
		return Board{}, SectionParseError
	}
	pos += 4
	background := readUint16(data, &pos)

	return Board{Name: name, Background: background, Objects: objects}, nil

}

func Load(input string) (Board, error) {
	log.Println("Load strategy board")
	data, err := Unpack(input)
	if err != nil {
		return Board{}, err
	}
	return Parse(data)
}

func parseSectionHeader(expectedSectionNumber int, data []byte, pos *int, objects []Object) error {
	if readUint16(data, pos) != expectedSectionNumber {
		return SectionParseError
	}
	*pos += 2
	if readUint16(data, pos) != len(objects) {
		return ObjectCountParseError
	}
	return nil
}

func translateRune(c rune, translationTable map[rune]rune) rune {
	if translationTable[c] != 0 {
		return translationTable[c]
	}
	return c
}

func forwardTranslateRune(c rune) rune {
	return translateRune(c, forwardTranslationTable)
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

func readUint16(data []byte, pos *int) int {
	out := binary.LittleEndian.Uint16(data[*pos:])
	*pos += 2
	return int(out)
}

func readInt16(data []byte, pos *int) int {
	return int(int16(readUint16(data, pos)))
}

func readString(data []byte, pos *int) string {
	length := int(binary.LittleEndian.Uint16(data[*pos:]))
	*pos += 2
	out := string(data[*pos : *pos+length])
	*pos += length
	return out
}
