package hl7_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/esequiel378/hl7"
)

func mustParseSchema(t *testing.T, data string) *hl7.MessageSchema {
	t.Helper()
	schema, err := hl7.ParseSchema([]byte(data))
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	return schema
}

func TestUnmarshalWithSchemaBasicMSH(t *testing.T) {
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

	raw := []byte("MSH|^~\\&|App1|Fac1|App2|Fac2|20250205120000||ADT^A01|1234|P|2.3")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	msh, ok := result["MSH"].(map[string]any)
	if !ok {
		t.Fatal("expected MSH segment in result")
	}

	if msh["fieldSeparator"] != "|" {
		t.Errorf("fieldSeparator = %v, want |", msh["fieldSeparator"])
	}
	if msh["encodingCharacters"] != "^~\\&" {
		t.Errorf("encodingCharacters = %v, want ^~\\&", msh["encodingCharacters"])
	}
	if msh["sendingApplication"] != "App1" {
		t.Errorf("sendingApplication = %v, want App1", msh["sendingApplication"])
	}
	if msh["sendingFacility"] != "Fac1" {
		t.Errorf("sendingFacility = %v, want Fac1", msh["sendingFacility"])
	}
	if msh["messageControlID"] != "1234" {
		t.Errorf("messageControlID = %v, want 1234", msh["messageControlID"])
	}
	if msh["versionID"] != "2.3" {
		t.Errorf("versionID = %v, want 2.3", msh["versionID"])
	}
}

func TestUnmarshalWithSchemaComponents(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator": { "index": 1 },
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

	raw := []byte("MSH|^~\\&|App1|Fac1|App2|Fac2|20250205120000||ADT^A01|1234|P|2.3")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	msh := result["MSH"].(map[string]any)
	mt := msh["messageType"].(map[string]any)

	if mt["code"] != "ADT" {
		t.Errorf("messageType.code = %v, want ADT", mt["code"])
	}
	if mt["trigger"] != "A01" {
		t.Errorf("messageType.trigger = %v, want A01", mt["trigger"])
	}
}

func TestUnmarshalWithSchemaArrayOfStrings(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator": { "index": 1 }
				}
			},
			"PID": {
				"fields": {
					"patientIDList": {
						"index": 3, "type": "array",
						"items": { "type": "string" }
					}
				}
			}
		}
	}`)

	raw := []byte("MSH|^~\\&|App|Fac|||20250205120000||ADT^A01|123|P|2.3\nPID|1||ID001~ID002~ID003")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	pid := result["PID"].(map[string]any)
	ids := pid["patientIDList"].([]any)

	expected := []any{"ID001", "ID002", "ID003"}
	if !reflect.DeepEqual(ids, expected) {
		t.Errorf("patientIDList = %v, want %v", ids, expected)
	}
}

func TestUnmarshalWithSchemaArrayOfObjects(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator": { "index": 1 }
				}
			},
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

	raw := []byte("MSH|^~\\&|App|Fac|||20250205120000||ADT^A01|123|P|2.3\nPID|1||12345^^^^MR~67890^^^^LN")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	pid := result["PID"].(map[string]any)
	ids := pid["patientIDList"].([]any)

	if len(ids) != 2 {
		t.Fatalf("expected 2 items, got %d", len(ids))
	}

	id1 := ids[0].(map[string]any)
	if id1["id"] != "12345" {
		t.Errorf("first id = %v, want 12345", id1["id"])
	}
	if id1["type"] != "MR" {
		t.Errorf("first type = %v, want MR", id1["type"])
	}

	id2 := ids[1].(map[string]any)
	if id2["id"] != "67890" {
		t.Errorf("second id = %v, want 67890", id2["id"])
	}
	if id2["type"] != "LN" {
		t.Errorf("second type = %v, want LN", id2["type"])
	}
}

func TestUnmarshalWithSchemaMultiSegment(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator":    { "index": 1 },
					"sendingApplication": { "index": 3 },
					"messageType": {
						"index": 9, "type": "object",
						"components": {
							"code":    { "index": 1 },
							"trigger": { "index": 2 }
						}
					}
				}
			},
			"PID": {
				"fields": {
					"setID":     { "index": 1 },
					"patientID": { "index": 3 },
					"gender":    { "index": 8 }
				}
			},
			"PV1": {
				"fields": {
					"patientClass":    { "index": 2 },
					"hospitalService": { "index": 10 }
				}
			}
		}
	}`)

	raw := []byte(`MSH|^~\&|HIS|RIH|EKG|EKG|200101011230||ADT^A01|MSG00001|P|2.3
PID|1||123456^^^HOSP^MR||Doe^John^A||19700101|M||2106-3
PV1|1|I|2000^2012^01||||1234^Jones^Bob|||SUR`)

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	// Check MSH
	msh := result["MSH"].(map[string]any)
	if msh["sendingApplication"] != "HIS" {
		t.Errorf("sendingApplication = %v, want HIS", msh["sendingApplication"])
	}
	mt := msh["messageType"].(map[string]any)
	if mt["code"] != "ADT" {
		t.Errorf("messageType.code = %v, want ADT", mt["code"])
	}

	// Check PID
	pid := result["PID"].(map[string]any)
	if pid["setID"] != "1" {
		t.Errorf("setID = %v, want 1", pid["setID"])
	}
	if pid["gender"] != "M" {
		t.Errorf("gender = %v, want M", pid["gender"])
	}

	// Check PV1
	pv1 := result["PV1"].(map[string]any)
	if pv1["patientClass"] != "I" {
		t.Errorf("patientClass = %v, want I", pv1["patientClass"])
	}
	if pv1["hospitalService"] != "SUR" {
		t.Errorf("hospitalService = %v, want SUR", pv1["hospitalService"])
	}
}

