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

var (
	ErrSegmentInvalid     = errors.New("hl7: invalid segment")
	ErrSegmentTypeInvalid = errors.New("hl7: invalid segment type, expected a struct")
)

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
		if errors.Is(err, errTagEmtpy) {
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

	// Scan the HL7 message line by line
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")
		if len(parts) == 0 {
			continue
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
		if err := setValuesByIndex(segmentField, parts); err != nil {
			return err
		}
	}

	return nil
}

// setValuesByIndex maps HL7 field values to struct fields using the hl7 tags.
func setValuesByIndex(parent reflect.Value, values []string) error {
	for i := 0; i < parent.NumField(); i++ {
		field := parent.Field(i)
		idx, err := getHL7FieldIndexFromTag(parent.Type().Field(i).Tag.Get("hl7"))
		if err != nil {
			return err
		}
		if idx >= len(values) {
			continue // Skip if the index is out of bounds
		}

		value := values[idx]
		parts := strings.Split(value, "^")

		// Handle nested structs recursively
		if len(parts) > 1 && field.Kind() == reflect.Struct {
			if err := setValuesByIndex(field, parts); err != nil {
				return err
			}
			continue
		}

		// Set field value based on its type
		if err := setFieldValue(field, value); err != nil {
			return err
		}
	}
	return nil
}

var (
	errTagEmtpy            = errors.New("hl7: tag is empty")
	ErrTagInvalidFormat    = errors.New("hl7: tag is not in the correct format, expected `hl7:\"segment:<name>\"`")
	ErrInvalidBooleanValue = errors.New("hl7: invalid boolean value")
)

// getHL7SegmentTypeFromTag parses the "hl7" tag to extract the segment name.
func getHL7SegmentTypeFromTag(tag string) (Segment, error) {
	if tag == "" {
		return "", errTagEmtpy
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
		return 0, errTagEmtpy
	}
	return strconv.Atoi(tag) // Convert tag to int
}

// setFieldValue sets the value for a struct field using reflection.
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert value '%s' to int: %w", value, err)
		}
		field.SetInt(intVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("cannot convert value '%s' to float: %w", value, err)
		}
		field.SetFloat(floatVal)

	case reflect.String:
		field.SetString(value)

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("cannot convert value '%s' to bool: %w", value, ErrInvalidBooleanValue)
		}
		field.SetBool(boolVal)

	default:
		// Support for other types can be added here
		return fmt.Errorf("unsupported kind: %s", field.Kind())
	}
	return nil
}
