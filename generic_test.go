package hl7

import (
	"encoding/json"
	"testing"
)

func TestParseGeneric(t *testing.T) {
	input := "MSH|^~\\&|HIS|General Hospital|EHR|EHR|202501151030||ADT^A01|MSG00001|P|2.5\n" +
		"PID|1||123456||Doe^John^A||19850315|M|||123 Main St^^Springfield^IL^62701||555-867-5309\n" +
		"PV1|1|I|4N^401^A||||||||SUR"

	msg, err := ParseGeneric([]byte(input))
	if err != nil {
		t.Fatalf("ParseGeneric() error = %v", err)
	}

	if len(msg.Segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(msg.Segments))
	}

	// Verify MSH segment
	msh := msg.Segments[0]
	if msh.Name != "MSH" {
		t.Errorf("expected segment name MSH, got %s", msh.Name)
	}
	if len(msh.Fields) != 12 {
		t.Errorf("expected 12 MSH fields, got %d", len(msh.Fields))
	}
	// MSH-1: field separator
	if msh.Fields[0].Value != "|" {
		t.Errorf("MSH-1: expected |, got %s", msh.Fields[0].Value)
	}
	// MSH-2: encoding characters
	if msh.Fields[1].Value != "^~\\&" {
		t.Errorf("MSH-2: expected ^~\\&, got %s", msh.Fields[1].Value)
	}
	// MSH-2 should have repeats (split by ~)
	if len(msh.Fields[1].Repeats) != 2 {
		t.Errorf("MSH-2: expected 2 repeats, got %d", len(msh.Fields[1].Repeats))
	}
	// MSH-9: ADT^A01 should have components
	if msh.Fields[8].Value != "ADT^A01" {
		t.Errorf("MSH-9: expected ADT^A01, got %s", msh.Fields[8].Value)
	}
	if len(msh.Fields[8].Components) != 2 {
		t.Errorf("MSH-9: expected 2 components, got %d", len(msh.Fields[8].Components))
	}
	if msh.Fields[8].Components[0].Value != "ADT" {
		t.Errorf("MSH-9 component 1: expected ADT, got %s", msh.Fields[8].Components[0].Value)
	}
	if msh.Fields[8].Components[1].Value != "A01" {
		t.Errorf("MSH-9 component 2: expected A01, got %s", msh.Fields[8].Components[1].Value)
	}

	// Verify PID segment
	pid := msg.Segments[1]
	if pid.Name != "PID" {
		t.Errorf("expected segment name PID, got %s", pid.Name)
	}
	if len(pid.Fields) != 13 {
		t.Errorf("expected 13 PID fields, got %d", len(pid.Fields))
	}
	// PID-5: Doe^John^A
	if len(pid.Fields[4].Components) != 3 {
		t.Errorf("PID-5: expected 3 components, got %d", len(pid.Fields[4].Components))
	}
	// PID-11: address with components
	if len(pid.Fields[10].Components) != 5 {
		t.Errorf("PID-11: expected 5 components, got %d", len(pid.Fields[10].Components))
	}

	// Verify PV1 segment
	pv1 := msg.Segments[2]
	if pv1.Name != "PV1" {
		t.Errorf("expected segment name PV1, got %s", pv1.Name)
	}
	// PV1-3: 4N^401^A
	if len(pv1.Fields[2].Components) != 3 {
		t.Errorf("PV1-3: expected 3 components, got %d", len(pv1.Fields[2].Components))
	}
}

func TestParseGeneric_JSON(t *testing.T) {
	input := "MSH|^~\\&|HIS|General Hospital|EHR|EHR|202501151030||ADT^A01|MSG00001|P|2.5\n" +
		"PID|1||123456||Doe^John^A||19850315|M|||123 Main St^^Springfield^IL^62701||555-867-5309\n" +
		"PV1|1|I|4N^401^A||||||||SUR"

	msg, err := ParseGeneric([]byte(input))
	if err != nil {
		t.Fatalf("ParseGeneric() error = %v", err)
	}

	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	t.Logf("JSON output:\n%s", string(data))
}

func TestParseGeneric_WithRepetitions(t *testing.T) {
	input := "MSH|^~\\&|HIS|Hospital|EHR|EHR|202501151030||ADT^A01|MSG00001|P|2.5\n" +
		"PID|1||123~456~789"

	msg, err := ParseGeneric([]byte(input))
	if err != nil {
		t.Fatalf("ParseGeneric() error = %v", err)
	}

	pid := msg.Segments[1]
	// PID-3 should have 3 repeats
	if len(pid.Fields[2].Repeats) != 3 {
		t.Fatalf("PID-3: expected 3 repeats, got %d", len(pid.Fields[2].Repeats))
	}
	if pid.Fields[2].Repeats[0].Value != "123" {
		t.Errorf("PID-3 repeat 1: expected 123, got %s", pid.Fields[2].Repeats[0].Value)
	}
	if pid.Fields[2].Repeats[1].Value != "456" {
		t.Errorf("PID-3 repeat 2: expected 456, got %s", pid.Fields[2].Repeats[1].Value)
	}
}

func TestParseGeneric_RepetitionsWithComponents(t *testing.T) {
	input := "MSH|^~\\&|HIS|Hospital|EHR|EHR|202501151030||ADT^A01|MSG00001|P|2.5\n" +
		"PID|1||ID1^AUTH1~ID2^AUTH2"

	msg, err := ParseGeneric([]byte(input))
	if err != nil {
		t.Fatalf("ParseGeneric() error = %v", err)
	}

	pid := msg.Segments[1]
	field := pid.Fields[2]
	if len(field.Repeats) != 2 {
		t.Fatalf("expected 2 repeats, got %d", len(field.Repeats))
	}
	if len(field.Repeats[0].Components) != 2 {
		t.Errorf("repeat 1: expected 2 components, got %d", len(field.Repeats[0].Components))
	}
	if field.Repeats[0].Components[0].Value != "ID1" {
		t.Errorf("repeat 1 component 1: expected ID1, got %s", field.Repeats[0].Components[0].Value)
	}
}

func TestParseGeneric_EmptyMessage(t *testing.T) {
	msg, err := ParseGeneric([]byte(""))
	if err != nil {
		t.Fatalf("ParseGeneric() error = %v", err)
	}
	if len(msg.Segments) != 0 {
		t.Errorf("expected 0 segments, got %d", len(msg.Segments))
	}
}
