package hl7_test

import (
	"strings"
	"testing"
	"time"

	"github.com/esequiel378/hl7"
)

func TestMarshalWithSchemaBasicMSH(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator":     { "index": 1 },
					"encodingCharacters": { "index": 2 },
					"sendingApplication": { "index": 3 },
					"sendingFacility":    { "index": 4 },
					"messageControlID":   { "index": 10 },
					"versionID":          { "index": 12 }
				}
			}
		}
	}`)

	data := map[string]any{
		"MSH": map[string]any{
			"fieldSeparator":     "|",
			"encodingCharacters": "^~\\&",
			"sendingApplication": "App1",
			"sendingFacility":    "Fac1",
			"messageControlID":   "1234",
			"versionID":          "2.3",
		},
	}

	result, err := hl7.MarshalWithSchema(data, schema)
	if err != nil {
		t.Fatalf("MarshalWithSchema failed: %v", err)
	}

	str := string(result)
	if !strings.HasPrefix(str, "MSH|^~\\&|") {
		t.Errorf("expected MSH|^~\\&| prefix, got: %s", str)
	}
	if !strings.Contains(str, "App1") {
		t.Errorf("expected App1 in output, got: %s", str)
	}
	if !strings.Contains(str, "1234") {
		t.Errorf("expected 1234 in output, got: %s", str)
	}
}

func TestMarshalWithSchemaComponents(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator":     { "index": 1 },
					"encodingCharacters": { "index": 2 },
					"sendingApplication": { "index": 3 },
					"messageType": {
						"index": 9, "type": "object",
						"components": {
							"code":    { "index": 1 },
							"trigger": { "index": 2 }
						}
					}
				}
			}
		}
	}`)

	data := map[string]any{
		"MSH": map[string]any{
			"fieldSeparator":     "|",
			"encodingCharacters": "^~\\&",
			"sendingApplication": "App1",
			"messageType": map[string]any{
				"code":    "ADT",
				"trigger": "A01",
			},
		},
	}

	result, err := hl7.MarshalWithSchema(data, schema)
	if err != nil {
		t.Fatalf("MarshalWithSchema failed: %v", err)
	}

	str := string(result)
	if !strings.Contains(str, "ADT^A01") {
		t.Errorf("expected ADT^A01, got: %s", str)
	}
}

func TestMarshalWithSchemaArray(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"PID": {
				"fields": {
					"setID": { "index": 1 },
					"patientIDList": {
						"index": 3, "type": "array",
						"items": { "type": "string" }
					}
				}
			}
		}
	}`)

	data := map[string]any{
		"PID": map[string]any{
			"setID":         "1",
			"patientIDList": []any{"ID001", "ID002", "ID003"},
		},
	}

	result, err := hl7.MarshalWithSchema(data, schema)
	if err != nil {
		t.Fatalf("MarshalWithSchema failed: %v", err)
	}

	str := string(result)
	if !strings.Contains(str, "ID001~ID002~ID003") {
		t.Errorf("expected ID001~ID002~ID003, got: %s", str)
	}
}

func TestMarshalWithSchemaArrayOfObjects(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"PID": {
				"fields": {
					"patientIDList": {
						"index": 3, "type": "array",
						"items": {
							"type": "object",
							"components": {
								"id":   { "index": 1 },
								"type": { "index": 5 }
							}
						}
					}
				}
			}
		}
	}`)

	data := map[string]any{
		"PID": map[string]any{
			"patientIDList": []any{
				map[string]any{"id": "12345", "type": "MR"},
				map[string]any{"id": "67890", "type": "LN"},
			},
		},
	}

	result, err := hl7.MarshalWithSchema(data, schema)
	if err != nil {
		t.Fatalf("MarshalWithSchema failed: %v", err)
	}

	str := string(result)
	if !strings.Contains(str, "12345^^^^MR~67890^^^^LN") {
		t.Errorf("expected 12345^^^^MR~67890^^^^LN, got: %s", str)
	}
}

