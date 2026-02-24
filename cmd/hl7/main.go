// Command hl7 parses HL7 v2.x messages and outputs JSON.
//
// Usage:
//
//	hl7 [flags] [file]
//
// Input is read from --file, the positional [file] argument, or stdin (in
// that order of precedence). Output is written to stdout.
//
// Flags:
//
//	-f, --file <file>     HL7 input file.
//	-s, --schema <file>   Path to a JSON schema file (query mode).
//	                      Without this flag the message is parsed generically.
//	-c, --compact         Emit compact JSON instead of pretty-printed output.
//	-h, --help            Show this help text.
//
// Examples:
//
//	# Generic parse from stdin
//	echo 'MSH|^~\&|App|Fac|||20250101||ADT^A01|1|P|2.7' | hl7
//
//	# Generic parse using --file flag
//	hl7 --file message.hl7
//
//	# Schema-based parse from a file
//	hl7 --schema adt_a01.json --file message.hl7
//
//	# Compact output
//	hl7 -c -s schema.json -f message.hl7
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/esequiel378/hl7"
)

func main() {
	fs := flag.NewFlagSet("hl7", flag.ContinueOnError)
	fs.Usage = usage(fs)

	var inputFile string
	fs.StringVar(&inputFile, "file", "", "HL7 input file")
	fs.StringVar(&inputFile, "f", "", "HL7 input file (shorthand)")

	var schemaFile string
	fs.StringVar(&schemaFile, "schema", "", "path to JSON schema file (query mode)")
	fs.StringVar(&schemaFile, "s", "", "path to JSON schema file (shorthand)")

	var compact bool
	fs.BoolVar(&compact, "compact", false, "emit compact JSON")
	fs.BoolVar(&compact, "c", false, "emit compact JSON (shorthand)")

	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}

	if inputFile == "" && len(fs.Args()) == 0 && isTerminal(os.Stdin) {
		fs.Usage()
		os.Exit(0)
	}

	input, err := readInput(inputFile, fs.Args())
	if err != nil {
		fatalf("error reading input: %v", err)
	}

	var result any

	if schemaFile != "" {
		schema, err := hl7.LoadSchemaFile(schemaFile)
		if err != nil {
			fatalf("error loading schema: %v", err)
		}
		result, err = hl7.UnmarshalWithSchema(input, schema)
		if err != nil {
			fatalf("error parsing HL7: %v", err)
		}
	} else {
		result, err = hl7.ParseGeneric(input)
		if err != nil {
			fatalf("error parsing HL7: %v", err)
		}
	}

	out, err := marshalJSON(result, compact)
	if err != nil {
		fatalf("error serializing JSON: %v", err)
	}

	os.Stdout.Write(out)
	os.Stdout.Write([]byte("\n"))
}

func readInput(file string, args []string) ([]byte, error) {
	if file != "" {
		return os.ReadFile(file)
	}
	if len(args) > 0 {
		return os.ReadFile(args[0])
	}
	return io.ReadAll(os.Stdin)
}

func marshalJSON(v any, compact bool) ([]byte, error) {
	if compact {
		return json.Marshal(v)
	}
	return json.MarshalIndent(v, "", "  ")
}

func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "hl7: "+format+"\n", args...)
	os.Exit(1)
}

func usage(fs *flag.FlagSet) func() {
	return func() {
		fmt.Fprintln(os.Stderr, `Usage: hl7 [flags] [file]

Parse an HL7 v2.x message and output JSON. Input is read from --file, the
positional [file] argument, or stdin (in that order of precedence).

Flags:
  -f, --file <file>     HL7 input file.
  -s, --schema <file>   JSON schema file for query mode (schema-based parsing).
                        Without this flag the message is parsed generically.
  -c, --compact         Emit compact JSON instead of pretty-printed output.
  -h, --help            Show this help text.

Examples:
  # Generic parse from stdin
  echo 'MSH|^~\&|App|Fac|||20250101||ADT^A01|1|P|2.7' | hl7

  # Generic parse using --file flag
  hl7 --file message.hl7

  # Schema-based parse
  hl7 --schema adt_a01.json --file message.hl7

  # Compact output
  hl7 -c -s schema.json -f message.hl7`)
		_ = fs
	}
}
