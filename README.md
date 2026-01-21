# hl7 - A Golang Library for Parsing and Building HL7 Messages

[![Go Reference](https://pkg.go.dev/badge/github.com/esequiel378/hl7.svg)](https://pkg.go.dev/github.com/esequiel378/hl7)
[![Go Report Card](https://goreportcard.com/badge/github.com/esequiel378/hl7)](https://goreportcard.com/report/github.com/esequiel378/hl7)

A lightweight, dependency-free Go library for parsing and building HL7 v2.x messages. It provides a simple and powerful way to handle one of the most common file formats in healthcare, leveraging Go's robust type system and error handling.

## Features

- **Bidirectional Conversion**: Both `Unmarshal` (parse) and `Marshal` (build) HL7 messages
- **Struct Tag Parsing**: Define HL7 mappings with intuitive struct tags (`hl7:"segment:<name>"` and `hl7:"<index>"`)
- **Nested Structs**: Easily manage complex fields like patient names or addresses using component separators (^)
- **Repetition Support**: Parse repeating fields (~) into Go slices
- **Timestamp Type**: Built-in `hl7.Timestamp` type for automatic date/time parsing
- **Custom Types**: Implement `Unmarshaler` or `Marshaler` interfaces for custom field handling
- **Version Agnostic**: Supports parsing any HL7 v2.x version
- **Rich Errors**: Field-level error context for easier debugging
- **Dependency-Free**: No external dependencies—ready to go out of the box

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
