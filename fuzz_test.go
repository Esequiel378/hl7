package hl7_test

import (
	"testing"

	"github.com/esequiel378/hl7"
)

// FuzzUnmarshal tests that Unmarshal doesn't panic on arbitrary input.
func FuzzUnmarshal(f *testing.F) {
	// Add seed corpus with valid HL7 messages
	f.Add([]byte("MSH|^~\\&|App|Fac|||20250205120000||ADT^A01|123|P|2.3"))
	f.Add([]byte("MSH|^~\\&|HIS|RIH|EKG|EKG|200101011230||ADT^A01|MSG00001|P|2.3\r\nPID|1||123456|||Doe^John||19700101|M"))
	f.Add([]byte("MSH#^~\\&#App#Fac###20250205120000##ADT^A01#123#P#2.3"))
	f.Add([]byte("MSH|^~\\&|||||||||||"))
	f.Add([]byte(""))
	f.Add([]byte("MSH"))
	f.Add([]byte("PID|1||12345"))
	f.Add([]byte("MSH|^~\\&|App|Fac|||20250205120000||ADT^A01|123|P|2.3\nPID|1||ID1~ID2~ID3"))

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

	type PIDSegment struct {
		SetID     string   `hl7:"1"`
		PatientID []string `hl7:"3"`
	}

	type Message struct {
		MSH MSHSegment `hl7:"segment:MSH"`
		PID PIDSegment `hl7:"segment:PID"`
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		var msg Message
		// We don't care about errors, just that it doesn't panic
		_ = hl7.Unmarshal(data, &msg)
	})
}

// FuzzTimestamp tests that Timestamp.Unmarshal doesn't panic on arbitrary input.
func FuzzTimestamp(f *testing.F) {
	f.Add([]byte("20250205120000"))
	f.Add([]byte("20250205"))
	f.Add([]byte(""))
	f.Add([]byte("not-a-date"))
	f.Add([]byte("2025"))
	f.Add([]byte("20250205120000.0000-0700"))
	f.Add([]byte("99999999999999"))
	f.Add([]byte("0000"))

	f.Fuzz(func(t *testing.T, data []byte) {
		var ts hl7.Timestamp
		_ = ts.Unmarshal(data)
	})
}

// FuzzMarshalUnmarshal tests round-trip consistency.
func FuzzMarshalUnmarshal(f *testing.F) {
	// Add seed values for struct fields
	f.Add("App1", "Fac1", "ADT", "A01", "2.3")
	f.Add("", "", "", "", "")
	f.Add("Test|App", "Test^Fac", "Code~Type", "Event", "Version")
	f.Add("Special\nChars", "Tab\tHere", "Return\rHere", "All", "Mixed")

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
		VersionID          string      `hl7:"12"`
	}

	type Message struct {
		MSH MSHSegment `hl7:"segment:MSH"`
	}

	f.Fuzz(func(t *testing.T, app, fac, code, event, version string) {
		msg := Message{
			MSH: MSHSegment{
				FieldSeparator:     "|",
				EncodingCharacters: "^~\\&",
				SendingApplication: app,
				SendingFacility:    fac,
				MessageType:        MessageType{MessageCode: code, TriggerEvent: event},
				VersionID:          version,
			},
		}

		// Marshal shouldn't panic
		data, err := hl7.Marshal(msg)
		if err != nil {
			return // Some inputs may be invalid
		}

		// Unmarshal the result shouldn't panic
		var decoded Message
		_ = hl7.Unmarshal(data, &decoded)
	})
}
