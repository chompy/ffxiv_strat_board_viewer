package main

import "errors"

var (
	ParseError                      = errors.New("decode error: invalid strategy board")
	SectionParseError               = errors.New("parse error: read unexpected section number")
	ObjectCountParseError           = errors.New("parse error: unexpected number of objects in section")
	DrawUnexpectedObjectError       = errors.New("draw error: unexpected object type")
	DrawInvalidObjectParameterError = errors.New("draw error: invalid object parameter")
	AssetNotFound                   = errors.New("asset not found")
)
