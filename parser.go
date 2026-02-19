package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"image/color"
	"io"
	"math"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

const strategyBoardPrefix = "[stgy:a"
const strategyBoardSuffix = "]"

type StrategyBoardObjectFlag uint8

const (
	Visible StrategyBoardObjectFlag = 1 << iota
	FlipHorizontal
	FlipVertical
	Locked
)

type StrategyBoardObject struct {
	TypeID         int
	Text           string
	Visible        bool
	FlipHorizontal bool
	FlipVertical   bool
	X              int
	Y              int
	Angle          int
	Color          color.NRGBA
	Scale          int
	Params         []int
}

func (o StrategyBoardObject) ScaleFactor(factor float64) (float64, float64) {
	scale := float64(o.Scale) * factor
	flipH := 1.0
	if o.FlipHorizontal {
		flipH = -1.0
	}
	flipV := 1.0
	if o.FlipVertical {
		flipV = -1.0
	}
	return scale * flipH, scale * flipV
}

type StrategyBoard = struct {
	Name       string
	Background int
	Objects    []StrategyBoardObject
}

func DecodeStrategyBoard(input string) ([]byte, error) {
	if !strings.HasPrefix(input, strategyBoardPrefix) || !strings.HasSuffix(input, strategyBoardSuffix) || len(input) < len(strategyBoardPrefix)+len(strategyBoardSuffix)+1 {
		return nil, ParseError
	}

	input = input[len(strategyBoardPrefix) : len(input)-len(strategyBoardSuffix)]

	inputRunes := []rune(input)
	seed := mapIn(forwardTranslateRune(inputRunes[0]))

	var buffer = make([]rune, len(inputRunes)-1)
	for i, c := range inputRunes[1:] {
		t := forwardTranslateRune(c)
		x := mapIn(t)
		y := (x - seed - i) & 0x3f
		forwardTranslate(string(c))
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

func ParseStrategyBoard(data []byte) (StrategyBoard, error) {

	pos := 24
	if readUint16(data, &pos) != 1 {
		return StrategyBoard{}, SectionParseError
	}

	name := readString(data, &pos)

	objects := make([]StrategyBoardObject, 0)

	// parse objects and object text
	for {
		if readUint16(data, &pos) != 2 {
			pos -= 2
			break
		}
		typeId := readUint16(data, &pos)
		text := ""
		if typeId == 100 {
			if readUint16(data, &pos) != 3 {
				return StrategyBoard{}, SectionParseError
			}
			text = readString(data, &pos)
		}
		objects = append(objects, StrategyBoardObject{TypeID: int(typeId), Text: text})
	}

	// parse object flags
	if err := parseSectionHeader(4, data, &pos, objects); err != nil {
		return StrategyBoard{}, err
	}
	for i := range objects {
		flags := StrategyBoardObjectFlag(readUint16(data, &pos))
		objects[i].Visible = Visible&flags != 0
		objects[i].FlipHorizontal = FlipHorizontal&flags != 0
		objects[i].FlipVertical = FlipVertical&flags != 0
	}

	// parse object coordinates
	if err := parseSectionHeader(5, data, &pos, objects); err != nil {
		return StrategyBoard{}, err
	}
	for i := range objects {
		objects[i].X = int(math.Round((float64(readUint16(data, &pos)) / 5120) * 1024))
		objects[i].Y = int(math.Round((float64(readUint16(data, &pos)) / 3840) * 768))
	}

	// parse object angle
	if err := parseSectionHeader(6, data, &pos, objects); err != nil {
		return StrategyBoard{}, err
	}
	for i := range objects {
		objects[i].Angle = readInt16(data, &pos)
	}

	// parse object scale
	if err := parseSectionHeader(7, data, &pos, objects); err != nil {
		return StrategyBoard{}, err
	}
	for i := range objects {
		objects[i].Scale = int(readByte(data, &pos))
	}
	pos += len(objects) % 2

	// parse object color
	if err := parseSectionHeader(8, data, &pos, objects); err != nil {
		return StrategyBoard{}, err
	}
	for i := range objects {
		objects[i].Color = color.NRGBA{
			uint8(readByte(data, &pos)),
			uint8(readByte(data, &pos)),
			uint8(readByte(data, &pos)),
			uint8(math.Round(255.0 * (1.0 - float64(uint8(readByte(data, &pos)))/100.0))),
		}
	}

	// parse object params
	for _, section := range []int{10, 11, 12} {
		if err := parseSectionHeader(section, data, &pos, objects); err != nil {
			return StrategyBoard{}, err
		}
		for i := range objects {
			objects[i].Params = append(objects[i].Params, readInt16(data, &pos))
		}
	}
	return StrategyBoard{Name: name, Objects: objects}, nil

}

func parseSectionHeader(expectedSectionNumber int, data []byte, pos *int, objects []StrategyBoardObject) error {
	if readUint16(data, pos) != expectedSectionNumber {
		return SectionParseError
	}
	*pos += 2
	if readUint16(data, pos) != len(objects) {
		return ObjectCountParseError
	}
	return nil
}
