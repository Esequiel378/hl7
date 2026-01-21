package hl7_test

import (
	"strings"
	"testing"
	"time"

	"github.com/esequiel378/hl7"
)

func TestMarshalBasic(t *testing.T) {
	type MessageType struct {
		MessageCode  string `hl7:"1"`
		TriggerEvent string `hl7:"2"`
	}

	type MSHSegment struct {
		FieldSeparator     string      `hl7:"1"`
		EncodingCharacters string      `hl7:"2"`
		SendingApplication string      `hl7:"3"`
		SendingFacility    string      `hl7:"4"`
		DateTimeOfMessage  string      `hl7:"7"`
		MessageType        MessageType `hl7:"9"`
		MessageControlID   string      `hl7:"10"`
		ProcessingID       string      `hl7:"11"`
		VersionID          string      `hl7:"12"`
	}

	type Message struct {
		MSH MSHSegment `hl7:"segment:MSH"`
	}

	msg := Message{
		MSH: MSHSegment{
			FieldSeparator:     "|",
			EncodingCharacters: "^~\\&",
			SendingApplication: "App1",
			SendingFacility:    "Fac1",
			DateTimeOfMessage:  "20250205120000",
			MessageType:        MessageType{MessageCode: "ADT", TriggerEvent: "A01"},
			MessageControlID:   "123456",
			ProcessingID:       "P",
			VersionID:          "2.3",
		},
	}

	data, err := hl7.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(data)

	// Check that it starts with MSH|^~\&
	if !strings.HasPrefix(result, "MSH|^~\\&|") {
		t.Errorf("Expected MSH|^~\\&| prefix, got: %s", result)
	}

	// Check that key fields are present
	expected := []string{
		"App1",
		"Fac1",
		"20250205120000",
		"ADT^A01",
		"123456",
		"P",
		"2.3",
	}
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected %q in output, got: %s", exp, result)
		}
	}
}

func TestMarshalMultipleSegments(t *testing.T) {
	type MessageType struct {
		MessageCode  string `hl7:"1"`
		TriggerEvent string `hl7:"2"`
	}

	type MSHSegment struct {
		FieldSeparator     string      `hl7:"1"`
		EncodingCharacters string      `hl7:"2"`
		SendingApplication string      `hl7:"3"`
		MessageType        MessageType `hl7:"9"`
		VersionID          string      `hl7:"12"`
	}

	type PatientName struct {
		FamilyName string `hl7:"1"`
		GivenName  string `hl7:"2"`
	}

	type PIDSegment struct {
		SetID       string      `hl7:"1"`
		PatientID   string      `hl7:"3"`
		PatientName PatientName `hl7:"5"`
		DateOfBirth string      `hl7:"7"`
	}

	type Message struct {
		MSH MSHSegment `hl7:"segment:MSH"`
		PID PIDSegment `hl7:"segment:PID"`
	}

	msg := Message{
		MSH: MSHSegment{
			FieldSeparator:     "|",
			EncodingCharacters: "^~\\&",
			SendingApplication: "TestApp",
			MessageType:        MessageType{MessageCode: "ADT", TriggerEvent: "A01"},
			VersionID:          "2.3",
		},
		PID: PIDSegment{
			SetID:       "1",
			PatientID:   "12345",
			PatientName: PatientName{FamilyName: "Doe", GivenName: "John"},
			DateOfBirth: "19850315",
		},
	}

	data, err := hl7.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(data)

	// Should have two segments separated by \r
	if !strings.Contains(result, "\r") {
		t.Error("Expected \\r segment separator")
	}

	// Check MSH segment
	if !strings.HasPrefix(result, "MSH|^~\\&|") {
		t.Errorf("Expected MSH prefix, got: %s", result)
	}

	// Check PID segment
	if !strings.Contains(result, "PID|1||12345||Doe^John||19850315") {
		t.Errorf("PID segment not found or malformed in: %s", result)
	}
}

