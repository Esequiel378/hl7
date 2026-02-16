// Example: schema-based HL7 parsing and marshaling.
//
// This demonstrates dynamic parsing using a JSON schema definition instead of
// Go structs. Useful when schemas are loaded at runtime from files or databases.
//
// Run: go run ./examples/schema-based
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/esequiel378/hl7"
)

func main() {
	// Define the schema as JSON. In production this could be loaded from a file
	// with hl7.LoadSchemaFile("path/to/schema.json").
	schemaJSON := []byte(`{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator":      { "index": 1 },
					"encodingCharacters":  { "index": 2 },
					"sendingApplication":  { "index": 3 },
					"sendingFacility":     { "index": 4 },
					"dateTimeOfMessage":   { "index": 7, "type": "timestamp" },
					"messageType": {
						"index": 9, "type": "object",
						"components": {
							"code":    { "index": 1 },
							"trigger": { "index": 2 }
						}
					},
					"messageControlID":    { "index": 10 },
					"processingID":        { "index": 11 },
					"versionID":           { "index": 12 }
				}
			},
			"PID": {
				"fields": {
					"setID": { "index": 1, "type": "int" },
					"patientIDList": {
						"index": 3, "type": "array",
						"items": {
							"type": "object",
							"components": {
								"id":   { "index": 1 },
								"type": { "index": 5 }
							}
						}
					},
					"patientName": {
						"index": 5, "type": "object",
						"components": {
							"familyName": { "index": 1 },
							"givenName":  { "index": 2 }
						}
					},
					"dateOfBirth": { "index": 7, "type": "timestamp" },
					"gender":      { "index": 8 }
				}
			}
		}
	}`)

	schema, err := hl7.ParseSchema(schemaJSON)
	if err != nil {
		log.Fatal(err)
	}

	raw := []byte(`MSH|^~\&|HIS|General Hospital|||||ADT^A01|MSG00001|P|2.5
PID|1||12345^^^^MR~67890^^^^SS||Doe^John||19850315120000|M`)

	// --- Unmarshal ---
	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Parsed HL7 Message (schema-based) ===")

	msh := result["MSH"].(map[string]any)
	fmt.Printf("Sending App:  %s\n", msh["sendingApplication"])
	mt := msh["messageType"].(map[string]any)
	fmt.Printf("Message Type: %s^%s\n", mt["code"], mt["trigger"])
	fmt.Printf("Control ID:   %s\n", msh["messageControlID"])

	pid := result["PID"].(map[string]any)
	fmt.Printf("Set ID:       %d\n", pid["setID"])
	name := pid["patientName"].(map[string]any)
	fmt.Printf("Patient Name: %s, %s\n", name["familyName"], name["givenName"])
	dob := pid["dateOfBirth"].(time.Time)
	fmt.Printf("Date of Birth: %s\n", dob.Format("2006-01-02"))
	fmt.Printf("Gender:        %s\n", pid["gender"])

	ids := pid["patientIDList"].([]any)
	fmt.Println("Patient IDs:")
	for _, item := range ids {
		id := item.(map[string]any)
		fmt.Printf("  - %s (type: %s)\n", id["id"], id["type"])
	}

	// --- Marshal back ---
	opts := hl7.DefaultMarshalOptions()
	opts.LineEnding = "\n"
	output, err := hl7.MarshalWithSchemaOptions(result, schema, opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n=== Marshaled HL7 Message ===")
	fmt.Println(string(output))
}
