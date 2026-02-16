// Example: generic (schema-less) HL7 parsing.
//
// This demonstrates parsing any HL7 message without defining Go structs or
// JSON schemas. The output is a structured representation of segments, fields,
// components, and repetitions that can be serialized to JSON.
//
// Run: go run ./examples/generic
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/esequiel378/hl7"
)

func main() {
	raw := []byte("MSH|^~\\&|HIS|General Hospital|EHR|EHR|202501151030||ADT^A01|MSG00001|P|2.5\n" +
		"PID|1||12345^^^^MR~67890^^^^SS||Doe^John^A||19850315|M|||123 Main St^^Springfield^IL^62701||555-867-5309\n" +
		"PV1|1|I|4N^401^A||||||||SUR")

	// Parse the message generically — no struct or schema needed.
	msg, err := hl7.ParseGeneric(raw)
	if err != nil {
		log.Fatal(err)
	}

	// Print as JSON.
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	// You can also access the parsed data programmatically.
	fmt.Println("\n=== Programmatic Access ===")
	for _, seg := range msg.Segments {
		fmt.Printf("Segment: %s (%d fields)\n", seg.Name, len(seg.Fields))
		for _, field := range seg.Fields {
			if field.Value == "" {
				continue
			}
			fmt.Printf("  %s-%d: %s", seg.Name, field.Index, field.Value)
			if len(field.Components) > 0 {
				fmt.Printf(" (components: %d)", len(field.Components))
			}
			if len(field.Repeats) > 0 {
				fmt.Printf(" (repeats: %d)", len(field.Repeats))
			}
			fmt.Println()
		}
	}
}
