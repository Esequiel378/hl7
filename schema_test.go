package hl7_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/esequiel378/hl7"
)

func TestParseSchemaValid(t *testing.T) {
	data := []byte(`{
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

	schema, err := hl7.ParseSchema(data)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	if schema == nil {
		t.Fatal("expected non-nil schema")
	}

	if len(schema.Segments) != 1 {
		t.Errorf("expected 1 segment, got %d", len(schema.Segments))
	}

	msh := schema.Segments["MSH"]
	if msh == nil {
		t.Fatal("expected MSH segment")
	}

	if len(msh.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(msh.Fields))
	}
}

func TestParseSchemaInvalidJSON(t *testing.T) {
	_, err := hl7.ParseSchema([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseSchemaNoSegments(t *testing.T) {
	_, err := hl7.ParseSchema([]byte(`{"segments": {}}`))
	if err == nil {
		t.Fatal("expected error for empty segments")
	}
}

func TestParseSchemaNoFields(t *testing.T) {
	_, err := hl7.ParseSchema([]byte(`{"segments": {"MSH": {"fields": {}}}}`))
	if err == nil {
		t.Fatal("expected error for empty fields")
	}
}

func TestParseSchemaInvalidType(t *testing.T) {
	data := []byte(`{
		"segments": {
			"MSH": {
				"fields": {
					"1": { "name": "test", "type": "invalid" }
				}
			}
		}
	}`)
	_, err := hl7.ParseSchema(data)
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestParseSchemaObjectWithoutComponents(t *testing.T) {
	data := []byte(`{
		"segments": {
			"MSH": {
				"fields": {
					"1": { "name": "test", "type": "object" }
				}
			}
		}
	}`)
	_, err := hl7.ParseSchema(data)
	if err == nil {
		t.Fatal("expected error for object without components")
	}
}

func TestParseSchemaArrayWithoutItems(t *testing.T) {
	data := []byte(`{
		"segments": {
			"MSH": {
				"fields": {
					"1": { "name": "test", "type": "array" }
				}
			}
		}
	}`)
	_, err := hl7.ParseSchema(data)
	if err == nil {
		t.Fatal("expected error for array without items")
	}
}

func TestParseSchemaFieldWithoutName(t *testing.T) {
	data := []byte(`{
		"segments": {
			"MSH": {
				"fields": {
					"1": { "name": "", "type": "string" }
				}
			}
		}
	}`)
	_, err := hl7.ParseSchema(data)
	if err == nil {
		t.Fatal("expected error for field without name")
	}
}

func TestLoadSchemaFile(t *testing.T) {
	data := []byte(`{
		"segments": {
			"MSH": {
				"fields": {
					"1": { "name": "fieldSeparator", "type": "string" }
				}
			}
		}
	}`)

	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	schema, err := hl7.LoadSchemaFile(path)
	if err != nil {
		t.Fatalf("LoadSchemaFile failed: %v", err)
	}

	if schema == nil {
		t.Fatal("expected non-nil schema")
	}
}

func TestLoadSchemaFileNotFound(t *testing.T) {
	_, err := hl7.LoadSchemaFile("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}
