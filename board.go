package main

import (
	"image/color"
)

type BoardObject struct {
	TypeID         int         `json:"type_id"`
	Text           string      `json:"text"`
	Visible        bool        `json:"visible"`
	FlipHorizontal bool        `json:"flip_horizontal"`
	FlipVertical   bool        `json:"flip_vertical"`
	X              int         `json:"x"`
	Y              int         `json:"y"`
	Angle          int         `json:"angle"`
	Color          color.NRGBA `json:"color"`
	Scale          int         `json:"scale"`
	Params         []int       `json:"params"`
}

func (o BoardObject) ScaleFactor(factor float64) (float64, float64) {
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

type Board struct {
	Name       string        `json:"name"`
	Background int           `json:"background"`
	Objects    []BoardObject `json:"object"`
}

func (b Board) Assets() ([]Asset, error) {
	return LoadBoardAssets(b)
}
