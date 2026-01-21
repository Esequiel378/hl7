package hl7

import (
	"fmt"
	"time"
)

// Segment represents an HL7 segment identifier (e.g., "MSH", "PID", "OBR").
type Segment string

// timestampLayouts contains the HL7 timestamp formats in order of specificity.
// The parser tries each format until one succeeds.
var timestampLayouts = []string{
	"20060102150405.0000-0700", // Full with timezone
	"20060102150405.0000+0700", // Full with positive timezone
	"20060102150405.0000",      // Full with fractional seconds
	"20060102150405-0700",      // With timezone, no fraction
	"20060102150405+0700",      // With positive timezone, no fraction
	"20060102150405",           // YYYYMMDDHHMMSS
	"200601021504",             // YYYYMMDDHHMM
	"2006010215",               // YYYYMMDDHH
	"20060102",                 // YYYYMMDD
	"200601",                   // YYYYMM
	"2006",                     // YYYY
}

// Timestamp represents an HL7 timestamp (DTM data type).
// It implements the Unmarshaler interface to automatically parse
// HL7 date/time formats into time.Time values.
//
// Supported formats (from most to least specific):
//   - YYYYMMDDHHMMSS.SSSS±ZZZZ (full with timezone and fractional seconds)
//   - YYYYMMDDHHMMSS.SSSS (with fractional seconds)
//   - YYYYMMDDHHMMSS±ZZZZ (with timezone)
//   - YYYYMMDDHHMMSS
//   - YYYYMMDDHHMM
//   - YYYYMMDDHH
//   - YYYYMMDD
//   - YYYYMM
//   - YYYY
//
// Example usage:
//
//	type PIDSegment struct {
//	    DateOfBirth hl7.Timestamp `hl7:"7"`
//	}
type Timestamp struct {
	time.Time
}

// Unmarshal parses HL7 timestamp formats into the Timestamp.
// Empty values result in a zero time.Time.
func (t *Timestamp) Unmarshal(data []byte) error {
	s := string(data)
	if s == "" {
		t.Time = time.Time{}
		return nil
	}

	for _, layout := range timestampLayouts {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("hl7: unrecognized timestamp format: %q", s)
}

// String returns the timestamp in HL7 format (YYYYMMDDHHMMSS).
// Returns an empty string if the timestamp is zero.
func (t Timestamp) String() string {
	if t.IsZero() {
		return ""
	}
	return t.Time.Format("20060102150405")
}

// MarshalHL7 implements the Marshaler interface for HL7 serialization.
func (t Timestamp) MarshalHL7() ([]byte, error) {
	return []byte(t.String()), nil
}
