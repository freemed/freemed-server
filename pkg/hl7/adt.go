package hl7

import (
	"fmt"
	"strings"
	"time"
)

// Supported ADT event types.
const (
	ADT_A01 = "A01" // Patient Admit
	ADT_A03 = "A03" // Patient Discharge
	ADT_A04 = "A04" // Patient Registration
	ADT_A08 = "A08" // Patient Update
)

// PatientDemographics holds the patient information extracted from an ADT message.
type PatientDemographics struct {
	MRN       string // PID-3.1 — Medical Record Number
	LastName  string // PID-5.1
	FirstName string // PID-5.2
	DOB       string // PID-7 (YYYYMMDD)
	Sex       string // PID-8 (M/F/O/U)
	Address1  string // PID-11.1
	City      string // PID-11.3
	State     string // PID-11.4
	Zip       string // PID-11.5
	Phone     string // PID-13
	SSN       string // PID-19
}

// ParseADT extracts patient demographics and event type from an ADT message.
// Returns demographics, event type code (A01/A03/A04/A08), and any error.
func ParseADT(msg *Message) (*PatientDemographics, string, error) {
	if msg == nil {
		return nil, "", fmt.Errorf("hl7: nil message")
	}

	msh := msg.FindSegment("MSH")
	if msh == nil {
		return nil, "", fmt.Errorf("hl7: missing MSH segment")
	}

	// MSH-9 contains message type, e.g. "ADT^A04"
	msgType := msh.GetComponent(9, 1)
	if !strings.EqualFold(msgType, "ADT") {
		return nil, "", fmt.Errorf("hl7: not an ADT message, got %q", msh.GetField(9))
	}

	eventType := msh.GetComponent(9, 2)

	pid := msg.FindSegment("PID")
	if pid == nil {
		return nil, eventType, fmt.Errorf("hl7: missing PID segment")
	}

	pt := &PatientDemographics{
		MRN:       pid.GetComponent(3, 1),
		LastName:  pid.GetComponent(5, 1),
		FirstName: pid.GetComponent(5, 2),
		DOB:       pid.GetField(7),
		Sex:       pid.GetField(8),
		Address1:  pid.GetComponent(11, 1),
		City:      pid.GetComponent(11, 3),
		State:     pid.GetComponent(11, 4),
		Zip:       pid.GetComponent(11, 5),
		Phone:     pid.GetField(13),
		SSN:       pid.GetField(19),
	}

	return pt, eventType, nil
}

// BuildADT_A04 builds an ADT^A04 (patient registration) message from demographics.
func BuildADT_A04(pt *PatientDemographics) (*Message, error) {
	if pt == nil {
		return nil, fmt.Errorf("hl7: nil patient demographics")
	}

	now := time.Now().Format("20060102150405")
	controlID := fmt.Sprintf("FREEMED_%s", now)

	msg := &Message{}

	// MSH segment
	// MSH|^~\&|FREEMED|FREEMED_FACILITY|DEST_APP|DEST_FACILITY|datetime||ADT^A04|controlID|P|2.5
	msh := Segment{
		Name: "MSH",
		Fields: [][]string{
			{"|"},                // MSH-1: field separator (always |)
			{string(ComponentSep) + string(RepetitionSep) + string(EscapeChar) + "&"}, // MSH-2: encoding chars
			{"FREEMED"},          // MSH-3: sending application
			{"FREEMED_FACILITY"}, // MSH-4: sending facility
			{"DEST_APP"},         // MSH-5: receiving application
			{"DEST_FACILITY"},    // MSH-6: receiving facility
			{now},                // MSH-7: date/time of message
			{""},                 // MSH-8: security
			{"ADT", "A04"},       // MSH-9: message type
			{controlID},          // MSH-10: message control ID
			{"P"},                // MSH-11: processing ID
			{"2.5"},              // MSH-12: version ID
		},
	}
	msg.Segments = append(msg.Segments, msh)

	// PID segment
	pid := Segment{
		Name: "PID",
		Fields: [][]string{
			{"1"},                                                    // PID-1: set ID
			{""},                                                     // PID-2: patient ID (external)
			{pt.MRN, "", "", "FREEMED", "MR"},                        // PID-3: patient identifier list
			{""},                                                     // PID-4: alternate ID
			{pt.LastName, pt.FirstName},                              // PID-5: patient name
			{""},                                                     // PID-6: mother's maiden name
			{pt.DOB},                                                 // PID-7: DOB
			{pt.Sex},                                                 // PID-8: sex
			{""},                                                     // PID-9: alias
			{""},                                                     // PID-10: race
			{pt.Address1, "", pt.City, pt.State, pt.Zip},            // PID-11: address
			{""},                                                    // PID-12: county
			{pt.Phone},                                              // PID-13: phone
			{""},                                                    // PID-14: business phone
			{""},                                                    // PID-15: primary language
			{""},                                                    // PID-16: marital status
			{""},                                                    // PID-17: religion
			{""},                                                    // PID-18: account number
			{pt.SSN},                                                // PID-19: SSN
		},
	}
	msg.Segments = append(msg.Segments, pid)

	return msg, nil
}
