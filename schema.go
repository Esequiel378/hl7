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
	Repeat bool                    `json:"repeat,omitempty"`
}

// FieldSchema defines a single field, including its HL7 index, type, and optional
// components (for object types) or items (for array types).
// The field name is the map key in the parent's Fields or Components map.
// If Type is omitted, it defaults to "string".
type FieldSchema struct {
	Index      int                    `json:"index,omitempty"`
	Type       SchemaType             `json:"type,omitempty"`
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
		for fieldName, field := range seg.Fields {
			path := fmt.Sprintf("segments.%s.fields.%s", segName, fieldName)
			if err := validateField(path, field, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateField(path string, f *FieldSchema, requireIndex bool) error {
	if f == nil {
		return &SchemaError{Path: path, Err: errors.New("nil field")}
	}

	if requireIndex && f.Index <= 0 {
		return &SchemaError{Path: path, Err: fmt.Errorf("index is required and must be > 0, got %d", f.Index)}
	}

	// Default type to string
	if f.Type == "" {
		f.Type = SchemaTypeString
	}

	if !validSchemaTypes[f.Type] {
		return &SchemaError{Path: path, Err: fmt.Errorf("invalid type %q", f.Type)}
	}
	if f.Type == SchemaTypeObject {
		if len(f.Components) == 0 {
			return &SchemaError{Path: path, Err: errors.New("object type requires components")}
		}
		for compName, comp := range f.Components {
			compPath := fmt.Sprintf("%s.components.%s", path, compName)
			if err := validateField(compPath, comp, true); err != nil {
				return err
			}
		}
	}
	if f.Type == SchemaTypeArray {
		if f.Items == nil {
			return &SchemaError{Path: path, Err: errors.New("array type requires items")}
		}
		itemsPath := path + ".items"
		if err := validateField(itemsPath, f.Items, false); err != nil {
			return err
		}
	}
	return nil
}
