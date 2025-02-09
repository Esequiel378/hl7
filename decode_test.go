package hl7_test

import (
	"reflect"
	"strings"
	"testing"

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
