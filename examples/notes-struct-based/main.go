// Example: NTE (notes) segments with struct-based parsing.
//
// NTE segments attach notes to the preceding segment. Place a []NTE field
// tagged `hl7:"notes"` inside any segment struct to collect them.
//
// Run: go run ./examples/notes-struct-based
package main

import (
	"fmt"
	"log"

	"github.com/esequiel378/hl7"
)

type NTE struct {
	SetID   string `hl7:"1"`
	Comment string `hl7:"3"`
}

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

type PIDSegment struct {
	SetID     string `hl7:"1"`
	PatientID string `hl7:"3"`
	Notes     []NTE  `hl7:"notes"` // NTE segments following PID are collected here
}

type OBXSegment struct {
	SetID            string `hl7:"1"`
	ObservationValue string `hl7:"5"`
	Notes            []NTE  `hl7:"notes"` // each OBX accumulates its own notes
}

type ORUMessage struct {
	MSH MSHSegment `hl7:"segment:MSH"`
	PID PIDSegment `hl7:"segment:PID"`
	OBX OBXSegment `hl7:"segment:OBX"`
}

func main() {
	raw := []byte("MSH|^~\\&|LIS|Lab|EHR|Hospital|20250315120000||ORU^R01|MSG001|P|2.5\r" +
		"PID|1||PAT001\r" +
		"NTE|1||Patient fasting for 12 hours\r" +
		"NTE|2||Specimen collected at bedside\r" +
		"OBX|1||||7.4\r" +
		"NTE|1||Result within normal range")

	// --- Unmarshal ---
	var msg ORUMessage
	if err := hl7.Unmarshal(raw, &msg); err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Parsed HL7 Message (notes / struct-based) ===")
	fmt.Printf("Sending App: %s\n", msg.MSH.SendingApplication)
	fmt.Printf("Patient ID:  %s\n", msg.PID.PatientID)

	fmt.Printf("\nPID notes (%d):\n", len(msg.PID.Notes))
	for _, n := range msg.PID.Notes {
		fmt.Printf("  [%s] %s\n", n.SetID, n.Comment)
	}

	fmt.Printf("\nOBX value: %s\n", msg.OBX.ObservationValue)
	fmt.Printf("OBX notes (%d):\n", len(msg.OBX.Notes))
	for _, n := range msg.OBX.Notes {
		fmt.Printf("  [%s] %s\n", n.SetID, n.Comment)
	}

	// --- Marshal back ---
	output, err := hl7.MarshalWithOptions(msg, hl7.MarshalOptions{
		FieldSeparator:        '|',
		ComponentSeparator:    '^',
		RepetitionSeparator:   '~',
		EscapeCharacter:       '\\',
		SubcomponentSeparator: '&',
		LineEnding:            "\n",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== Marshaled HL7 Message ===")
	fmt.Println(string(output))
}
