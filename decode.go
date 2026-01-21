package hl7

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshaler is the interface implemented by types
// that can unmarshal themselves.
// Unmarshal must copy the input data if it wishes
// to retain the data after returning.
type Unmarshaler interface {
	Unmarshal([]byte) error
}

var unmarshalerType = reflect.TypeFor[Unmarshaler]()

// implementsUnmarshaler checks if a field implements the Unmarshaler interface
func implementsUnmarshaler(val reflect.Value) bool {
	// If the value is invalid (e.g., a nil value), return false
	if !val.IsValid() {
		return false
	}

	// Check if the value itself implements Unmarshaler
	if val.Type().Implements(unmarshalerType) {
		return true
	}

	// If the value is addressable, check if the pointer to it implements Unmarshaler
	if val.CanAddr() {
		return val.Addr().Type().Implements(unmarshalerType)
	}

	// Otherwise, it does not implement the interface
	return false
}

// Unmarshal parses the HL7 data into the provided struct v.
// v must be a pointer to a struct, and its fields should be tagged with `hl7:"segment:<name>"` for segments,
// and `hl7:"<index>"` for fields within the segment.
func Unmarshal(data []byte, v any) error {
	// Validate that v is a pointer to a struct
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return InvalidMessageParserError{reflect.TypeOf(v)}
	}

	// Map to store the tag to struct field mapping
	tagToField := make(map[Segment]reflect.Value)

	// Iterate over struct fields to map segment names to fields
	for i := 0; i < rv.Elem().NumField(); i++ {
		field := rv.Elem().Field(i)
		tag, err := getHL7SegmentTypeFromTag(rv.Elem().Type().Field(i).Tag.Get("hl7"))
		if errors.Is(err, errTagEmpty) {
			continue
		}
		if err != nil {
			return err
		}
		// Ensure the segment field is a struct
		if field.Kind() != reflect.Struct {
			return fmt.Errorf("%w: %s", ErrSegmentTypeInvalid, reflect.TypeOf(field).Name())
		}
		tagToField[tag] = field
	}

	fieldSeparator := "|"
	encodingCharacters := "^~\\&"

	// Scan the HL7 message line by line
	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Increase the scanner buffer to handle long HL7 segments
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()

		// Extract the field separator from the MSH segment if present
		if strings.HasPrefix(line, "MSH") {
			fieldSeparator = string(line[3])
		}

		parts := strings.Split(line, fieldSeparator)
		if len(parts) == 0 {
			continue
		}

		if strings.HasPrefix(parts[0], "MSH") {
			encodingCharacters = parts[1]
		}

		segment := Segment(parts[0])
		if segment == "" {
			return fmt.Errorf("%w: %s", ErrSegmentInvalid, segment)
		}

		segmentField, ok := tagToField[segment]
		if !ok {
			continue // Ignore unknown segments
		}

		// Populate the struct fields with parsed values
		if err := setValuesByIndex(segment, segmentField, parts, fieldSeparator, encodingCharacters, 0); err != nil {
			return err
		}
	}

	return nil
}

var ErrFieldIndexOutOfBounds = errors.New("hl7: field index out of bounds")

// setValuesByIndex maps HL7 field values to struct fields using the hl7 tags.
func setValuesByIndex(segment Segment, parent reflect.Value, fields []string, fs, ec string, level uint) error {
	componentSeparator := string(ec[0])

	for i := 0; i < parent.NumField(); i++ {
		parentField := parent.Field(i)
		sIndex, err := getHL7FieldIndexFromTag(parent.Type().Field(i).Tag.Get("hl7"))
		if err != nil {
			return err
		}

		// HL7 is 1-based, so we need to decrement the index
		sIndex = sIndex - 1

		// If the field index is out of bounds or invalid, skip assignment (treat as optional)
		if sIndex >= len(fields) || sIndex < 0 {
			continue
		}

		sField := fields[sIndex]
		shouldSetFS := segment == "MSH" && level == 0 && sIndex == 0
		if shouldSetFS {
			sField = fs
		}

		components := strings.Split(sField, componentSeparator)

		// TODO: Add support for subcomponent separator
		// TODO: Add support for repetition separator
		shouldParseComponents := len(components) > 1 && parentField.Kind() == reflect.Struct

		// Handle components recursively
		if shouldParseComponents {
			if err := setValuesByIndex(segment, parentField, components, fs, ec, level+1); err != nil {
				return err
			}

			continue
		}

		// If the destination is a struct but we don't have component separators,
		// skip assignment (treat as optional/empty) to avoid unsupported kind errors.
		if parentField.Kind() == reflect.Struct {
			continue
		}

		// Set field value based on its type
		if err := setFieldValue(parentField, sField); err != nil {
			return err
		}
	}

	return nil
}

var errTagEmpty = errors.New("hl7: tag is empty")

// getHL7SegmentTypeFromTag parses the "hl7" tag to extract the segment name.
func getHL7SegmentTypeFromTag(tag string) (Segment, error) {
	if tag == "" {
		return "", errTagEmpty
	}
	parts := strings.Split(tag, ":")
	if len(parts) < 2 || parts[0] != "segment" {
		return "", ErrTagInvalidFormat
	}
	return Segment(parts[1]), nil
}

// getHL7FieldIndexFromTag parses the "hl7" tag to extract the field index.
func getHL7FieldIndexFromTag(tag string) (int, error) {
	if tag == "" {
		return 0, errTagEmpty
	}
	return strconv.Atoi(tag) // Convert tag to int
}
