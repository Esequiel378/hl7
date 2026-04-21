// Package hl7 provides encoding and decoding of HL7 v2.x healthcare
// messages in Go. It supports three parsing modes — struct-based,
// JSON schema-based, and generic (schema-less) — all of which are
// bidirectional (decode and encode) and designed to coexist in the
// same application.
//
// # Choosing a parsing mode
//
// A real HL7 integration often handles multiple vendors in different
// states at the same time. This library offers three parsing modes
// that can all be used in the same service, sharing the same internals
// (NTE attachment, timestamp parsing, error types, encoding rules).
//
// Use struct-based parsing for vendors whose message formats are
// known and stable, when you want compile-time type safety:
//
//	var msg MyMessage
//	err := hl7.Unmarshal(data, &msg)
//
// Use schema-based parsing for vendors whose formats are known but
// evolving, when you want to update a config file instead of
// redeploying:
//
//	schema, _ := hl7.LoadSchemaFile("vendor.json")
//	result, _ := hl7.UnmarshalWithSchema(data, schema)
//
// Use generic parsing for unknown or experimental vendors, and as a
// safe fallback path for messages that fail stricter parsing:
//
//	msg, _ := hl7.ParseGeneric(data)
//
// # Features
//
// The library has zero external dependencies, supports every HL7 v2.x
// version, handles NTE segment attachment automatically, and ships with
// a CLI tool for HL7-to-JSON conversion.
//
// # Tagging
//
//   - Segment fields must be tagged with `hl7:"segment:<NAME>"` where <NAME> is the 3-letter segment ID (e.g., MSH, PID).
//   - Fields within a segment are tagged with their 1-based HL7 field index: `hl7:"1"`, `hl7:"2"`, ...
//   - Component parsing is supported when the destination field is a struct: the component separator (default '^') splits the value
//     and maps to nested struct fields by their 1-based `hl7` indices as well.
//   - Repetition parsing is supported when the destination field is a slice: the repetition separator (default '~') splits the value
//     and populates slice elements.
//
// # Special MSH Handling
//
//   - MSH-1 (Field Separator) is populated with the single-character separator detected in the message (e.g., '|').
//   - MSH-2 (Encoding Characters) is populated as-is (e.g., "^~\\&").
//
// # Built-in Types
//
//   - [Timestamp] parses HL7 date/time formats (DTM) into time.Time values automatically.
//
// # Custom Types
//
//   - Implement [Unmarshaler] interface for custom parsing during Unmarshal.
//   - Implement [Marshaler] interface for custom serialization during Marshal.
//
// # Behavior
//
//   - Unknown segments are ignored during unmarshaling.
//   - Missing or out-of-bounds fields are treated as optional and skipped, leaving zero values in the destination struct.
//   - Subcomponents ('&') are not yet parsed into separate structures.
//
// # Error Handling
//
//   - [FieldError] provides context about which segment and field caused an error.
//   - Use errors.As to extract field-level error information.
//
// See the README for comprehensive examples and usage patterns.
package hl7
