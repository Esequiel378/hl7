# hl7 - A Golang Library for Parsing and Building HL7 Messages

[![Go Reference](https://pkg.go.dev/badge/github.com/esequiel378/hl7.svg)](https://pkg.go.dev/github.com/esequiel378/hl7)
[![Go Report Card](https://goreportcard.com/badge/github.com/esequiel378/hl7)](https://goreportcard.com/report/github.com/esequiel378/hl7)

A lightweight, dependency-free Go library for parsing and building HL7 v2.x messages. It provides three flexible approaches to handle one of the most common formats in healthcare, leveraging Go's robust type system and error handling.

## Features

- **Three Parsing Modes**: Struct-based, schema-based (JSON), and generic (schema-less)
- **Bidirectional Conversion**: Both parse and build HL7 messages in all modes
- **Struct Tag Parsing**: Define HL7 mappings with intuitive struct tags (`hl7:"segment:<name>"` and `hl7:"<index>"`)
- **JSON Schema Support**: Define message schemas as JSON for dynamic, runtime-configurable parsing
- **Generic Parsing**: Parse any HL7 message without structs or schemas into a structured representation
- **Nested Structs**: Manage complex fields like patient names using component separators (`^`)
- **Repetition Support**: Parse repeating fields (`~`) into Go slices
- **Timestamp Type**: Built-in `hl7.Timestamp` type for automatic date/time parsing
- **Custom Types**: Implement `Unmarshaler` or `Marshaler` interfaces for custom field handling
- **Version Agnostic**: Supports any HL7 v2.x version
- **Rich Errors**: Field-level error context for easier debugging
- **Dependency-Free**: No external dependencies

## Installation

```bash
go get github.com/esequiel378/hl7
```

## Quick Start

### Parsing (Unmarshal)

```go
package main

import (
    "fmt"
    "github.com/esequiel378/hl7"
)

func main() {
    raw := "MSH|^~\\&|App1|Fac1|||20250205120000||ADT^A04|MSG001|P|2.7"

    var message struct {
        Header struct {
            FieldSeparator     string `hl7:"1"`
            EncodingCharacters string `hl7:"2"`
            SendingApplication string `hl7:"3"`
            DateTimeOfMessage  string `hl7:"7"`
            MessageType        struct {
                Code    string `hl7:"1"`
                Trigger string `hl7:"2"`
            } `hl7:"9"`
            VersionID string `hl7:"12"`
        } `hl7:"segment:MSH"`
    }

    if err := hl7.Unmarshal([]byte(raw), &message); err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Printf("App: %s, Version: %s\n",
        message.Header.SendingApplication,
        message.Header.VersionID)
    // Output: App: App1, Version: 2.7
}
```

### Building (Marshal)

```go
package main

import (
    "fmt"
    "github.com/esequiel378/hl7"
)

func main() {
    type MessageType struct {
        Code    string `hl7:"1"`
        Trigger string `hl7:"2"`
    }

    type MSHSegment struct {
        FieldSeparator     string      `hl7:"1"`
        EncodingCharacters string      `hl7:"2"`
        SendingApplication string      `hl7:"3"`
        SendingFacility    string      `hl7:"4"`
        DateTimeOfMessage  string      `hl7:"7"`
        MessageType        MessageType `hl7:"9"`
        MessageControlID   string      `hl7:"10"`
        ProcessingID       string      `hl7:"11"`
        VersionID          string      `hl7:"12"`
    }

    type Message struct {
        MSH MSHSegment `hl7:"segment:MSH"`
    }

    msg := Message{
        MSH: MSHSegment{
            FieldSeparator:     "|",
            EncodingCharacters: "^~\\&",
            SendingApplication: "MyApp",
            SendingFacility:    "MyFac",
            DateTimeOfMessage:  "20250205120000",
            MessageType:        MessageType{Code: "ADT", Trigger: "A01"},
            MessageControlID:   "MSG001",
            ProcessingID:       "P",
            VersionID:          "2.5",
        },
    }

    data, err := hl7.Marshal(msg)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println(string(data))
    // Output: MSH|^~\&|MyApp|MyFac|||20250205120000||ADT^A01|MSG001|P|2.5
}
```

## Parsing Approaches

This library offers three ways to parse HL7 messages. Choose the one that fits your use case:

| Approach | Best For | Output Type |
|----------|----------|-------------|
| **Struct-based** | Known message formats with compile-time safety | Go structs |
| **Schema-based** | Dynamic schemas loaded at runtime (files, databases) | `map[string]any` |
| **Generic** | Exploring unknown messages or building tooling | `*GenericMessage` |

### Struct-Based

Define Go structs with `hl7` tags. Best when you know the message structure at compile time.

```go
type Message struct {
    MSH struct {
        FieldSeparator     string       `hl7:"1"`
        EncodingCharacters string       `hl7:"2"`
        SendingApplication string       `hl7:"3"`
        MessageType        struct {
            Code    string `hl7:"1"`
            Trigger string `hl7:"2"`
        } `hl7:"9"`
    } `hl7:"segment:MSH"`
    PID struct {
        SetID       string        `hl7:"1"`
        PatientName struct {
            FamilyName string `hl7:"1"`
            GivenName  string `hl7:"2"`
        } `hl7:"5"`
        DateOfBirth hl7.Timestamp `hl7:"7"`
    } `hl7:"segment:PID"`
}

var msg Message
err := hl7.Unmarshal(data, &msg)
```

### Schema-Based

Define schemas as JSON and parse at runtime. Useful when message structures are configured externally or vary between integrations.

```go
schemaJSON := []byte(`{
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
                },
                "versionID": { "index": 12 }
            }
        },
        "PID": {
            "fields": {
                "setID":       { "index": 1, "type": "int" },
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
// or load from a file:
// schema, err := hl7.LoadSchemaFile("path/to/schema.json")

result, err := hl7.UnmarshalWithSchema(data, schema)
// result is map[string]any:
// result["MSH"].(map[string]any)["sendingApplication"] => "HIS"
// result["PID"].(map[string]any)["patientName"].(map[string]any)["familyName"] => "Doe"
```

Schema field types: `string` (default), `int`, `float`, `bool`, `timestamp`, `object` (with `components`), `array` (with `items`).

Schemas can also be used for marshaling:

```go
output, err := hl7.MarshalWithSchema(result, schema)
```

### Generic (Schema-Less)

Parse any HL7 message into a structured representation without defining structs or schemas. Ideal for building tools, inspecting unknown messages, or converting to JSON.

```go
msg, err := hl7.ParseGeneric(data)

// Access segments, fields, components, and repetitions programmatically
for _, seg := range msg.Segments {
    fmt.Printf("Segment: %s\n", seg.Name)
    for _, field := range seg.Fields {
        fmt.Printf("  %s-%d: %s\n", seg.Name, field.Index, field.Value)
    }
}

// Or serialize directly to JSON
jsonData, _ := json.MarshalIndent(msg, "", "  ")
fmt.Println(string(jsonData))
```

## Advanced Usage

### Using the Timestamp Type

The `hl7.Timestamp` type automatically parses HL7 date/time formats:

```go
type PIDSegment struct {
    SetID       string        `hl7:"1"`
    PatientID   string        `hl7:"3"`
    DateOfBirth hl7.Timestamp `hl7:"7"`
}

type Message struct {
    PID PIDSegment `hl7:"segment:PID"`
}

raw := `PID|1||12345||||19850315120000`
var msg Message
hl7.Unmarshal([]byte(raw), &msg)

fmt.Println(msg.PID.DateOfBirth.Time) // 1985-03-15 12:00:00 +0000 UTC
```

Supported timestamp formats:
- `YYYYMMDDHHMMSS.SSSS±ZZZZ` (full with timezone and fractional seconds)
- `YYYYMMDDHHMMSS±ZZZZ` (with timezone)
- `YYYYMMDDHHMMSS`
- `YYYYMMDDHHMM`
- `YYYYMMDDHH`
- `YYYYMMDD`
- `YYYYMM`
- `YYYY`

### Handling Repetitions

Use slices to capture repeating fields (separated by `~`):

```go
type PIDSegment struct {
    SetID         string   `hl7:"1"`
    PatientIDList []string `hl7:"3"` // Captures ID1~ID2~ID3
}

// With structured repetitions:
type PatientID struct {
    ID     string `hl7:"1"`
    Type   string `hl7:"5"`
}

type PIDWithStructs struct {
    SetID         string      `hl7:"1"`
    PatientIDList []PatientID `hl7:"3"` // Each repetition parsed as struct
}
```

Schema-based repetitions use the `array` type with `items`:

```json
{
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
```

### Custom Field Types

Implement the `Unmarshaler` interface for custom parsing:

```go
type CustomDate struct {
    time.Time
}

func (d *CustomDate) Unmarshal(data []byte) error {
    t, err := time.Parse("20060102", string(data))
    if err != nil {
        return err
    }
    d.Time = t
    return nil
}
```

For marshaling, implement `Marshaler`:

```go
func (d CustomDate) MarshalHL7() ([]byte, error) {
    return []byte(d.Format("20060102")), nil
}
```

### Custom Separators

Use `MarshalWithOptions` for non-standard separators:

```go
opts := hl7.MarshalOptions{
    FieldSeparator:        '#',
    ComponentSeparator:    '*',
    RepetitionSeparator:   '~',
    EscapeCharacter:       '\\',
    SubcomponentSeparator: '&',
    LineEnding:            "\n",
}

data, err := hl7.MarshalWithOptions(msg, opts)
```

### HL7 to JSON Conversion

A common integration pattern for bridging HL7 v2 systems with modern REST/JSON APIs:

```go
// HL7 -> JSON
schema, _ := hl7.LoadSchemaFile("adt_a01.json")
result, _ := hl7.UnmarshalWithSchema(hl7Data, schema)
jsonData, _ := json.MarshalIndent(result, "", "  ")

// JSON -> HL7
var parsed map[string]any
json.Unmarshal(jsonData, &parsed)
hl7Data, _ := hl7.MarshalWithSchema(parsed, schema)
```

## Error Handling

Errors include field-level context for debugging:

```go
err := hl7.Unmarshal(data, &msg)
if err != nil {
    var fieldErr *hl7.FieldError
    if errors.As(err, &fieldErr) {
        fmt.Printf("Error in %s.%d: %v (value=%q)\n",
            fieldErr.Segment, fieldErr.Field, fieldErr.Err, fieldErr.Value)
    }
}
```

Schema errors include the JSON path to the problematic definition:

```go
schema, err := hl7.ParseSchema(schemaJSON)
if err != nil {
    var schemaErr *hl7.SchemaError
    if errors.As(err, &schemaErr) {
        fmt.Printf("Schema error at %s: %v\n", schemaErr.Path, schemaErr.Err)
    }
}
```

## Examples

Complete runnable examples are available in the [`examples/`](./examples) directory:

- [`struct-based`](./examples/struct-based) - Traditional struct tag approach
- [`schema-based`](./examples/schema-based) - Dynamic JSON schema parsing
- [`generic`](./examples/generic) - Schema-less parsing
- [`hl7-to-json`](./examples/hl7-to-json) - HL7/JSON conversion pipeline

Run any example with:

```bash
go run ./examples/struct-based
go run ./examples/schema-based
go run ./examples/generic
go run ./examples/hl7-to-json
```

## Benchmarks

```
goos: darwin
goarch: arm64
pkg: github.com/esequiel378/hl7
cpu: Apple M1 Max
BenchmarkUnmarshalSimple-10                        17454             71802 ns/op         1049386 B/op         21 allocs/op
BenchmarkUnmarshalMultiSegment-10                  14560             81404 ns/op         1051180 B/op         48 allocs/op
BenchmarkUnmarshalWithRepetitions-10               15181             76750 ns/op         1050051 B/op         28 allocs/op
BenchmarkMarshalSimple-10                         514537              2224 ns/op            3032 B/op         14 allocs/op
BenchmarkMarshalMultiSegment-10                   205815              5777 ns/op            7112 B/op         34 allocs/op
BenchmarkRoundTrip-10                              14221             83589 ns/op         1058869 B/op         83 allocs/op
```

## Contributing

We welcome contributions to improve the library and the healthcare ecosystem. Open issues, suggest features, or contribute code to make healthcare development in Go even better.

## License

This library is distributed under the MIT License. See [LICENSE](./LICENSE) for more information.
