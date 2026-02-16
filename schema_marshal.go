package hl7

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MarshalWithSchema serializes a map[string]any into HL7 format using a schema
// definition and default marshal options.
func MarshalWithSchema(v map[string]any, schema *MessageSchema) ([]byte, error) {
	return MarshalWithSchemaOptions(v, schema, DefaultMarshalOptions())
}

// MarshalWithSchemaOptions serializes a map[string]any into HL7 format using a schema
// definition and the provided marshal options.
func MarshalWithSchemaOptions(v map[string]any, schema *MessageSchema, opts MarshalOptions) ([]byte, error) {
	fs := string(opts.FieldSeparator)
	cs := string(opts.ComponentSeparator)
	rs := string(opts.RepetitionSeparator)
	ec := string([]byte{
		opts.ComponentSeparator,
		opts.RepetitionSeparator,
		opts.EscapeCharacter,
		opts.SubcomponentSeparator,
	})

	var buf bytes.Buffer

	// We need a stable ordering for segments. HL7 messages typically start with MSH.
	// Collect segment names in schema order, but ensure MSH comes first.
	segNames := make([]string, 0, len(schema.Segments))
	hasMSH := false
	for name := range schema.Segments {
		if name == "MSH" {
			hasMSH = true
			continue
		}
		segNames = append(segNames, name)
	}
	if hasMSH {
		segNames = append([]string{"MSH"}, segNames...)
	}

	first := true
	for _, segName := range segNames {
		segData, ok := v[segName]
		if !ok {
			continue
		}

		segSchema := schema.Segments[segName]

		if segSchema.Repeat {
			arr, ok := segData.([]any)
			if !ok {
				continue
			}
			for _, item := range arr {
				segMap, ok := item.(map[string]any)
				if !ok {
					continue
				}
				line, err := marshalSegmentFromMap(segName, segMap, segSchema, fs, cs, rs, ec, opts)
				if err != nil {
					return nil, err
				}
				if !first {
					buf.WriteString(opts.LineEnding)
				}
				buf.Write(line)
				first = false
			}
		} else {
			segMap, ok := segData.(map[string]any)
			if !ok {
				continue
			}

			line, err := marshalSegmentFromMap(segName, segMap, segSchema, fs, cs, rs, ec, opts)
			if err != nil {
				return nil, err
			}

			if !first {
				buf.WriteString(opts.LineEnding)
			}
			buf.Write(line)
			first = false
		}
	}

	return buf.Bytes(), nil
}

func marshalSegmentFromMap(name string, data map[string]any, schema *SegmentSchema, fs, cs, rs, ec string, opts MarshalOptions) ([]byte, error) {
	// Find max field index
	maxIdx := 0
	for _, fieldSchema := range schema.Fields {
		if fieldSchema.Index > maxIdx {
			maxIdx = fieldSchema.Index
		}
	}

	// Build index-to-name lookup
	indexToName := make(map[int]string, len(schema.Fields))
	for fieldName, fieldSchema := range schema.Fields {
		indexToName[fieldSchema.Index] = fieldName
	}

	isMSH := name == "MSH"

	var buf bytes.Buffer
	buf.WriteString(name)

	for idx := 1; idx <= maxIdx; idx++ {
		// MSH-1 is the field separator itself
		if isMSH && idx == 1 {
			buf.WriteByte(opts.FieldSeparator)
			continue
		}

		// MSH-2 follows directly after MSH-1 (no extra separator)
		if !(isMSH && idx == 2) {
			buf.WriteString(fs)
		}

		// MSH-2 is encoding characters
		if isMSH && idx == 2 {
			buf.WriteString(ec)
			continue
		}

		fieldName, ok := indexToName[idx]
		if !ok {
			continue
		}
		fieldSchema := schema.Fields[fieldName]

		val, ok := data[fieldName]
		if !ok {
			continue
		}

		str, err := marshalValueFromMap(val, fieldSchema, cs, rs)
		if err != nil {
			return nil, fmt.Errorf("hl7: %s.%d: %w", name, idx, err)
		}
		buf.WriteString(str)
	}

	return buf.Bytes(), nil
}

func marshalValueFromMap(val any, schema *FieldSchema, cs, rs string) (string, error) {
	if val == nil {
		return "", nil
	}

	switch schema.Type {
	case SchemaTypeObject:
		return marshalObjectFromMap(val, schema, cs)
	case SchemaTypeArray:
		return marshalArrayFromMap(val, schema, cs, rs)
	default:
		return marshalScalarValue(val, schema.Type)
	}
}

func marshalObjectFromMap(val any, schema *FieldSchema, cs string) (string, error) {
	m, ok := val.(map[string]any)
	if !ok {
		return "", fmt.Errorf("expected map[string]any for object type, got %T", val)
	}

	// Find max component index
	maxIdx := 0
	for _, compSchema := range schema.Components {
		if compSchema.Index > maxIdx {
			maxIdx = compSchema.Index
		}
	}

	// Build index-to-name lookup
	indexToName := make(map[int]string, len(schema.Components))
	for compName, compSchema := range schema.Components {
		indexToName[compSchema.Index] = compName
	}

	parts := make([]string, maxIdx)
	for compName, compSchema := range schema.Components {
		compVal, ok := m[compName]
		if !ok {
			continue
		}
		str, err := marshalScalarValue(compVal, compSchema.Type)
		if err != nil {
			return "", err
		}
		parts[compSchema.Index-1] = str
	}

	// Trim trailing empty parts
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}

	return strings.Join(parts, cs), nil
}

func marshalArrayFromMap(val any, schema *FieldSchema, cs, rs string) (string, error) {
	arr, ok := val.([]any)
	if !ok {
		return "", fmt.Errorf("expected []any for array type, got %T", val)
	}

	parts := make([]string, 0, len(arr))
	for _, item := range arr {
		switch schema.Items.Type {
		case SchemaTypeObject:
			str, err := marshalObjectFromMap(item, schema.Items, cs)
			if err != nil {
				return "", err
			}
			parts = append(parts, str)
		default:
			str, err := marshalScalarValue(item, schema.Items.Type)
			if err != nil {
				return "", err
			}
			parts = append(parts, str)
		}
	}

	return strings.Join(parts, rs), nil
}

func marshalScalarValue(val any, typ SchemaType) (string, error) {
	switch typ {
	case SchemaTypeString:
		s, ok := val.(string)
		if !ok {
			return fmt.Sprintf("%v", val), nil
		}
		return s, nil
	case SchemaTypeInt:
		switch v := val.(type) {
		case int64:
			return strconv.FormatInt(v, 10), nil
		case int:
			return strconv.Itoa(v), nil
		case float64:
			return strconv.FormatInt(int64(v), 10), nil
		default:
			return fmt.Sprintf("%v", val), nil
		}
	case SchemaTypeFloat:
		switch v := val.(type) {
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), nil
		case float32:
			return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
		default:
			return fmt.Sprintf("%v", val), nil
		}
	case SchemaTypeBool:
		b, ok := val.(bool)
		if !ok {
			return fmt.Sprintf("%v", val), nil
		}
		if b {
			return "Y", nil
		}
		return "N", nil
	case SchemaTypeTimestamp:
		switch v := val.(type) {
		case time.Time:
			if v.IsZero() {
				return "", nil
			}
			return v.Format("20060102150405"), nil
		case string:
			return v, nil
		default:
			return fmt.Sprintf("%v", val), nil
		}
	default:
		return fmt.Sprintf("%v", val), nil
	}
}
