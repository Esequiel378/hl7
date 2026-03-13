// Example: NTE (notes) segments with schema-based parsing.
//
// Add a "notes" key to any segment schema to collect the NTE segments that
// follow it. The value is a nested segment schema describing the NTE fields.
//
// Run: go run ./examples/notes-schema-based
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/esequiel378/hl7"
)

func main() {
	schemaJSON := []byte(`{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator":     { "index": 1 },
					"encodingCharacters": { "index": 2 },
					"sendingApplication": { "index": 3 },
					"messageControlID":   { "index": 10 },
					"versionID":          { "index": 12 }
				}
			},
			"PID": {
				"fields": {
					"setID":     { "index": 1 },
					"patientID": { "index": 3 }
				},
				"notes": {
					"fields": {
						"setID":   { "index": 1, "type": "int" },
						"comment": { "index": 3 }
					}
				}
			},
			"OBX": {
				"repeat": true,
				"fields": {
					"setID":  { "index": 1, "type": "int" },
					"value":  { "index": 5 }
				},
				"notes": {
					"fields": {
						"setID":   { "index": 1, "type": "int" },
						"comment": { "index": 3 }
					}
				}
			}
		}
	}`)

	schema, err := hl7.ParseSchema(schemaJSON)
	if err != nil {
		log.Fatal(err)
	}

	raw := []byte("MSH|^~\\&|LIS|Lab|EHR|Hospital|20250315120000||ORU^R01|MSG001|P|2.5\r" +
		"PID|1||PAT001\r" +
		"NTE|1||Patient fasting for 12 hours\r" +
		"NTE|2||Specimen collected at bedside\r" +
		"OBX|1||||7.4\r" +
		"NTE|1||Result within normal range\r" +
		"OBX|2||||negative\r" +
		"NTE|1||Confirmed by repeat test")

	// --- Unmarshal ---
	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Parsed HL7 Message (notes / schema-based) ===")

	msh := result["MSH"].(map[string]any)
	fmt.Printf("Sending App: %s\n", msh["sendingApplication"])

	pid := result["PID"].(map[string]any)
	fmt.Printf("Patient ID:  %s\n", pid["patientID"])

	pidNotes := pid["notes"].([]any)
	fmt.Printf("\nPID notes (%d):\n", len(pidNotes))
	for _, n := range pidNotes {
		note := n.(map[string]any)
		fmt.Printf("  [%v] %s\n", note["setID"], note["comment"])
	}

	obxList := result["OBX"].([]any)
	fmt.Printf("\nOBX segments (%d):\n", len(obxList))
	for _, item := range obxList {
		obx := item.(map[string]any)
		fmt.Printf("  OBX[%v] value=%s\n", obx["setID"], obx["value"])
		if notes, ok := obx["notes"].([]any); ok {
			for _, n := range notes {
				note := n.(map[string]any)
				fmt.Printf("    note: %s\n", note["comment"])
			}
		}
	}

	// Print the raw input for reference
	fmt.Println("\n=== Raw HL7 Input ===")
	fmt.Println(strings.ReplaceAll(string(raw), "\r", "\n"))
}
