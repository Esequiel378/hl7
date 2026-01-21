package hl7

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

// Marshaler is the interface implemented by types that can marshal themselves
// into valid HL7 field values.
type Marshaler interface {
	MarshalHL7() ([]byte, error)
}

var marshalerType = reflect.TypeFor[Marshaler]()

// MarshalOptions configures how HL7 messages are serialized.
type MarshalOptions struct {
	// FieldSeparator is the character used to separate fields (default: |)
	FieldSeparator byte
	// ComponentSeparator is the character used to separate components (default: ^)
	ComponentSeparator byte
	// RepetitionSeparator is the character used to separate repetitions (default: ~)
	RepetitionSeparator byte
	// EscapeCharacter is the character used for escape sequences (default: \)
	EscapeCharacter byte
	// SubcomponentSeparator is the character used to separate subcomponents (default: &)
	SubcomponentSeparator byte
	// LineEnding is the line terminator for segments (default: \r)
	LineEnding string
}

// DefaultMarshalOptions returns the standard HL7 encoding options.
func DefaultMarshalOptions() MarshalOptions {
	return MarshalOptions{
		FieldSeparator:        '|',
		ComponentSeparator:    '^',
		RepetitionSeparator:   '~',
		EscapeCharacter:       '\\',
		SubcomponentSeparator: '&',
		LineEnding:            "\r",
	}
}

// Marshal serializes a struct into HL7 format using default options.
// The struct fields must be tagged with `hl7:"segment:<name>"` for segments,
// and `hl7:"<index>"` for fields within the segment.
func Marshal(v any) ([]byte, error) {
	return MarshalWithOptions(v, DefaultMarshalOptions())
}

// MarshalWithOptions serializes a struct into HL7 format using the provided options.
func MarshalWithOptions(v any, opts MarshalOptions) ([]byte, error) {
	rv := reflect.ValueOf(v)

	// Handle pointer types
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil, fmt.Errorf("hl7: Marshal(nil)")
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("hl7: Marshal requires a struct, got %s", rv.Kind())
	}

	// Build encoding characters string
	encodingChars := string([]byte{
		opts.ComponentSeparator,
		opts.RepetitionSeparator,
		opts.EscapeCharacter,
		opts.SubcomponentSeparator,
	})

	var buf bytes.Buffer

	// Collect segments with their order
	type segmentData struct {
		name  string
		value reflect.Value
		order int
	}
	var segments []segmentData

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		tag, err := getHL7SegmentTypeFromTag(rv.Type().Field(i).Tag.Get("hl7"))
		if err != nil {
			continue // Skip fields without valid segment tags
		}

		segments = append(segments, segmentData{
			name:  string(tag),
			value: field,
			order: i,
		})
	}

	// Marshal each segment
	for idx, seg := range segments {
		line, err := marshalSegment(seg.name, seg.value, opts, encodingChars)
		if err != nil {
			return nil, err
		}

		buf.Write(line)
		if idx < len(segments)-1 {
			buf.WriteString(opts.LineEnding)
		}
	}

	return buf.Bytes(), nil
}

