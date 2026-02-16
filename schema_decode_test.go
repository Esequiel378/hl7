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
					"1": { "name": "fieldSeparator", "type": "string" },
					"2": { "name": "encodingCharacters", "type": "string" },
					"3": { "name": "sendingApplication", "type": "string" },
					"4": { "name": "sendingFacility", "type": "string" },
					"10": { "name": "messageControlID", "type": "string" },
					"12": { "name": "versionID", "type": "string" }
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
					"1": { "name": "fieldSeparator", "type": "string" },
					"9": {
						"name": "messageType", "type": "object",
						"components": {
							"1": { "name": "code", "type": "string" },
							"2": { "name": "trigger", "type": "string" }
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
					"1": { "name": "fieldSeparator", "type": "string" }
				}
			},
			"PID": {
				"fields": {
					"3": {
						"name": "patientIDList", "type": "array",
						"items": { "name": "id", "type": "string" }
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
					"1": { "name": "fieldSeparator", "type": "string" }
				}
			},
			"PID": {
				"fields": {
					"3": {
						"name": "patientIDList", "type": "array",
						"items": {
							"name": "patientID", "type": "object",
							"components": {
								"1": { "name": "id", "type": "string" },
								"5": { "name": "type", "type": "string" }
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
					"1": { "name": "fieldSeparator", "type": "string" },
					"3": { "name": "sendingApplication", "type": "string" },
					"9": {
						"name": "messageType", "type": "object",
						"components": {
							"1": { "name": "code", "type": "string" },
							"2": { "name": "trigger", "type": "string" }
						}
					}
				}
			},
			"PID": {
				"fields": {
					"1": { "name": "setID", "type": "string" },
					"3": { "name": "patientID", "type": "string" },
					"8": { "name": "gender", "type": "string" }
				}
			},
			"PV1": {
				"fields": {
					"2": { "name": "patientClass", "type": "string" },
					"10": { "name": "hospitalService", "type": "string" }
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
					"1": { "name": "setID", "type": "int" },
					"5": { "name": "value", "type": "float" },
					"11": { "name": "active", "type": "bool" }
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
					"7": { "name": "dateOfBirth", "type": "timestamp" }
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
					"1": { "name": "setID", "type": "int" }
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
					"1": { "name": "fieldSeparator", "type": "string" }
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
					"1": { "name": "setID", "type": "string" },
					"50": { "name": "customField", "type": "string" }
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
					"1": { "name": "fieldSeparator", "type": "string" },
					"3": { "name": "sendingApplication", "type": "string" },
					"9": {
						"name": "messageType", "type": "object",
						"components": {
							"1": { "name": "code", "type": "string" },
							"2": { "name": "trigger", "type": "string" }
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
					"1": { "name": "f1", "type": "bool" },
					"2": { "name": "f2", "type": "bool" },
					"3": { "name": "f3", "type": "bool" },
					"4": { "name": "f4", "type": "bool" }
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
