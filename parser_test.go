package hl7_test

import (
	"testing"

	"github.com/esequiel378/hl7"
)

func TestAll(t *testing.T) {
	raw := "MSH|^~\\&||.|||199908180016||ADT^A04|ADT.1.1698593|P|2.7"

	var st struct {
		Header struct {
			Segment              string `hl7:"0"`
			FieldSeparator       string `hl7:"1"`
			EncodingCharacters   string `hl7:"2"`
			SendingApplication   string `hl7:"3"`
			SendingFacility      string `hl7:"4"`
			ReceivingApplication string `hl7:"5"`
			ReceivingFacility    string `hl7:"6"`
			DateTimeOfMessage    string `hl7:"7"`
			Security             struct {
				Type  string `hl7:"0"`
				Value string `hl7:"1"`
			} `hl7:"8"`
			SecurityRaw      string `hl7:"8"`
			MessageType      string `hl7:"9"`
			MessageControlID string `hl7:"10"`
			ProcessingID     string `hl7:"11"`
			VersionID        string `hl7:"12"`
		} `hl7:"segment:MSH"`
	}

	if err := hl7.Unmarshal([]byte(raw), &st); err != nil {
		t.Fatal(err)
	}
}
