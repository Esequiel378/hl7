package hl7_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/esequiel378/hl7"
)

func TestMSHStandardSetup(t *testing.T) {
	type MessageType struct {
		MessageCode  string `hl7:"1"`
		TriggerEvent string `hl7:"2"`
	}

	type MSHSegment struct {
		FieldSeparator       string      `hl7:"1"`
		EncodingCharacters   string      `hl7:"2"`
		SendingApplication   string      `hl7:"3"`
		SendFacility         string      `hl7:"4"`
		ReceivingApplication string      `hl7:"5"`
		ReceivingFacility    string      `hl7:"6"`
		DateTimeOfMessage    string      `hl7:"7"`
		Security             string      `hl7:"8"`
		MessageType          MessageType `hl7:"9"`
		MessageControlID     string      `hl7:"10"`
		ProcessingID         string      `hl7:"11"`
		VersionID            string      `hl7:"12"`
	}

	type Message struct {
		MSH MSHSegment `hl7:"segment:MSH"`
	}

	tests := []struct {
		name     string
		segment  string
		expected Message
	}{
		{
			name:    "MSH",
			segment: "MSH|^~\\&|App1|Fac1|App2|Fac2|20250205120000||ADT^A01|1234|P|2.1",
			expected: Message{
				MSH: MSHSegment{
					FieldSeparator:       "|",
					EncodingCharacters:   "^~\\&",
					SendingApplication:   "App1",
					SendFacility:         "Fac1",
					ReceivingApplication: "App2",
					ReceivingFacility:    "Fac2",
					DateTimeOfMessage:    "20250205120000",
					Security:             "",
					MessageType: MessageType{
						MessageCode:  "ADT",
						TriggerEvent: "A01",
					},
					MessageControlID: "1234",
					ProcessingID:     "P",
					VersionID:        "2.1",
				},
			},
		},
		{
			name:    "MSH_trainling_pipe",
			segment: "MSH|^~\\&|Sys1|Loc1|Sys2|Loc2|20250205120101|1234|ORM^O01|5678|T|2.1|",
			expected: Message{
				MSH: MSHSegment{
					FieldSeparator:       "|",
					EncodingCharacters:   "^~\\&",
					SendingApplication:   "Sys1",
					SendFacility:         "Loc1",
					ReceivingApplication: "Sys2",
					ReceivingFacility:    "Loc2",
					DateTimeOfMessage:    "20250205120101",
					Security:             "1234",
					MessageType: MessageType{
						MessageCode:  "ORM",
						TriggerEvent: "O01",
					},
					MessageControlID: "5678",
					ProcessingID:     "T",
					VersionID:        "2.1",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var message Message
			if err := hl7.Unmarshal([]byte(tc.segment), &message); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(tc.expected, message) {
				t.Errorf("expected: %v, got: %v", tc.expected, message.MSH)
			}
		})
	}
}

func TestMSHCustomFieldSeparator(t *testing.T) {
	type MessageType struct {
		MessageCode  string `hl7:"1"`
		TriggerEvent string `hl7:"2"`
	}

	type MSHSegment struct {
		FieldSeparator       string      `hl7:"1"`
		EncodingCharacters   string      `hl7:"2"`
		SendingApplication   string      `hl7:"3"`
		SendFacility         string      `hl7:"4"`
		ReceivingApplication string      `hl7:"5"`
		ReceivingFacility    string      `hl7:"6"`
		DateTimeOfMessage    string      `hl7:"7"`
		Security             string      `hl7:"8"`
		MessageType          MessageType `hl7:"9"`
		MessageControlID     string      `hl7:"10"`
		ProcessingID         string      `hl7:"11"`
		VersionID            string      `hl7:"12"`
	}

	type Message struct {
		MSH MSHSegment `hl7:"segment:MSH"`
	}

	tests := []struct {
		name           string
		fieldSeparator string
	}{
		{
			name:           "MSH_#",
			fieldSeparator: "#",
		},
		{
			name:           "MSH_@",
			fieldSeparator: "@",
		},
		{
			name:           "MSH_$",
			fieldSeparator: "$",
		},
		{
			name:           "MSH_!",
			fieldSeparator: "!",
		},
		{
			name:           "MSH_+",
			fieldSeparator: "+",
		},
		{
			name:           "MSH_?",
			fieldSeparator: "?",
		},
	}

	segment := "MSH|^~\\&|App1|Fac1|App2|Fac2|20250205120000||ADT^A01|1234|P|2.1"

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_segment := strings.ReplaceAll(segment, "|", tc.fieldSeparator)

			var message Message
			if err := hl7.Unmarshal([]byte(_segment), &message); err != nil {
				t.Fatal(err)
			}

			expected := Message{
				MSH: MSHSegment{
					FieldSeparator:       tc.fieldSeparator,
					EncodingCharacters:   "^~\\&",
					SendingApplication:   "App1",
					SendFacility:         "Fac1",
					ReceivingApplication: "App2",
					ReceivingFacility:    "Fac2",
					DateTimeOfMessage:    "20250205120000",
					Security:             "",
					MessageType: MessageType{
						MessageCode:  "ADT",
						TriggerEvent: "A01",
					},
					MessageControlID: "1234",
					ProcessingID:     "P",
					VersionID:        "2.1",
				},
			}

			if !reflect.DeepEqual(expected, message) {
				t.Errorf("expected: %v, got: %v", expected, message.MSH)
			}
		})
	}
}

