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

// segmentLine represents a parsed HL7 segment line with its name, fields, and separators.
type segmentLine struct {
	name               Segment
	fields             []string
	fieldSeparator     string
	encodingCharacters string
}

// splitMessages splits raw HL7 data into individual message byte slices,
// starting a new message at each MSH segment boundary.
func splitMessages(data []byte) [][]byte {
	// Normalize \r\n and standalone \r (HL7 standard segment terminator) to \n.
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
	lines := bytes.Split(data, []byte("\n"))
	var messages [][]byte
	var current [][]byte

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmed, []byte("MSH")) && len(current) > 0 {
			messages = append(messages, bytes.Join(current, []byte("\n")))
			current = nil
		}
		if len(trimmed) > 0 {
			current = append(current, line)
		}
	}
	if len(current) > 0 {
		messages = append(messages, bytes.Join(current, []byte("\n")))
	}
	return messages
}

// parseMessage scans raw HL7 data and returns parsed segment lines with detected separators.
func parseMessage(data []byte) ([]segmentLine, error) {
	// Normalize \r\n and standalone \r (HL7 standard segment terminator) to \n.
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))

	fieldSeparator := "|"
	encodingCharacters := "^~\\&"

	var lines []segmentLine

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "MSH") && len(line) > 3 {
			fieldSeparator = string(line[3])
		}

		parts := strings.Split(line, fieldSeparator)
		if len(parts) == 0 {
			continue
		}

		if strings.HasPrefix(parts[0], "MSH") && len(parts) > 1 {
			encodingCharacters = parts[1]
		}

		segment := Segment(parts[0])
		if segment == "" {
			return nil, fmt.Errorf("%w: %s", ErrSegmentInvalid, segment)
		}

		lines = append(lines, segmentLine{
			name:               segment,
			fields:             parts,
			fieldSeparator:     fieldSeparator,
			encodingCharacters: encodingCharacters,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("hl7: failed to read message: %w", err)
	}

	return lines, nil
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

	segments, err := parseMessage(data)
	if err != nil {
		return err
	}

	for _, seg := range segments {
		segmentField, ok := tagToField[seg.name]
		if !ok {
			continue // Ignore unknown segments
		}

		// Populate the struct fields with parsed values
		if err := setValuesByIndex(seg.name, segmentField, seg.fields, seg.fieldSeparator, seg.encodingCharacters, 0); err != nil {
			return err
		}
	}

	return nil
}

var ErrFieldIndexOutOfBounds = errors.New("hl7: field index out of bounds")

// setValuesByIndex maps HL7 field values to struct fields using the hl7 tags.
func setValuesByIndex(segment Segment, parent reflect.Value, fields []string, fs, ec string, level uint) error {
	componentSeparator := "^" // Default component separator
	if len(ec) > 0 {
		componentSeparator = string(ec[0])
	}
	repetitionSeparator := ""
	if len(ec) > 1 {
		repetitionSeparator = string(ec[1])
	}

	for i := 0; i < parent.NumField(); i++ {
		parentField := parent.Field(i)
		sIndex, err := getHL7FieldIndexFromTag(parent.Type().Field(i).Tag.Get("hl7"))
		if err != nil {
			return err
		}

		// HL7 field indexing:
		// - For MSH at level 0: MSH-1 is the field separator (not in parts array),
		//   so MSH-N maps to parts[N-1]
		// - For other segments at level 0: parts[0] is the segment name,
		//   so FIELD-N maps to parts[N]
		// - For components (level > 0): components are 1-based,
		//   so component N maps to parts[N-1]
		if segment == "MSH" || level > 0 {
			sIndex = sIndex - 1
		}

		// If the field index is out of bounds or invalid, skip assignment (treat as optional)
		if sIndex >= len(fields) || sIndex < 0 {
			continue
		}

		sField := fields[sIndex]
		shouldSetFS := segment == "MSH" && level == 0 && sIndex == 0
		if shouldSetFS {
			sField = fs
		}

		// Handle repetitions (~) for slice fields
		if parentField.Kind() == reflect.Slice && repetitionSeparator != "" {
			repetitions := strings.Split(sField, repetitionSeparator)
			sliceType := parentField.Type()
			elemType := sliceType.Elem()
			newSlice := reflect.MakeSlice(sliceType, len(repetitions), len(repetitions))

			for ri, rep := range repetitions {
				elem := newSlice.Index(ri)

				// If element is a struct, parse components recursively
				if elemType.Kind() == reflect.Struct {
					repComponents := strings.Split(rep, componentSeparator)
					if len(repComponents) > 1 {
						if err := setValuesByIndex(segment, elem, repComponents, fs, ec, level+1); err != nil {
							return err
						}
						continue
					}
				}

				if err := setFieldValue(elem, rep); err != nil {
					return &FieldError{
						Segment: string(segment),
						Field:   sIndex + 1,
						Value:   rep,
						Err:     err,
					}
				}
			}
			parentField.Set(newSlice)
			continue
		}

		components := strings.Split(sField, componentSeparator)

		// TODO: Add support for subcomponent separator
		shouldParseComponents := len(components) > 1 && parentField.Kind() == reflect.Struct

		// Handle components recursively
		if shouldParseComponents {
			if err := setValuesByIndex(segment, parentField, components, fs, ec, level+1); err != nil {
				return err
			}

			continue
		}

		// If the destination is a struct but we don't have component separators,
		// check if it implements Unmarshaler (like Timestamp). If not, skip assignment
		// (treat as optional/empty) to avoid unsupported kind errors.
		if parentField.Kind() == reflect.Struct && !implementsUnmarshaler(parentField) {
			continue
		}

		// Set field value based on its type
		if err := setFieldValue(parentField, sField); err != nil {
			return &FieldError{
				Segment: string(segment),
				Field:   sIndex + 1, // Convert back to 1-based for user-facing error
				Value:   sField,
				Err:     err,
			}
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
