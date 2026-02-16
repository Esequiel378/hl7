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
					"1":  { "name": "fieldSeparator",     "type": "string" },
					"2":  { "name": "encodingCharacters",  "type": "string" },
					"3":  { "name": "sendingApplication",  "type": "string" },
					"4":  { "name": "sendingFacility",     "type": "string" },
					"7":  { "name": "dateTimeOfMessage",   "type": "timestamp" },
					"9":  {
						"name": "messageType", "type": "object",
						"components": {
							"1": { "name": "code",    "type": "string" },
							"2": { "name": "trigger", "type": "string" }
						}
					},
					"10": { "name": "messageControlID",    "type": "string" },
					"11": { "name": "processingID",        "type": "string" },
					"12": { "name": "versionID",           "type": "string" }
				}
			},
			"PID": {
				"fields": {
					"1": { "name": "setID", "type": "int" },
					"3": {
						"name": "patientIDList", "type": "array",
						"items": {
							"name": "patientID", "type": "object",
							"components": {
								"1": { "name": "id",   "type": "string" },
								"5": { "name": "type", "type": "string" }
							}
						}
					},
					"5": {
						"name": "patientName", "type": "object",
						"components": {
							"1": { "name": "familyName", "type": "string" },
							"2": { "name": "givenName",  "type": "string" }
						}
					},
					"7": { "name": "dateOfBirth", "type": "timestamp" },
					"8": { "name": "gender",      "type": "string" }
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