func TestUnmarshalWithSchemaTypeCoercion(t *testing.T) {
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

	raw := []byte("OBX|1||GL||98.6||||||Y")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	obx := result["OBX"].(map[string]any)

	if obx["setID"] != int64(1) {
		t.Errorf("setID = %v (%T), want int64(1)", obx["setID"], obx["setID"])
	}
	if obx["value"] != 98.6 {
		t.Errorf("value = %v (%T), want 98.6", obx["value"], obx["value"])
	}
	if obx["active"] != true {
		t.Errorf("active = %v (%T), want true", obx["active"], obx["active"])
	}
}

func TestUnmarshalWithSchemaTimestamp(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"PID": {
				"fields": {
					"dateOfBirth": { "index": 7, "type": "timestamp" }
				}
			}
		}
	}`)

	raw := []byte("PID|1||12345||||19850315120000")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	pid := result["PID"].(map[string]any)
	dob := pid["dateOfBirth"].(time.Time)

	expected := time.Date(1985, 3, 15, 12, 0, 0, 0, time.UTC)
	if !dob.Equal(expected) {
		t.Errorf("dateOfBirth = %v, want %v", dob, expected)
	}
}

func TestUnmarshalWithSchemaCoercionError(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"OBX": {
				"fields": {
					"setID": { "index": 1, "type": "int" }
				}
			}
		}
	}`)

	raw := []byte("OBX|not_a_number")

	_, err := hl7.UnmarshalWithSchema(raw, schema)
	if err == nil {
		t.Fatal("expected error for invalid int coercion")
	}
}

func TestUnmarshalWithSchemaUnknownSegmentIgnored(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator": { "index": 1 }
				}
			}
		}
	}`)

	raw := []byte("MSH|^~\\&|App|Fac|||20250205120000||ADT^A01|123|P|2.3\nZZZ|custom|data")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	if _, ok := result["ZZZ"]; ok {
		t.Error("ZZZ should not be in result")
	}
}

func TestUnmarshalWithSchemaMissingFieldsOmitted(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"PID": {
				"fields": {
					"setID":       { "index": 1 },
					"customField": { "index": 50 }
				}
			}
		}
	}`)

	raw := []byte("PID|1||12345")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	pid := result["PID"].(map[string]any)
	if pid["setID"] != "1" {
		t.Errorf("setID = %v, want 1", pid["setID"])
	}
	if _, ok := pid["customField"]; ok {
		t.Error("customField should be omitted for missing data")
	}
}