func TestMSHCustomComponentSeparator(t *testing.T) {
	type MessageType struct {
		MessageCode  string `hl7:"1"`
		TriggerEvent string `hl7:"2"`
	}

	type MSHSegment struct {
		FieldSeparator       string      `hl7:"1"`
		EncodingCharacters   string      `hl7:"2"`
		SendingApplication   string      `hl7:"3"`
		SendFacility         string      `hl7:"4"`
		ReceivingApplication string      `hl7:"5"`
		ReceivingFacility    string      `hl7:"6"`
		DateTimeOfMessage    string      `hl7:"7"`
		Security             string      `hl7:"8"`
		MessageType          MessageType `hl7:"9"`
		MessageControlID     string      `hl7:"10"`
		ProcessingID         string      `hl7:"11"`
		VersionID            string      `hl7:"12"`
	}

	type Message struct {
		MSH MSHSegment `hl7:"segment:MSH"`
	}

	tests := []struct {
		name                string
		compontentSeparator string
	}{
		{
			name:                "MSH_component_*",
			compontentSeparator: "*",
		},
		{
			name:                "MSH_component_$",
			compontentSeparator: "$",
		},
		{
			name:                "MSH_component_@",
			compontentSeparator: "@",
		},
		{
			name:                "MSH_component_+",
			compontentSeparator: "+",
		},
		{
			name:                "MSH_component_%",
			compontentSeparator: "%",
		},
	}

	segment := "MSH|^~\\&|App1|Fac1|App2|Fac2|20250205120000||ADT^A01|1234|P|2.1"

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_segment := strings.ReplaceAll(segment, "^", tc.compontentSeparator)

			var message Message
			if err := hl7.Unmarshal([]byte(_segment), &message); err != nil {
				t.Fatal(err)
			}

			expected := Message{
				MSH: MSHSegment{
					FieldSeparator:       "|",
					EncodingCharacters:   tc.compontentSeparator + "~\\&",
					SendingApplication:   "App1",
					SendFacility:         "Fac1",
					ReceivingApplication: "App2",
					ReceivingFacility:    "Fac2",
					DateTimeOfMessage:    "20250205120000",
					Security:             "",
					MessageType: MessageType{
						MessageCode:  "ADT",
						TriggerEvent: "A01",
					},
					MessageControlID: "1234",
					ProcessingID:     "P",
					VersionID:        "2.1",
				},
			}

			if !reflect.DeepEqual(expected, message) {
				t.Errorf("expected: %v, got: %v", expected, message.MSH)
			}
		})
	}
}

