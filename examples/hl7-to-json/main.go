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
					"fieldSeparator":       { "index": 1 },
					"encodingCharacters":   { "index": 2 },
					"sendingApplication":   { "index": 3 },
					"sendingFacility":      { "index": 4 },
					"receivingApplication": { "index": 5 },
					"receivingFacility":    { "index": 6 },
					"dateTimeOfMessage":    { "index": 7 },
					"messageType": {
						"index": 9, "type": "object",
						"components": {
							"code":    { "index": 1 },
							"trigger": { "index": 2 }
						}
					},
					"messageControlID": { "index": 10 },
					"processingID":     { "index": 11 },
					"versionID":        { "index": 12 }
				}
			},
			"PID": {
				"fields": {
					"setID":     { "index": 1 },
					"patientID": { "index": 3 },
					"patientName": {
						"index": 5, "type": "object",
						"components": {
							"familyName": { "index": 1 },
							"givenName":  { "index": 2 },
							"middleName": { "index": 3 }
						}
					},
					"dateOfBirth": { "index": 7 },
					"gender":      { "index": 8 },
					"address":     { "index": 11 },
					"phoneNumber": { "index": 13 }
				}
			},
			"PV1": {
				"fields": {
					"setID":                   { "index": 1 },
					"patientClass":            { "index": 2 },
					"assignedPatientLocation": { "index": 3 },
					"hospitalService":         { "index": 10 }
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
