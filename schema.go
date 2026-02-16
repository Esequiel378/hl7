package hl7

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// SchemaType represents the type of a field in a schema.
type SchemaType string

const (
	SchemaTypeString    SchemaType = "string"
	SchemaTypeInt       SchemaType = "int"
	SchemaTypeFloat     SchemaType = "float"
	SchemaTypeBool      SchemaType = "bool"
	SchemaTypeTimestamp SchemaType = "timestamp"
	SchemaTypeObject   SchemaType = "object"
	SchemaTypeArray    SchemaType = "array"
)

var validSchemaTypes = map[SchemaType]bool{
	SchemaTypeString:    true,
	SchemaTypeInt:       true,
	SchemaTypeFloat:     true,
	SchemaTypeBool:      true,
	SchemaTypeTimestamp: true,
	SchemaTypeObject:   true,
	SchemaTypeArray:    true,
}

// MessageSchema defines the structure of an HL7 message for schema-based parsing.
type MessageSchema struct {
	Segments map[string]*SegmentSchema `json:"segments"`
}

// SegmentSchema defines the fields within an HL7 segment.
type SegmentSchema struct {
	Fields map[string]*FieldSchema `json:"fields"`
}

// FieldSchema defines a single field, including its name, type, and optional
// components (for object types) or items (for array types).
type FieldSchema struct {
	Name       string                 `json:"name"`
	Type       SchemaType             `json:"type"`
	Components map[string]*FieldSchema `json:"components,omitempty"`
	Items      *FieldSchema           `json:"items,omitempty"`
}

// ParseSchema parses a JSON schema definition into a MessageSchema.
func ParseSchema(data []byte) (*MessageSchema, error) {
	var schema MessageSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("hl7: failed to parse schema: %w", err)
	}
	if err := schema.Validate(); err != nil {
		return nil, err
	}
	return &schema, nil
}

// LoadSchemaFile reads a JSON schema file and parses it into a MessageSchema.
func LoadSchemaFile(path string) (*MessageSchema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("hl7: failed to read schema file: %w", err)
	}
	return ParseSchema(data)
}

// Validate checks the schema for structural consistency.
func (s *MessageSchema) Validate() error {
	if len(s.Segments) == 0 {
		return &SchemaError{Path: "segments", Err: errors.New("no segments defined")}
	}
	for segName, seg := range s.Segments {
		if seg == nil {
			return &SchemaError{Path: "segments." + segName, Err: errors.New("nil segment")}
		}
		if len(seg.Fields) == 0 {
			return &SchemaError{Path: "segments." + segName + ".fields", Err: errors.New("no fields defined")}
		}
		for fieldIdx, field := range seg.Fields {
			path := fmt.Sprintf("segments.%s.fields.%s", segName, fieldIdx)
			if err := validateField(path, field); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateField(path string, f *FieldSchema) error {
	if f == nil {
		return &SchemaError{Path: path, Err: errors.New("nil field")}
	}
	if f.Name == "" {
		return &SchemaError{Path: path, Err: errors.New("field name is required")}
	}
	if !validSchemaTypes[f.Type] {
		return &SchemaError{Path: path, Err: fmt.Errorf("invalid type %q", f.Type)}
	}
	if f.Type == SchemaTypeObject {
		if len(f.Components) == 0 {
			return &SchemaError{Path: path, Err: errors.New("object type requires components")}
		}
		for compIdx, comp := range f.Components {
			compPath := fmt.Sprintf("%s.components.%s", path, compIdx)
			if err := validateField(compPath, comp); err != nil {
				return err
			}
		}
	}
	if f.Type == SchemaTypeArray {
		if f.Items == nil {
			return &SchemaError{Path: path, Err: errors.New("array type requires items")}
		}
		itemsPath := path + ".items"
		if err := validateField(itemsPath, f.Items); err != nil {
			return err
		}
	}
	return nil
}
