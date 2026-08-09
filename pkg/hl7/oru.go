package hl7

import (
	"fmt"
	"strings"
)

// ORU_R01 is the message type code for observation result (lab results).
const ORU_R01 = "R01"

// LabResult holds the parsed contents of an ORU^R01 lab results message.
type LabResult struct {
	PatientMRN   string          // PID-3.1
	OrderNumber  string          // OBR-2 (placer) or OBR-3 (filler)
	TestCode     string          // OBR-4.1 (LOINC code)
	TestName     string          // OBR-4.2
	ResultDate   string          // OBR-7 (YYYYMMDDHHMMSS)
	Observations []LabObservation
}

// LabObservation holds a single OBX segment from an ORU message.
type LabObservation struct {
	OBXSetID     string // OBX-1
	ValueType    string // OBX-2 (NM=numeric, ST=string, CE=coded)
	TestCode     string // OBX-3.1 (LOINC)
	TestName     string // OBX-3.2
	Value        string // OBX-5
	Units        string // OBX-6
	RefRange     string // OBX-7
	AbnormalFlag string // OBX-8 (H=high, L=low, A=abnormal, N=normal)
	ResultStatus string // OBX-11 (F=final, P=preliminary)
}

// ParseORU extracts lab results from an ORU^R01 message.
func ParseORU(msg *Message) (*LabResult, error) {
	if msg == nil {
		return nil, fmt.Errorf("hl7: nil message")
	}

	msh := msg.FindSegment("MSH")
	if msh == nil {
		return nil, fmt.Errorf("hl7: missing MSH segment")
	}

	// MSH-9.1 should be "ORU" and MSH-9.2 should be "R01"
	msgType := msh.GetComponent(9, 1)
	if !strings.EqualFold(msgType, "ORU") {
		return nil, fmt.Errorf("hl7: not an ORU message, got %q", msh.GetField(9))
	}

	result := &LabResult{}

	// Extract patient MRN from PID-3.1
	if pid := msg.FindSegment("PID"); pid != nil {
		result.PatientMRN = pid.GetComponent(3, 1)
	}

	// Extract OBR information
	obr := msg.FindSegment("OBR")
	if obr == nil {
		return nil, fmt.Errorf("hl7: missing OBR segment")
	}

	result.OrderNumber = obr.GetField(2)
	if result.OrderNumber == "" {
		result.OrderNumber = obr.GetField(3) // fallback to filler order number
	}
	result.TestCode = obr.GetComponent(4, 1)
	result.TestName = obr.GetComponent(4, 2)
	result.ResultDate = obr.GetField(7)

	// Extract all OBX segments
	for _, seg := range msg.Segments {
		if !strings.EqualFold(seg.Name, "OBX") {
			continue
		}
		obs := LabObservation{
			OBXSetID:     seg.GetField(1),
			ValueType:    seg.GetField(2),
			TestCode:     seg.GetComponent(3, 1),
			TestName:     seg.GetComponent(3, 2),
			Value:        seg.GetField(5),
			Units:        seg.GetField(6),
			RefRange:     seg.GetField(7),
			AbnormalFlag: seg.GetField(8),
			ResultStatus: seg.GetField(11),
		}
		result.Observations = append(result.Observations, obs)
	}

	if len(result.Observations) == 0 {
		return nil, fmt.Errorf("hl7: no OBX segments found")
	}

	return result, nil
}
