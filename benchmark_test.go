package hl7_test

import (
	"testing"

	"github.com/esequiel378/hl7"
)

// Sample HL7 messages for benchmarking
var (
	simpleMSH = []byte("MSH|^~\\&|App1|Fac1|App2|Fac2|20250205120000||ADT^A01|1234|P|2.3")

	multiSegmentMessage = []byte(`MSH|^~\&|HIS|RIH|EKG|EKG|200101011230||ADT^A01|MSG00001|P|2.3
PID|1||123456^^^HOSP^MR||Doe^John^A||19700101|M||2106-3|123 Main St^^Metropolis^NY^10101||555-1234|||S||123456789|987-65-4321
PV1|1|I|2000^2012^01||||1234^Jones^Bob|||SUR|||||1234567|A0|`)

	messageWithRepetitions = []byte(`MSH|^~\&|App|Fac|||20250205120000||ADT^A01|123|P|2.3
PID|1||ID001~ID002~ID003~ID004~ID005|||||||||||||||`)
)

// Benchmark structs
type BenchMessageType struct {
	MessageCode  string `hl7:"1"`
	TriggerEvent string `hl7:"2"`
}

type BenchMSHSegment struct {
	FieldSeparator       string           `hl7:"1"`
	EncodingCharacters   string           `hl7:"2"`
	SendingApplication   string           `hl7:"3"`
	SendingFacility      string           `hl7:"4"`
	ReceivingApplication string           `hl7:"5"`
	ReceivingFacility    string           `hl7:"6"`
	DateTimeOfMessage    string           `hl7:"7"`
	Security             string           `hl7:"8"`
	MessageType          BenchMessageType `hl7:"9"`
	MessageControlID     string           `hl7:"10"`
	ProcessingID         string           `hl7:"11"`
	VersionID            string           `hl7:"12"`
}

type BenchPatientName struct {
	FamilyName string `hl7:"1"`
	GivenName  string `hl7:"2"`
	MiddleName string `hl7:"3"`
}

type BenchPIDSegment struct {
	SetID                 string           `hl7:"1"`
	PatientIdentifierList string           `hl7:"3"`
	PatientName           BenchPatientName `hl7:"5"`
	DateOfBirth           string           `hl7:"7"`
	Gender                string           `hl7:"8"`
	Race                  string           `hl7:"10"`
	Address               string           `hl7:"11"`
	PhoneNumber           string           `hl7:"13"`
	MaritalStatus         string           `hl7:"16"`
	AccountNumber         string           `hl7:"18"`
	SocialSecurityNumber  string           `hl7:"19"`
}

type BenchPV1Segment struct {
	SetID                   string `hl7:"1"`
	PatientClass            string `hl7:"2"`
	AssignedPatientLocation string `hl7:"3"`
	AttendingDoctor         string `hl7:"7"`
	HospitalService         string `hl7:"10"`
	VisitNumber             string `hl7:"15"`
	FinancialClass          string `hl7:"16"`
}

type BenchSimpleMessage struct {
	MSH BenchMSHSegment `hl7:"segment:MSH"`
}

type BenchMultiMessage struct {
	MSH BenchMSHSegment `hl7:"segment:MSH"`
	PID BenchPIDSegment `hl7:"segment:PID"`
	PV1 BenchPV1Segment `hl7:"segment:PV1"`
}

type BenchRepetitionPID struct {
	SetID         string   `hl7:"1"`
	PatientIDList []string `hl7:"3"`
}

type BenchRepetitionMessage struct {
	MSH BenchMSHSegment    `hl7:"segment:MSH"`
	PID BenchRepetitionPID `hl7:"segment:PID"`
}

