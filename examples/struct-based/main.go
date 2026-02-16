// Example: struct-based HL7 parsing and marshaling.
//
// This demonstrates the traditional approach using Go structs with `hl7` tags
// to parse and produce HL7 messages.
//
// Run: go run ./examples/struct-based
package main

import (
	"fmt"
	"log"

	"github.com/esequiel378/hl7"
)

// Define component types for composite fields.
type MessageType struct {
	MessageCode  string `hl7:"1"`
	TriggerEvent string `hl7:"2"`
}

type PatientName struct {
	FamilyName string `hl7:"1"`
	GivenName  string `hl7:"2"`
	MiddleName string `hl7:"3"`
}

type PatientID struct {
	ID     string `hl7:"1"`
	IDType string `hl7:"5"`
}

// Define segments.
type MSHSegment struct {
	FieldSeparator       string      `hl7:"1"`
	EncodingCharacters   string      `hl7:"2"`
	SendingApplication   string      `hl7:"3"`
	SendingFacility      string      `hl7:"4"`
	ReceivingApplication string      `hl7:"5"`
	ReceivingFacility    string      `hl7:"6"`
	DateTimeOfMessage    string      `hl7:"7"`
	MessageType          MessageType `hl7:"9"`
	MessageControlID     string      `hl7:"10"`
	ProcessingID         string      `hl7:"11"`
	VersionID            string      `hl7:"12"`
}

type PIDSegment struct {
	SetID         string        `hl7:"1"`
	PatientIDList []PatientID   `hl7:"3"`
	PatientName   PatientName   `hl7:"5"`
	DateOfBirth   hl7.Timestamp `hl7:"7"`
	Gender        string        `hl7:"8"`
}

// Define the top-level message.
type ADTMessage struct {
	MSH MSHSegment `hl7:"segment:MSH"`
	PID PIDSegment `hl7:"segment:PID"`
}

func main() {
	raw := []byte(`MSH|^~\&|HIS|General Hospital|EHR|EHR|202501151030||ADT^A01|MSG00001|P|2.5
PID|1||12345^^^^MR~67890^^^^SS||Doe^John^A||19850315120000|M`)

	// --- Unmarshal ---
	var msg ADTMessage
	if err := hl7.Unmarshal(raw, &msg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Parsed HL7 Message (struct-based) ===")
	fmt.Printf("Sending App:   %s\n", msg.MSH.SendingApplication)
	fmt.Printf("Message Type:  %s^%s\n", msg.MSH.MessageType.MessageCode, msg.MSH.MessageType.TriggerEvent)
	fmt.Printf("Patient Name:  %s, %s %s\n", msg.PID.PatientName.FamilyName, msg.PID.PatientName.GivenName, msg.PID.PatientName.MiddleName)
	fmt.Printf("Date of Birth: %s\n", msg.PID.DateOfBirth.Time.Format("2006-01-02"))
	fmt.Printf("Gender:        %s\n", msg.PID.Gender)
	fmt.Printf("Patient IDs:\n")
	for _, id := range msg.PID.PatientIDList {
		fmt.Printf("  - %s (type: %s)\n", id.ID, id.IDType)
	}

	// --- Marshal ---
	output, err := hl7.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== Marshaled HL7 Message ===")
	fmt.Println(string(output))
}