func TestMarshalWithSchemaMultipleSegments(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator":     { "index": 1 },
					"encodingCharacters": { "index": 2 },
					"sendingApplication": { "index": 3 }
				}
			},
			"PID": {
				"fields": {
					"setID":     { "index": 1 },
					"patientID": { "index": 3 }
				}
			}
		}
	}`)

	data := map[string]any{
		"MSH": map[string]any{
			"fieldSeparator":     "|",
			"encodingCharacters": "^~\\&",
			"sendingApplication": "App1",
		},
		"PID": map[string]any{
			"setID":     "1",
			"patientID": "12345",
		},
	}

	result, err := hl7.MarshalWithSchema(data, schema)
	if err != nil {
		t.Fatalf("MarshalWithSchema failed: %v", err)
	}

	str := string(result)
	if !strings.HasPrefix(str, "MSH|") {
		t.Errorf("expected MSH prefix, got: %s", str)
	}
	if !strings.Contains(str, "PID|1||12345") {
		t.Errorf("expected PID|1||12345, got: %s", str)
	}
}

func TestMarshalWithSchemaTypedValues(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"OBX": {
				"fields": {
					"setID":  { "index": 1, "type": "int" },
					"value":  { "index": 5, "type": "float" },
					"active": { "index": 11, "type": "bool" }
				}
			}
		}
	}`)

	data := map[string]any{
		"OBX": map[string]any{
			"setID":  int64(1),
			"value":  98.6,
			"active": true,
		},
	}

	result, err := hl7.MarshalWithSchema(data, schema)
	if err != nil {
		t.Fatalf("MarshalWithSchema failed: %v", err)
	}

	str := string(result)
	if !strings.Contains(str, "|1|") {
		t.Errorf("expected |1| for setID, got: %s", str)
	}
	if !strings.Contains(str, "98.6") {
		t.Errorf("expected 98.6, got: %s", str)
	}
	if !strings.Contains(str, "|Y") {
		t.Errorf("expected |Y for active=true, got: %s", str)
	}
}

func TestMarshalWithSchemaTimestamp(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"PID": {
				"fields": {
					"dateOfBirth": { "index": 7, "type": "timestamp" }
				}
			}
		}
	}`)

	data := map[string]any{
		"PID": map[string]any{
			"dateOfBirth": time.Date(1985, 3, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	result, err := hl7.MarshalWithSchema(data, schema)
	if err != nil {
		t.Fatalf("MarshalWithSchema failed: %v", err)
	}

	str := string(result)
	if !strings.Contains(str, "19850315120000") {
		t.Errorf("expected 19850315120000, got: %s", str)
	}
}

func TestMarshalWithSchemaOptions(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator":     { "index": 1 },
					"encodingCharacters": { "index": 2 },
					"sendingApplication": { "index": 3 }
				}
			}
		}
	}`)

	data := map[string]any{
		"MSH": map[string]any{
			"fieldSeparator":     "#",
			"encodingCharacters": "*~\\&",
			"sendingApplication": "App1",
		},
	}

	opts := hl7.MarshalOptions{
		FieldSeparator:        '#',
		ComponentSeparator:    '*',
		RepetitionSeparator:   '~',
		EscapeCharacter:       '\\',
		SubcomponentSeparator: '&',
		LineEnding:            "\n",
	}

	result, err := hl7.MarshalWithSchemaOptions(data, schema, opts)
	if err != nil {
		t.Fatalf("MarshalWithSchemaOptions failed: %v", err)
	}

	str := string(result)
	if !strings.HasPrefix(str, "MSH#*~\\&#App1") {
		t.Errorf("expected custom separators, got: %s", str)
	}
}

