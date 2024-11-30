package hl7

import (
	"errors"
	"reflect"
)

var (
	ErrSegmentInvalid     = errors.New("hl7: invalid segment")
	ErrSegmentTypeInvalid = errors.New("hl7: invalid segment type, expected a struct")
	ErrTagInvalidFormat   = errors.New("hl7: tag is not in the correct format, expected `hl7:\"segment:<name>\"`")
)

var errTagEmtpy = errors.New("hl7: tag is empty")

// InvalidMessageParserError describes an invalid argument passed to the parser.
type InvalidMessageParserError struct {
	Type reflect.Type
}

func (e InvalidMessageParserError) Error() string {
	if e.Type == nil {
		return "hl7: Unmarshal(nil)"
	}
	if e.Type.Kind() != reflect.Pointer {
		return "hl7: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "hl7: Unmarshal(nil " + e.Type.String() + ")"
}