func TestUnmarshalMultilineMessage(t *testing.T) {
	type MessageType struct {
		MessageCode  string `hl7:"1"`
		TriggerEvent string `hl7:"2"`
	}

	type MSHSegment struct {
		FieldSeparator       string      `hl7:"1"`
		EncodingCharacters   string      `hl7:"2"`
		SendingApplication   string      `hl7:"3"`
		SendFacility         string      `hl7:"4"`
		ReceivingApplication string      `hl7:"5"`
		ReceivingFacility    string      `hl7:"6"`
		DateTimeOfMessage    string      `hl7:"7"`
		Security             string      `hl7:"8"`
		MessageType          MessageType `hl7:"9"`
		MessageControlID     string      `hl7:"10"`
		ProcessingID         string      `hl7:"11"`
		VersionID            string      `hl7:"12"`
	}

	type PatientName struct {
		FamilyName string `hl7:"1"`
		GivenName  string `hl7:"2"`
		MiddleName string `hl7:"3"`
	}

	type PIDSegment struct {
		SetID                 string      `hl7:"1"`
		PatientIdentifierList string      `hl7:"3"`
		PatientName           PatientName `hl7:"5"`
		DateOfBirth           string      `hl7:"7"`
		Gender                string      `hl7:"8"`
		Race                  string      `hl7:"10"`
		Address               string      `hl7:"11"`
		PhoneNumber           string      `hl7:"13"`
		MaritalStatus         string      `hl7:"16"`
		AccountNumber         string      `hl7:"18"`
		SocialSecurityNumber  string      `hl7:"19"`
	}

	type PV1Segment struct {
		SetID                   string `hl7:"1"`
		PatientClass            string `hl7:"2"`
		AssignedPatientLocation string `hl7:"3"`
		AttendingDoctor         string `hl7:"7"`
		HospitalService         string `hl7:"10"`
		VisitNumber             string `hl7:"15"`
		FinancialClass          string `hl7:"16"`
	}

	type Message struct {
		MSH MSHSegment `hl7:"segment:MSH"`
		PID PIDSegment `hl7:"segment:PID"`
		PV1 PV1Segment `hl7:"segment:PV1"`
	}

	raw := `MSH|^~\&|HIS|RIH|EKG|EKG|200101011230||ADT^A01|MSG00001|P|2.3
PID|1||123456^^^HOSP^MR||Doe^John^A||19700101|M||2106-3|123 Main St^^Metropolis^NY^10101||555-1234|||S||123456789|987-65-4321
PV1|1|I|2000^2012^01||||1234^Jones^Bob|||SUR|||||1234567|A0|`

	var got Message
	if err := hl7.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatal(err)
	}

	expected := Message{
		MSH: MSHSegment{
			FieldSeparator:       "|",
			EncodingCharacters:   "^~\\&",
			SendingApplication:   "HIS",
			SendFacility:         "RIH",
			ReceivingApplication: "EKG",
			ReceivingFacility:    "EKG",
			DateTimeOfMessage:    "200101011230",
			Security:             "",
			MessageType:          MessageType{MessageCode: "ADT", TriggerEvent: "A01"},
			MessageControlID:     "MSG00001",
			ProcessingID:         "P",
			VersionID:            "2.3",
		},
		PID: PIDSegment{
			SetID:                 "1",
			PatientIdentifierList: "123456^^^HOSP^MR",
			PatientName:           PatientName{FamilyName: "Doe", GivenName: "John", MiddleName: "A"},
			DateOfBirth:           "19700101",
			Gender:                "M",
			Race:                  "2106-3",
			Address:               "123 Main St^^Metropolis^NY^10101",
			PhoneNumber:           "555-1234",
			MaritalStatus:         "S",
			AccountNumber:         "123456789",
			SocialSecurityNumber:  "987-65-4321",
		},
		PV1: PV1Segment{
			SetID:                   "1",
			PatientClass:            "I",
			AssignedPatientLocation: "2000^2012^01",
			AttendingDoctor:         "1234^Jones^Bob",
			HospitalService:         "SUR",
			VisitNumber:             "1234567",
			FinancialClass:          "A0",
		},
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("expected: %+v, got: %+v", expected, got)
	}
}

