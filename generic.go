package hl7

import "strings"

// GenericMessage represents a fully parsed HL7 message without a predefined schema.
type GenericMessage struct {
	Segments []GenericSegment `json:"segments"`
}

// GenericSegment represents a single HL7 segment (e.g., MSH, PID, PV1).
type GenericSegment struct {
	Name   string         `json:"name"`
	Fields []GenericField `json:"fields"`
}

// GenericField represents a single field within a segment.
type GenericField struct {
	Name       string             `json:"name,omitempty"`
	Index      int                `json:"index"`
	Value      string             `json:"value"`
	Components []GenericComponent `json:"components,omitempty"`
	Repeats    []GenericRepeat    `json:"repeats,omitempty"`
}

// GenericRepeat represents a single repetition of a field.
type GenericRepeat struct {
	Value      string             `json:"value"`
	Components []GenericComponent `json:"components,omitempty"`
}

// GenericComponent represents a component within a field.
type GenericComponent struct {
	Index int    `json:"index"`
	Value string `json:"value"`
}

// ParseGeneric parses an HL7 message into a GenericMessage without requiring
// a predefined schema or struct. All fields, components, and repetitions are
// preserved in the output.
func ParseGeneric(data []byte) (*GenericMessage, error) {
	segments, err := parseMessage(data)
	if err != nil {
		return nil, err
	}

	msg := &GenericMessage{}

	for _, seg := range segments {
		gs := GenericSegment{Name: string(seg.name)}

		componentSep := "^"
		if len(seg.encodingCharacters) > 0 {
			componentSep = string(seg.encodingCharacters[0])
		}
		repetitionSep := ""
		if len(seg.encodingCharacters) > 1 {
			repetitionSep = string(seg.encodingCharacters[1])
		}

		if seg.name == "MSH" {
			// MSH-1: field separator
			gs.Fields = append(gs.Fields, GenericField{
				Index: 1,
				Value: seg.fieldSeparator,
			})
			// MSH-2 onward: parts[1] = encoding chars, parts[2] = MSH-3, etc.
			for i := 1; i < len(seg.fields); i++ {
				field := parseGenericField(i+1, seg.fields[i], componentSep, repetitionSep)
				gs.Fields = append(gs.Fields, field)
			}
		} else {
			// Non-MSH: parts[0] = segment name, parts[1] = field 1, etc.
			for i := 1; i < len(seg.fields); i++ {
				field := parseGenericField(i, seg.fields[i], componentSep, repetitionSep)
				gs.Fields = append(gs.Fields, field)
			}
		}

		msg.Segments = append(msg.Segments, gs)
	}

	return msg, nil
}

// parseGenericField parses a single field value, detecting components and repetitions.
func parseGenericField(index int, value, componentSep, repetitionSep string) GenericField {
	field := GenericField{
		Index: index,
		Value: value,
	}

	// Check for repetitions first
	if repetitionSep != "" && strings.Contains(value, repetitionSep) {
		reps := strings.Split(value, repetitionSep)
		for _, rep := range reps {
			gr := GenericRepeat{Value: rep}
			if strings.Contains(rep, componentSep) {
				gr.Components = parseComponents(rep, componentSep)
			}
			field.Repeats = append(field.Repeats, gr)
		}
		return field
	}

	// Check for components
	if strings.Contains(value, componentSep) {
		field.Components = parseComponents(value, componentSep)
	}

	return field
}

// parseComponents splits a value by the component separator and returns indexed components.
func parseComponents(value, componentSep string) []GenericComponent {
	parts := strings.Split(value, componentSep)
	components := make([]GenericComponent, len(parts))
	for i, part := range parts {
		components[i] = GenericComponent{
			Index: i + 1,
			Value: part,
		}
	}
	return components
}
