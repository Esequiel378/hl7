package hl7_test

import (
	"testing"
	"time"

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

func TestTimeField(t *testing.T) {
	// Raw HL7 message with a timestamp in the "YYYYMMDDHHMM" format.
	raw := "MSH|^~\\&||.|||1999-05-29T23:00:00Z||ADT^A04|ADT.1.1698593|P|2.7"

	var st struct {
		Header struct {
			DateTimeOfMessage time.Time `hl7:"6"`
		} `hl7:"segment:MSH"`
	}

	if err := hl7.Unmarshal([]byte(raw), &st); err != nil {
		t.Fatal(err)
	}

	expectedTime, err := time.Parse("200601021504", "199905292300")
	if err != nil {
		t.Fatalf("Failed to parse expected time: %v", err)
	}

	if !st.Header.DateTimeOfMessage.Equal(expectedTime) {
		t.Fatalf("DateTimeOfMessage mismatch: got %v, want %v",
			st.Header.DateTimeOfMessage, expectedTime)
	}
}

func TestEmptyTimeField(t *testing.T) {
	// Raw HL7 message with a timestamp in the "YYYYMMDDHHMM" format.
	raw := "MSH|^~\\&||.|||||ADT^A04|ADT.1.1698593|P|2.7"

	var st struct {
		Header struct {
			DateTimeOfMessage *time.Time `hl7:"6"`
		} `hl7:"segment:MSH"`
	}

	if err := hl7.Unmarshal([]byte(raw), &st); err != nil {
		t.Fatal(err)
	}

	if st.Header.DateTimeOfMessage != nil {
		t.Fatalf("DateTimeOfMessage mismatch: got %v, want nil", st.Header.DateTimeOfMessage)
	}
}

func TestEmptyField(t *testing.T) {
	raw := "MSH|^~\\&||.|||199908180016||ADT^A04|ADT.1.1698593|P|2.7"

	var st struct {
		Header struct {
			EncodingCharacters *string `hl7:"2"`
		} `hl7:"segment:MSH"`
	}

	if err := hl7.Unmarshal([]byte(raw), &st); err != nil {
		t.Fatal(err)
	}

	if st.Header.EncodingCharacters != nil {
		t.Fatalf("expected nil got %s", *st.Header.EncodingCharacters)
	}

	raw = "MSH|^~\\&|*|.|||199908180016||ADT^A04|ADT.1.1698593|P|2.7"

	if err := hl7.Unmarshal([]byte(raw), &st); err != nil {
		t.Fatal(err)
	}

	if st.Header.EncodingCharacters == nil {
		t.Fatalf("expected * got %s", *st.Header.EncodingCharacters)
	}
}
