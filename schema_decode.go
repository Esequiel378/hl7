package hl7

import (
	"fmt"
	"strconv"
	"strings"
)

// UnmarshalMultiWithSchema parses multiple HL7 messages using a schema definition,
// returning a slice of maps, one per message.
func UnmarshalMultiWithSchema(data []byte, schema *MessageSchema) ([]map[string]any, error) {
	chunks := splitMessages(data)
	results := make([]map[string]any, 0, len(chunks))
	for _, chunk := range chunks {
		result, err := UnmarshalWithSchema(chunk, schema)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// UnmarshalWithSchema parses HL7 data using a schema definition,
// returning a map[string]any with field names as keys.
func UnmarshalWithSchema(data []byte, schema *MessageSchema) (map[string]any, error) {
	segments, err := parseMessage(data)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)

	for _, seg := range segments {
		segSchema, ok := schema.Segments[string(seg.name)]
		if !ok {
			continue
		}

		segMap, err := decodeSegmentWithSchema(seg, segSchema)
		if err != nil {
			return nil, err
		}

		if len(segMap) == 0 {
			continue
		}

		if segSchema.Repeat {
			existing, ok := result[string(seg.name)]
			if ok {
				arr := existing.([]any)
				result[string(seg.name)] = append(arr, segMap)
			} else {
				result[string(seg.name)] = []any{segMap}
			}
		} else {
			result[string(seg.name)] = segMap
		}
	}

	return result, nil
}

func decodeSegmentWithSchema(seg segmentLine, schema *SegmentSchema) (map[string]any, error) {
	componentSeparator := "^"
	if len(seg.encodingCharacters) > 0 {
		componentSeparator = string(seg.encodingCharacters[0])
	}
	repetitionSeparator := ""
	if len(seg.encodingCharacters) > 1 {
		repetitionSeparator = string(seg.encodingCharacters[1])
	}

	result := make(map[string]any)

	for fieldName, fieldSchema := range schema.Fields {
		idx := fieldSchema.Index

		// HL7 field indexing: MSH-1 is the field separator (maps to parts[0] offset),
		// other segments have parts[0] as segment name.
		partsIdx := idx
		if seg.name == "MSH" {
			partsIdx = idx - 1
		}

		if partsIdx < 0 || partsIdx >= len(seg.fields) {
			continue
		}

		rawValue := seg.fields[partsIdx]

		// MSH-1 is the field separator
		if seg.name == "MSH" && idx == 1 {
			rawValue = seg.fieldSeparator
		}

		if rawValue == "" {
			continue
		}

		val, err := decodeFieldWithSchema(string(seg.name), idx, rawValue, fieldSchema, componentSeparator, repetitionSeparator)
		if err != nil {
			return nil, err
		}

		if val != nil {
			result[fieldName] = val
		}
	}

	return result, nil
}

func decodeFieldWithSchema(segName string, fieldIdx int, raw string, schema *FieldSchema, cs, rs string) (any, error) {
	switch schema.Type {
	case SchemaTypeArray:
		return decodeArrayField(segName, fieldIdx, raw, schema, cs, rs)
	case SchemaTypeObject:
		return decodeObjectField(segName, fieldIdx, raw, schema, cs)
	default:
		return coerceValue(segName, fieldIdx, 0, raw, schema.Type)
	}
}

func decodeArrayField(segName string, fieldIdx int, raw string, schema *FieldSchema, cs, rs string) (any, error) {
	var reps []string
	if rs != "" {
		reps = strings.Split(raw, rs)
	} else {
		reps = []string{raw}
	}

	items := make([]any, 0, len(reps))
	for _, rep := range reps {
		if rep == "" {
			continue
		}
		itemSchema := schema.Items
		switch itemSchema.Type {
		case SchemaTypeObject:
			val, err := decodeObjectField(segName, fieldIdx, rep, itemSchema, cs)
			if err != nil {
				return nil, err
			}
			items = append(items, val)
		default:
			val, err := coerceValue(segName, fieldIdx, 0, rep, itemSchema.Type)
			if err != nil {
				return nil, err
			}
			items = append(items, val)
		}
	}

	return items, nil
}

func decodeObjectField(segName string, fieldIdx int, raw string, schema *FieldSchema, cs string) (any, error) {
	components := strings.Split(raw, cs)
	result := make(map[string]any)

	for compName, compSchema := range schema.Components {
		idx := compSchema.Index

		// Components are 1-based
		arrIdx := idx - 1
		if arrIdx < 0 || arrIdx >= len(components) {
			continue
		}

		compValue := components[arrIdx]
		if compValue == "" {
			continue
		}

		val, err := coerceValue(segName, fieldIdx, idx, compValue, compSchema.Type)
		if err != nil {
			return nil, err
		}

		result[compName] = val
	}

	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func coerceValue(segName string, fieldIdx, compIdx int, raw string, typ SchemaType) (any, error) {
	switch typ {
	case SchemaTypeString:
		return raw, nil
	case SchemaTypeInt:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, &FieldError{
				Segment:   segName,
				Field:     fieldIdx,
				Component: compIdx,
				Value:     raw,
				Err:       fmt.Errorf("%w: %v", ErrInvalidIntValue, err),
			}
		}
		return v, nil
	case SchemaTypeFloat:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, &FieldError{
				Segment:   segName,
				Field:     fieldIdx,
				Component: compIdx,
				Value:     raw,
				Err:       fmt.Errorf("%w: %v", ErrInvalidFloatValue, err),
			}
		}
		return v, nil
	case SchemaTypeBool:
		// HL7 uses Y/N, but also support true/false
		switch strings.ToUpper(raw) {
		case "Y", "TRUE", "1":
			return true, nil
		case "N", "FALSE", "0":
			return false, nil
		default:
			return nil, &FieldError{
				Segment:   segName,
				Field:     fieldIdx,
				Component: compIdx,
				Value:     raw,
				Err:       fmt.Errorf("%w: %q", ErrInvalidBooleanValue, raw),
			}
		}
	case SchemaTypeTimestamp:
		var ts Timestamp
		if err := ts.Unmarshal([]byte(raw)); err != nil {
			return nil, &FieldError{
				Segment:   segName,
				Field:     fieldIdx,
				Component: compIdx,
				Value:     raw,
				Err:       err,
			}
		}
		return ts.Time, nil
	default:
		return raw, nil
	}
}
