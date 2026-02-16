// Example: HL7-to-JSON conversion pipeline.
//
// This demonstrates using schema-based parsing to convert HL7 messages into
// JSON, a common integration pattern for bridging HL7 v2 systems with modern
// REST/JSON APIs.
//
// Run: go run ./examples/hl7-to-json
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/esequiel378/hl7"
)

func main() {
	schema, err := hl7.ParseSchema([]byte(`{
		"segments": {
			"MSH": {
				"fields": {
					"1":  { "name": "fieldSeparator",      "type": "string" },
					"2":  { "name": "encodingCharacters",   "type": "string" },
					"3":  { "name": "sendingApplication",   "type": "string" },
					"4":  { "name": "sendingFacility",      "type": "string" },
					"5":  { "name": "receivingApplication",  "type": "string" },
					"6":  { "name": "receivingFacility",    "type": "string" },
					"7":  { "name": "dateTimeOfMessage",    "type": "string" },
					"9":  {
						"name": "messageType", "type": "object",
						"components": {
							"1": { "name": "code",    "type": "string" },
							"2": { "name": "trigger", "type": "string" }
						}
					},
					"10": { "name": "messageControlID", "type": "string" },
					"11": { "name": "processingID",     "type": "string" },
					"12": { "name": "versionID",        "type": "string" }
				}
			},
			"PID": {
				"fields": {
					"1": { "name": "setID",     "type": "string" },
					"3": { "name": "patientID", "type": "string" },
					"5": {
						"name": "patientName", "type": "object",
						"components": {
							"1": { "name": "familyName", "type": "string" },
							"2": { "name": "givenName",  "type": "string" },
							"3": { "name": "middleName", "type": "string" }
						}
					},
					"7":  { "name": "dateOfBirth", "type": "string" },
					"8":  { "name": "gender",      "type": "string" },
					"11": { "name": "address",     "type": "string" },
					"13": { "name": "phoneNumber", "type": "string" }
				}
			},
			"PV1": {
				"fields": {
					"1":  { "name": "setID",                   "type": "string" },
					"2":  { "name": "patientClass",            "type": "string" },
					"3":  { "name": "assignedPatientLocation", "type": "string" },
					"10": { "name": "hospitalService",         "type": "string" }
				}
			}
		}
	}`))
	if err != nil {
		log.Fatal(err)
	}

	// A typical ADT^A01 message
	raw := []byte(`MSH|^~\&|HIS|General Hospital|EHR|EHR|202501151030||ADT^A01|MSG00001|P|2.5
PID|1||123456||Doe^John^A||19850315|M|||123 Main St^^Springfield^IL^62701||555-867-5309
PV1|1|I|4N^401^A||||||||SUR`)

	// Parse HL7 -> map
	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		log.Fatal(err)
	}

	// Convert map -> JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== HL7 to JSON ===")
	fmt.Println(string(jsonData))

	// You can also go the other way: JSON -> map -> HL7
	fmt.Println("\n=== JSON back to HL7 ===")

	var parsed map[string]any
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		log.Fatal(err)
	}

	opts := hl7.DefaultMarshalOptions()
	opts.LineEnding = "\n"
	hl7Data, err := hl7.MarshalWithSchemaOptions(parsed, schema, opts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(hl7Data))
}
