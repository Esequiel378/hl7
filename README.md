# hl7 - A Golang Library for Parsing HL7 Files

This `hl7` library is a lightweight and extensible Golang tool for parsing HL7 files into Go structs. It provides a simple and powerful way to handle one of the most common file formats in healthcare, leveraging Go's robust type system and error handling.

Whether you're building healthcare services or integrating HL7 messaging into your systems, `hl7` offers a foundation to simplify your work and enhance code reliability.

## Features

- Struct Tag Parsing: Define HL7 mappings with intuitive struct tags (hl7:"segment:<name>" and hl7:"<index>").
Nested Structs: Easily manage complex fields like patient names or addresses separated by the caret (^).
- Version Agnostic: Supports parsing any HL7 version, offering flexibility for custom and standard implementations.
- Error Handling: Leverages Go's built-in error system for detailed error reporting.
- Dependency-Free: No external dependencies—ready to go out of the box.

## Installation

Install the package using Go modules:

```bash
go get github.com/esequiel378/hl7
```

## Quick Start

Here's a basic example of using hl7 to parse a raw HL7 message into a Go struct:

```hl7
MSH|^~\&||.|||199908180016||ADT^A04|ADT.1.1698593|P|2.7
```

Corresponding Go Struct (1-based indices)

```go
package main

import (
	"fmt"
	"github.com/esequiel378/hl7"
)

func main() {
	raw := "MSH|^~\\&||.|||199908180016||ADT^A04|ADT.1.1698593|P|2.7"

	var message struct {
		Header struct {
			FieldSeparator     string `hl7:"1"`
			EncodingCharacters string `hl7:"2"`
			DateTimeOfMessage  string `hl7:"7"`
			MessageType        string `hl7:"9"`
			VersionID          string `hl7:"12"`
		} `hl7:"segment:MSH"`
	}

	if err := hl7.Unmarshal([]byte(raw), &message); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Parsed Message: %+v\n", message)
}
```

Output

```plaintext
Parsed Message: {Header:{FieldSeparator:| EncodingCharacters:^~\& DateTimeOfMessage:199908180016 MessageType:ADT^A04 VersionID:2.7}}
```

## Using hl7 with go-playground/validator

You can use hl7 together with [go-playground/validator](https://github.com/go-playground/validator) to validate parsed HL7 messages. This ensures that your data adheres to required constraints before further processing.

### Example: Parsing and Validating HL7 Messages

Here's an example demonstrating how to parse an HL7 message and validate its fields:

```go
package main

import (
	"fmt"
	"github.com/esequiel378/hl7"
)

func main() {
	raw := "MSH|^~\\&||.|||199908180016||ADT^A04|ADT.1.1698593|P|2.7"

	var message struct {
		Header struct {
			Segment            string `hl7:"0"`
			FieldSeparator     string `hl7:"1"`
			EncodingCharacters string `hl7:"2"`
			DateTimeOfMessage  string `hl7:"7"`
			MessageType        string `hl7:"9"`
			VersionID          string `hl7:"12"`
		} `hl7:"segment:MSH"`
	}

	if err := hl7.Unmarshal([]byte(raw), &message); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Parsed Message: %+v\n", message)
}
```

## Contribution

We welcome contributions to improve the library and the healthcare ecosystem. Open issues, suggest features, or contribute code to make healthcare development in Go even better.

## License

This library is distributed under the MIT License. See [LICENSE](./LICENSE) for more information.