func TestSchemaRoundTrip(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator":     { "index": 1 },
					"encodingCharacters": { "index": 2 },
					"sendingApplication": { "index": 3 },
					"sendingFacility":    { "index": 4 },
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
					"gender":    { "index": 8 }
				}
			}
		}
	}`)

	original := []byte("MSH|^~\\&|App1|Fac1|||||ADT^A01|1234|P|2.3\nPID|1||12345||||M")

	// Unmarshal
	result, err := hl7.UnmarshalWithSchema(original, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	// Marshal back (use \n line ending so Scanner can split lines)
	opts := hl7.DefaultMarshalOptions()
	opts.LineEnding = "\n"
	output, err := hl7.MarshalWithSchemaOptions(result, schema, opts)
	if err != nil {
		t.Fatalf("MarshalWithSchema failed: %v", err)
	}

	// Unmarshal again and compare
	result2, err := hl7.UnmarshalWithSchema(output, schema)
	if err != nil {
		t.Fatalf("second UnmarshalWithSchema failed: %v", err)
	}

	// Compare key fields
	msh1, ok := result["MSH"].(map[string]any)
	if !ok {
		t.Fatal("MSH missing from first unmarshal")
	}
	msh2, ok := result2["MSH"].(map[string]any)
	if !ok {
		t.Fatal("MSH missing from second unmarshal")
	}

	if msh1["sendingApplication"] != msh2["sendingApplication"] {
		t.Errorf("sendingApplication mismatch: %v vs %v",
			msh1["sendingApplication"], msh2["sendingApplication"])
	}

	mt1, ok := msh1["messageType"].(map[string]any)
	if !ok {
		t.Fatal("messageType missing from first unmarshal")
	}
	mt2, ok := msh2["messageType"].(map[string]any)
	if !ok {
		t.Fatalf("messageType missing from second unmarshal, MSH2=%v", msh2)
	}
	if mt1["code"] != mt2["code"] || mt1["trigger"] != mt2["trigger"] {
		t.Errorf("messageType mismatch: %v vs %v", mt1, mt2)
	}

	pid1, ok := result["PID"].(map[string]any)
	if !ok {
		t.Fatalf("PID missing from first unmarshal, result=%v, output=%s", result, output)
	}
	pid2, ok := result2["PID"].(map[string]any)
	if !ok {
		t.Fatalf("PID missing from second unmarshal, result2=%v, output=%s", result2, output)
	}
	if pid1["gender"] != pid2["gender"] {
		t.Errorf("gender mismatch: %v vs %v", pid1["gender"], pid2["gender"])
	}
}

func TestMarshalWithSchemaRepeatSegments(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator":     { "index": 1 },
					"encodingCharacters": { "index": 2 },
					"sendingApplication": { "index": 3 }
				}
			},
			"OBX": {
				"repeat": true,
				"fields": {
					"setID":            { "index": 1, "type": "int" },
					"valueType":        { "index": 2 },
					"observationValue": { "index": 5 }
				}
			}
		}
	}`)

	data := map[string]any{
		"MSH": map[string]any{
			"fieldSeparator":     "|",
			"encodingCharacters": "^~\\&",
			"sendingApplication": "App1",
		},
		"OBX": []any{
			map[string]any{
				"setID":            int64(1),
				"valueType":        "NM",
				"observationValue": "98.6",
			},
			map[string]any{
				"setID":            int64(2),
				"valueType":        "ST",
				"observationValue": "Normal",
			},
		},
	}

	opts := hl7.DefaultMarshalOptions()
	opts.LineEnding = "\n"
	result, err := hl7.MarshalWithSchemaOptions(data, schema, opts)
	if err != nil {
		t.Fatalf("MarshalWithSchemaOptions failed: %v", err)
	}

	str := string(result)
	if !strings.Contains(str, "OBX|1|NM|||98.6") {
		t.Errorf("expected OBX|1|NM|||98.6, got: %s", str)
	}
	if !strings.Contains(str, "OBX|2|ST|||Normal") {
		t.Errorf("expected OBX|2|ST|||Normal, got: %s", str)
	}
}