func TestUnmarshalWithRepetitions(t *testing.T) {
	type PIDSegment struct {
		SetID         string   `hl7:"1"`
		PatientIDList []string `hl7:"3"` // Multiple IDs separated by ~
	}

	type Message struct {
		PID PIDSegment `hl7:"segment:PID"`
	}

	raw := `MSH|^~\&|App|Fac|||20250205120000||ADT^A01|123|P|2.3
PID|1||ID001~ID002~ID003`

	var got Message
	if err := hl7.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if got.PID.SetID != "1" {
		t.Errorf("SetID = %q, want %q", got.PID.SetID, "1")
	}

	expectedIDs := []string{"ID001", "ID002", "ID003"}
	if !reflect.DeepEqual(got.PID.PatientIDList, expectedIDs) {
		t.Errorf("PatientIDList = %v, want %v", got.PID.PatientIDList, expectedIDs)
	}
}

func TestUnmarshalWithStructSliceRepetitions(t *testing.T) {
	type PatientID struct {
		ID         string `hl7:"1"`
		CheckDigit string `hl7:"2"`
		IDType     string `hl7:"5"`
	}

	type PIDSegment struct {
		SetID         string      `hl7:"1"`
		PatientIDList []PatientID `hl7:"3"` // Multiple structured IDs
	}

	type Message struct {
		PID PIDSegment `hl7:"segment:PID"`
	}

	// PID-3 has 3 repetitions, each with components
	raw := `MSH|^~\&|App|Fac|||20250205120000||ADT^A01|123|P|2.3
PID|1||12345^1^^HOSP^MR~67890^2^^LAB^LN~99999^3^^INS^PI`

	var got Message
	if err := hl7.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(got.PID.PatientIDList) != 3 {
		t.Fatalf("PatientIDList len = %d, want 3", len(got.PID.PatientIDList))
	}

	expected := []PatientID{
		{ID: "12345", CheckDigit: "1", IDType: "MR"},
		{ID: "67890", CheckDigit: "2", IDType: "LN"},
		{ID: "99999", CheckDigit: "3", IDType: "PI"},
	}

	if !reflect.DeepEqual(got.PID.PatientIDList, expected) {
		t.Errorf("PatientIDList mismatch\ngot:  %+v\nwant: %+v", got.PID.PatientIDList, expected)
	}
}

func TestUnmarshalWithTimestamp(t *testing.T) {
	type PIDSegment struct {
		SetID       string        `hl7:"1"`
		PatientID   string        `hl7:"3"`
		DateOfBirth hl7.Timestamp `hl7:"7"`
	}

	type Message struct {
		PID PIDSegment `hl7:"segment:PID"`
	}

	raw := `PID|1||12345||||19850315120000`

	var got Message
	if err := hl7.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expectedDOB := time.Date(1985, 3, 15, 12, 0, 0, 0, time.UTC)
	if !got.PID.DateOfBirth.Time.Equal(expectedDOB) {
		t.Errorf("DateOfBirth mismatch\ngot:  %v\nwant: %v", got.PID.DateOfBirth.Time, expectedDOB)
	}

	if got.PID.SetID != "1" {
		t.Errorf("SetID = %q, want %q", got.PID.SetID, "1")
	}

	if got.PID.PatientID != "12345" {
		t.Errorf("PatientID = %q, want %q", got.PID.PatientID, "12345")
	}
}
