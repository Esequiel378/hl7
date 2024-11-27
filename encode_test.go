package hl7

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestSetFieldValue_Time(t *testing.T) {
	// Struct with time.Time field
	type TestStruct struct {
		DateTime time.Time
	}

	tests := []struct {
		name     string
		value    string
		expected time.Time
		err      bool
	}{
		{
			name:     "Valid RFC3339 Time",
			value:    "1999-05-29T23:00:00Z",
			expected: time.Date(1999, 5, 29, 23, 0, 0, 0, time.UTC),
			err:      false,
		},
		{
			name:  "Invalid Time Format",
			value: "invalid-time",
			err:   true,
		},
		{
			name:  "Empty String fails",
			value: "",
			err:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts TestStruct
			field := reflect.ValueOf(&ts).Elem().FieldByName("DateTime")

			err := setFieldValue(field, tt.value)
			if (err != nil) != tt.err {
				t.Fatalf("setFieldValue() error = %v, wantErr %v", err, tt.err)
			}

			if !tt.err && !ts.DateTime.Equal(tt.expected) {
				t.Fatalf("setFieldValue() = %v, want %v", ts.DateTime, tt.expected)
			}
		})
	}
}

func TestSetFieldValue_TimePointer(t *testing.T) {
	// Struct with *time.Time field
	type TestStruct struct {
		DateTime *time.Time
	}

	tests := []struct {
		name     string
		value    string
		expected *time.Time
		err      bool
	}{
		{
			name:     "Valid RFC3339 Time",
			value:    "1999-05-29T23:00:00Z",
			expected: func() *time.Time { t := time.Date(1999, 5, 29, 23, 0, 0, 0, time.UTC); return &t }(),
			err:      false,
		},
		{
			name:  "Empty String (Nil Pointer)",
			value: "",
			err:   true,
		},
		{
			name:  "Invalid Time Format",
			value: "invalid-time",
			err:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts TestStruct
			field := reflect.ValueOf(&ts).Elem().FieldByName("DateTime")

			err := setFieldValue(field, tt.value)
			if (err != nil) != tt.err {
				t.Fatalf("setFieldValue() error = %v, wantErr %v", err, tt.err)
			}

			if tt.expected == nil && ts.DateTime != nil {
				t.Fatalf("setFieldValue() = %v, want nil", ts.DateTime)
			} else if tt.expected != nil && ts.DateTime == nil {
				t.Fatalf("setFieldValue() = nil, want %v", *tt.expected)
			} else if tt.expected != nil && !tt.expected.Equal(*ts.DateTime) {
				t.Fatalf("setFieldValue() = %v, want %v", *ts.DateTime, *tt.expected)
			}
		})
	}
}

func TestSetFieldValue_StringPointer(t *testing.T) {
	// Struct with *string field
	type TestStruct struct {
		EncodingCharacters *string
	}

	tests := []struct {
		name     string
		value    string
		expected *string
	}{
		{
			name:     "Empty String (Nil Pointer)",
			value:    "",
			expected: func() *string { var s string; return &s }(),
		},
		{
			name:     "Non-Empty String",
			value:    "*",
			expected: func() *string { s := "*"; return &s }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ts TestStruct
			field := reflect.ValueOf(&ts).Elem().FieldByName("EncodingCharacters")

			err := setFieldValue(field, tt.value)
			if err != nil {
				t.Fatalf("setFieldValue() error = %v", err)
			}

			if tt.expected == nil && ts.EncodingCharacters != nil {
				t.Fatalf("setFieldValue() = %v, want nil", ts.EncodingCharacters)
			} else if tt.expected != nil && ts.EncodingCharacters == nil {
				t.Fatalf("setFieldValue() = nil, want %v", *tt.expected)
			} else if tt.expected != nil && *ts.EncodingCharacters != *tt.expected {
				t.Fatalf("setFieldValue() = %v, want %v", *ts.EncodingCharacters, *tt.expected)
			}
		})
	}
}

func TestSetFieldValue_InvalidKind(t *testing.T) {
	// Struct with unsupported field type
	type TestStruct struct {
		UnsupportedField map[string]string
	}

	var ts TestStruct
	field := reflect.ValueOf(&ts).Elem().FieldByName("UnsupportedField")

	err := setFieldValue(field, "some value")
	if err == nil {
		t.Fatal("setFieldValue() did not return an error for unsupported field kind")
	}

	if errors.Is(ErrUnsupportedKind, err) {
		t.Fatalf("setFieldValue() error: %v, want %v", err, ErrUnsupportedKind)
	}
}
