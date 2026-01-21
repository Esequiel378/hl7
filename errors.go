package hl7

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrSegmentInvalid     = errors.New("hl7: invalid segment")
	ErrSegmentTypeInvalid = errors.New("hl7: invalid segment type, expected a struct")
	ErrTagInvalidFormat   = errors.New("hl7: tag is not in the correct format, expected `hl7:\"segment:<name>\"`")
)

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

// FieldError represents an error that occurred while processing a specific HL7 field.
// It provides context about which segment and field caused the error, making debugging
// in production healthcare systems significantly easier.
type FieldError struct {
	Segment   string // The segment name (e.g., "PID", "MSH")
	Field     int    // The 1-based field index
	Component int    // The 1-based component index (0 if not applicable)
	Value     string // The raw value that caused the error
	Err       error  // The underlying error
}

func (e *FieldError) Error() string {
	if e.Component > 0 {
		return fmt.Sprintf("hl7: %s.%d.%d: %v (value=%q)",
			e.Segment, e.Field, e.Component, e.Err, e.Value)
	}
	return fmt.Sprintf("hl7: %s.%d: %v (value=%q)",
		e.Segment, e.Field, e.Err, e.Value)
}

func (e *FieldError) Unwrap() error {
	return e.Err
}