func TestMarshalWithTimestamp(t *testing.T) {
	type PIDSegment struct {
		SetID       string        `hl7:"1"`
		DateOfBirth hl7.Timestamp `hl7:"7"`
	}

	type Message struct {
		PID PIDSegment `hl7:"segment:PID"`
	}

	msg := Message{
		PID: PIDSegment{
			SetID:       "1",
			DateOfBirth: hl7.Timestamp{Time: time.Date(1985, 3, 15, 12, 30, 0, 0, time.UTC)},
		},
	}

	data, err := hl7.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(data)
	if !strings.Contains(result, "19850315123000") {
		t.Errorf("Expected timestamp 19850315123000, got: %s", result)
	}
}

func TestMarshalWithSliceRepetitions(t *testing.T) {
	type PIDSegment struct {
		SetID         string   `hl7:"1"`
		PatientIDList []string `hl7:"3"`
	}

	type Message struct {
		PID PIDSegment `hl7:"segment:PID"`
	}

	msg := Message{
		PID: PIDSegment{
			SetID:         "1",
			PatientIDList: []string{"ID001", "ID002", "ID003"},
		},
	}

	data, err := hl7.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	result := string(data)
	if !strings.Contains(result, "ID001~ID002~ID003") {
		t.Errorf("Expected repetitions ID001~ID002~ID003, got: %s", result)
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	type MessageType struct {
		MessageCode  string `hl7:"1"`
		TriggerEvent string `hl7:"2"`
	}

	type MSHSegment struct {
		FieldSeparator     string      `hl7:"1"`
		EncodingCharacters string      `hl7:"2"`
		SendingApplication string      `hl7:"3"`
		SendingFacility    string      `hl7:"4"`
		MessageType        MessageType `hl7:"9"`
		MessageControlID   string      `hl7:"10"`
		VersionID          string      `hl7:"12"`
	}

	type Message struct {
		MSH MSHSegment `hl7:"segment:MSH"`
	}

	original := Message{
		MSH: MSHSegment{
			FieldSeparator:     "|",
			EncodingCharacters: "^~\\&",
			SendingApplication: "MyApp",
			SendingFacility:    "MyFac",
			MessageType:        MessageType{MessageCode: "ADT", TriggerEvent: "A01"},
			MessageControlID:   "MSG001",
			VersionID:          "2.5",
		},
	}

	// Marshal
	data, err := hl7.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded Message
	if err := hl7.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Compare
	if decoded.MSH.SendingApplication != original.MSH.SendingApplication {
		t.Errorf("SendingApplication mismatch: got %q, want %q",
			decoded.MSH.SendingApplication, original.MSH.SendingApplication)
	}
	if decoded.MSH.MessageType.MessageCode != original.MSH.MessageType.MessageCode {
		t.Errorf("MessageCode mismatch: got %q, want %q",
			decoded.MSH.MessageType.MessageCode, original.MSH.MessageType.MessageCode)
	}
	if decoded.MSH.VersionID != original.MSH.VersionID {
		t.Errorf("VersionID mismatch: got %q, want %q",
			decoded.MSH.VersionID, original.MSH.VersionID)
	}
}

func TestMarshalWithOptions(t *testing.T) {
	type MSHSegment struct {
		FieldSeparator     string `hl7:"1"`
		EncodingCharacters string `hl7:"2"`
		SendingApplication string `hl7:"3"`
	}

	type Message struct {
		MSH MSHSegment `hl7:"segment:MSH"`
	}

	msg := Message{
		MSH: MSHSegment{
			FieldSeparator:     "#",
			EncodingCharacters: "*~\\&",
			SendingApplication: "App",
		},
	}

	opts := hl7.MarshalOptions{
		FieldSeparator:        '#',
		ComponentSeparator:    '*',
		RepetitionSeparator:   '~',
		EscapeCharacter:       '\\',
		SubcomponentSeparator: '&',
		LineEnding:            "\n",
	}

	data, err := hl7.MarshalWithOptions(msg, opts)
	if err != nil {
		t.Fatalf("MarshalWithOptions failed: %v", err)
	}

	result := string(data)
	if !strings.HasPrefix(result, "MSH#*~\\&#App") {
		t.Errorf("Expected custom separators, got: %s", result)
	}
}
