// Package hl7 provides utilities to unmarshal HL7 messages into Go structs.
//
// Tagging
//   - Segment fields must be tagged with `hl7:"segment:<NAME>"` where <NAME> is the 3-letter segment ID (e.g., MSH, PID).
//   - Fields within a segment are tagged with their 1-based HL7 field index: `hl7:"1"`, `hl7:"2"`, ...
//   - Component parsing is supported when the destination field is a struct: the component separator (default '^') splits the value
//     and maps to nested struct fields by their 1-based `hl7` indices as well.
//
// Special MSH Handling
//   - MSH-1 (Field Separator) is populated with the single-character separator detected in the message (e.g., '|').
//   - MSH-2 (Encoding Characters) is populated as-is (e.g., "^~\\&").
//
// Behavior
//   - Unknown segments are ignored.
//   - Missing or out-of-bounds fields are treated as optional and skipped, leaving zero values in the destination struct.
//   - Subcomponents ('&') and repetitions ('~') are not yet parsed into separate structures.
//
// Example usage can be found in the README.
package hl7
