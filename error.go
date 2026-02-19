package main

import "errors"

var (
	MissingInput              = errors.New("missing strategy board input data")
	ParseError                = errors.New("parse error: invalid strategy board")
	SectionParseError         = errors.New("parse error: read unexpected section number")
	ObjectCountParseError     = errors.New("parse error: unexpected number of objects in section")
	DrawUnexpectedObjectError = errors.New("draw error: unexpected object type")
	AssetNotFound             = errors.New("asset not found")
)
