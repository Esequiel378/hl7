// Package hl7 provides utilities to marshal and unmarshal HL7 v2.x messages.
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
// # Example
//
//	// Unmarshal
//	var msg MyMessage
//	err := hl7.Unmarshal(data, &msg)
//
//	// Marshal
//	data, err := hl7.Marshal(msg)
//
// See the README for comprehensive examples and usage patterns.
package hl7