func TestUnmarshalWithSchemaCustomSeparators(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator":     { "index": 1 },
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

	raw := []byte("MSH#*~\\&#App1#Fac1#App2#Fac2#20250205120000##ADT*A01#1234#P#2.3")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	msh := result["MSH"].(map[string]any)
	if msh["fieldSeparator"] != "#" {
		t.Errorf("fieldSeparator = %v, want #", msh["fieldSeparator"])
	}
	if msh["sendingApplication"] != "App1" {
		t.Errorf("sendingApplication = %v, want App1", msh["sendingApplication"])
	}

	mt := msh["messageType"].(map[string]any)
	if mt["code"] != "ADT" {
		t.Errorf("messageType.code = %v, want ADT", mt["code"])
	}
	if mt["trigger"] != "A01" {
		t.Errorf("messageType.trigger = %v, want A01", mt["trigger"])
	}
}

func TestUnmarshalWithSchemaBoolVariants(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"OBX": {
				"fields": {
					"f1": { "index": 1, "type": "bool" },
					"f2": { "index": 2, "type": "bool" },
					"f3": { "index": 3, "type": "bool" },
					"f4": { "index": 4, "type": "bool" }
				}
			}
		}
	}`)

	raw := []byte("OBX|Y|N|TRUE|FALSE")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	obx := result["OBX"].(map[string]any)
	if obx["f1"] != true {
		t.Errorf("f1 = %v, want true", obx["f1"])
	}
	if obx["f2"] != false {
		t.Errorf("f2 = %v, want false", obx["f2"])
	}
	if obx["f3"] != true {
		t.Errorf("f3 = %v, want true", obx["f3"])
	}
	if obx["f4"] != false {
		t.Errorf("f4 = %v, want false", obx["f4"])
	}
}

func TestUnmarshalWithSchemaRepeatSegments(t *testing.T) {
	schema := mustParseSchema(t, `{
		"segments": {
			"MSH": {
				"fields": {
					"fieldSeparator": { "index": 1 }
				}
			},
			"OBX": {
				"repeat": true,
				"fields": {
					"setID":      { "index": 1, "type": "int" },
					"valueType":  { "index": 2 },
					"observationValue": { "index": 5 }
				}
			}
		}
	}`)

	raw := []byte("MSH|^~\\&|App|Fac|||20250205120000||ORU^R01|123|P|2.3\nOBX|1|NM|||98.6\nOBX|2|ST|||Normal")

	result, err := hl7.UnmarshalWithSchema(raw, schema)
	if err != nil {
		t.Fatalf("UnmarshalWithSchema failed: %v", err)
	}

	obxList, ok := result["OBX"].([]any)
	if !ok {
		t.Fatalf("expected OBX to be []any, got %T", result["OBX"])
	}

	if len(obxList) != 2 {
		t.Fatalf("expected 2 OBX segments, got %d", len(obxList))
	}

	obx1 := obxList[0].(map[string]any)
	if obx1["setID"] != int64(1) {
		t.Errorf("OBX[0].setID = %v, want 1", obx1["setID"])
	}
	if obx1["observationValue"] != "98.6" {
		t.Errorf("OBX[0].observationValue = %v, want 98.6", obx1["observationValue"])
	}

	obx2 := obxList[1].(map[string]any)
	if obx2["setID"] != int64(2) {
		t.Errorf("OBX[1].setID = %v, want 2", obx2["setID"])
	}
	if obx2["observationValue"] != "Normal" {
		t.Errorf("OBX[1].observationValue = %v, want Normal", obx2["observationValue"])
	}
}