// marshalSegment converts a segment struct to its HL7 representation.
func marshalSegment(name string, v reflect.Value, opts MarshalOptions, ec string) ([]byte, error) {
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("hl7: segment %s must be a struct", name)
	}

	fs := string(opts.FieldSeparator)
	cs := string(opts.ComponentSeparator)
	rs := string(opts.RepetitionSeparator)

	// Find the maximum field index to determine field count
	maxIndex := 0
	fieldMap := make(map[int]reflect.Value)
	fieldTypeMap := make(map[int]reflect.StructField)

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("hl7")
		if tag == "" {
			continue
		}
		idx, err := strconv.Atoi(tag)
		if err != nil {
			continue
		}
		if idx > maxIndex {
			maxIndex = idx
		}
		fieldMap[idx] = v.Field(i)
		fieldTypeMap[idx] = v.Type().Field(i)
	}

	// Build the segment
	var buf bytes.Buffer
	buf.WriteString(name)

	// For MSH, field 1 is the field separator itself
	isMSH := name == "MSH"

	for idx := 1; idx <= maxIndex; idx++ {
		// For MSH, field 1 is the separator itself (not preceded by a separator)
		// and field 2 follows directly after field 1 (no extra separator)
		if isMSH && idx == 1 {
			buf.WriteByte(opts.FieldSeparator)
			continue
		}

		// For MSH field 2, don't add a separator (field 1 was the separator)
		if !(isMSH && idx == 2) {
			buf.WriteString(fs)
		}

		field, exists := fieldMap[idx]
		if !exists {
			continue
		}

		// For MSH-2, write encoding characters
		if isMSH && idx == 2 {
			buf.WriteString(ec)
			continue
		}

		str, err := marshalValue(field, cs, rs)
		if err != nil {
			return nil, fmt.Errorf("hl7: %s.%d: %w", name, idx, err)
		}
		buf.WriteString(str)
	}

	return buf.Bytes(), nil
}

// marshalValue converts a reflect.Value to its HL7 string representation.
func marshalValue(v reflect.Value, componentSep, repetitionSep string) (string, error) {
	// Handle pointer types
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}

	// Check for Marshaler interface
	if v.CanAddr() {
		addr := v.Addr()
		if addr.Type().Implements(marshalerType) {
			if m, ok := addr.Interface().(Marshaler); ok {
				b, err := m.MarshalHL7()
				if err != nil {
					return "", err
				}
				return string(b), nil
			}
		}
	}
	if v.Type().Implements(marshalerType) {
		if m, ok := v.Interface().(Marshaler); ok {
			b, err := m.MarshalHL7()
			if err != nil {
				return "", err
			}
			return string(b), nil
		}
	}

	switch v.Kind() {
	case reflect.String:
		return v.String(), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil

	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), nil

	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil

	case reflect.Bool:
		if v.Bool() {
			return "Y", nil
		}
		return "N", nil

	case reflect.Struct:
		return marshalStruct(v, componentSep)

	case reflect.Slice:
		return marshalSlice(v, componentSep, repetitionSep)

	default:
		return "", fmt.Errorf("unsupported type: %s", v.Kind())
	}
}

// marshalStruct converts a struct to component-separated string.
func marshalStruct(v reflect.Value, componentSep string) (string, error) {
	// Find max component index
	maxIndex := 0
	compMap := make(map[int]reflect.Value)

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("hl7")
		if tag == "" {
			continue
		}
		idx, err := strconv.Atoi(tag)
		if err != nil {
			continue
		}
		if idx > maxIndex {
			maxIndex = idx
		}
		compMap[idx] = v.Field(i)
	}

	if maxIndex == 0 {
		return "", nil
	}

	var parts []string
	for idx := 1; idx <= maxIndex; idx++ {
		field, exists := compMap[idx]
		if !exists {
			parts = append(parts, "")
			continue
		}

		str, err := marshalValue(field, "", "") // No nested components for now
		if err != nil {
			return "", err
		}
		parts = append(parts, str)
	}

	// Trim trailing empty components
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	return joinWithSep(parts, componentSep), nil
}

// marshalSlice converts a slice to repetition-separated string.
func marshalSlice(v reflect.Value, componentSep, repetitionSep string) (string, error) {
	if v.Len() == 0 {
		return "", nil
	}

	var parts []string
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		str, err := marshalValue(elem, componentSep, repetitionSep)
		if err != nil {
			return "", err
		}
		parts = append(parts, str)
	}

	return joinWithSep(parts, repetitionSep), nil
}

// joinWithSep joins strings with a separator.
func joinWithSep(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}

	var buf bytes.Buffer
	for i, p := range parts {
		if i > 0 {
			buf.WriteString(sep)
		}
		buf.WriteString(p)
	}
	return buf.String()
}