func BenchmarkUnmarshalSimple(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var msg BenchSimpleMessage
		if err := hl7.Unmarshal(simpleMSH, &msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalMultiSegment(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var msg BenchMultiMessage
		if err := hl7.Unmarshal(multiSegmentMessage, &msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalWithRepetitions(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var msg BenchRepetitionMessage
		if err := hl7.Unmarshal(messageWithRepetitions, &msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalSimple(b *testing.B) {
	msg := BenchSimpleMessage{
		MSH: BenchMSHSegment{
			FieldSeparator:       "|",
			EncodingCharacters:   "^~\\&",
			SendingApplication:   "App1",
			SendingFacility:      "Fac1",
			ReceivingApplication: "App2",
			ReceivingFacility:    "Fac2",
			DateTimeOfMessage:    "20250205120000",
			MessageType:          BenchMessageType{MessageCode: "ADT", TriggerEvent: "A01"},
			MessageControlID:     "1234",
			ProcessingID:         "P",
			VersionID:            "2.3",
		},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := hl7.Marshal(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshalMultiSegment(b *testing.B) {
	msg := BenchMultiMessage{
		MSH: BenchMSHSegment{
			FieldSeparator:       "|",
			EncodingCharacters:   "^~\\&",
			SendingApplication:   "HIS",
			SendingFacility:      "RIH",
			ReceivingApplication: "EKG",
			ReceivingFacility:    "EKG",
			DateTimeOfMessage:    "200101011230",
			MessageType:          BenchMessageType{MessageCode: "ADT", TriggerEvent: "A01"},
			MessageControlID:     "MSG00001",
			ProcessingID:         "P",
			VersionID:            "2.3",
		},
		PID: BenchPIDSegment{
			SetID:                 "1",
			PatientIdentifierList: "123456^^^HOSP^MR",
			PatientName:           BenchPatientName{FamilyName: "Doe", GivenName: "John", MiddleName: "A"},
			DateOfBirth:           "19700101",
			Gender:                "M",
			Race:                  "2106-3",
			Address:               "123 Main St^^Metropolis^NY^10101",
			PhoneNumber:           "555-1234",
			MaritalStatus:         "S",
			AccountNumber:         "123456789",
			SocialSecurityNumber:  "987-65-4321",
		},
		PV1: BenchPV1Segment{
			SetID:                   "1",
			PatientClass:            "I",
			AssignedPatientLocation: "2000^2012^01",
			AttendingDoctor:         "1234^Jones^Bob",
			HospitalService:         "SUR",
			VisitNumber:             "1234567",
			FinancialClass:          "A0",
		},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := hl7.Marshal(msg); err != nil {
			b.Fatal(err)
		}
	}
}

// Schema-based benchmarks
var benchSchema *hl7.MessageSchema

func init() {
	var err error
	benchSchema, err = hl7.ParseSchema([]byte(`{
		"segments": {
			"MSH": {
				"fields": {
					"1": { "name": "fieldSeparator", "type": "string" },
					"2": { "name": "encodingCharacters", "type": "string" },
					"3": { "name": "sendingApplication", "type": "string" },
					"4": { "name": "sendingFacility", "type": "string" },
					"5": { "name": "receivingApplication", "type": "string" },
					"6": { "name": "receivingFacility", "type": "string" },
					"7": { "name": "dateTimeOfMessage", "type": "string" },
					"9": {
						"name": "messageType", "type": "object",
						"components": {
							"1": { "name": "messageCode", "type": "string" },
							"2": { "name": "triggerEvent", "type": "string" }
						}
					},
					"10": { "name": "messageControlID", "type": "string" },
					"11": { "name": "processingID", "type": "string" },
					"12": { "name": "versionID", "type": "string" }
				}
			},
			"PID": {
				"fields": {
					"1": { "name": "setID", "type": "string" },
					"3": { "name": "patientIdentifierList", "type": "string" },
					"5": { "name": "patientName", "type": "string" },
					"7": { "name": "dateOfBirth", "type": "string" },
					"8": { "name": "gender", "type": "string" }
				}
			},
			"PV1": {
				"fields": {
					"1": { "name": "setID", "type": "string" },
					"2": { "name": "patientClass", "type": "string" },
					"3": { "name": "assignedPatientLocation", "type": "string" },
					"10": { "name": "hospitalService", "type": "string" }
				}
			}
		}
	}`))
	if err != nil {
		panic(err)
	}
}

func BenchmarkSchemaUnmarshalSimple(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := hl7.UnmarshalWithSchema(simpleMSH, benchSchema); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSchemaUnmarshalMultiSegment(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := hl7.UnmarshalWithSchema(multiSegmentMessage, benchSchema); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSchemaMarshalSimple(b *testing.B) {
	data := map[string]any{
		"MSH": map[string]any{
			"fieldSeparator":       "|",
			"encodingCharacters":   "^~\\&",
			"sendingApplication":   "App1",
			"sendingFacility":      "Fac1",
			"receivingApplication": "App2",
			"receivingFacility":    "Fac2",
			"dateTimeOfMessage":    "20250205120000",
			"messageType": map[string]any{
				"messageCode":  "ADT",
				"triggerEvent": "A01",
			},
			"messageControlID": "1234",
			"processingID":     "P",
			"versionID":        "2.3",
		},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := hl7.MarshalWithSchema(data, benchSchema); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSchemaRoundTrip(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		result, err := hl7.UnmarshalWithSchema(multiSegmentMessage, benchSchema)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := hl7.MarshalWithSchema(result, benchSchema); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRoundTrip(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var msg BenchMultiMessage
		if err := hl7.Unmarshal(multiSegmentMessage, &msg); err != nil {
			b.Fatal(err)
		}
		if _, err := hl7.Marshal(msg); err != nil {
			b.Fatal(err)
		}
	}
}
