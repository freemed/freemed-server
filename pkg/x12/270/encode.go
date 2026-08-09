package x12_270

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

const (
	segmentTerminator = "~"
	elementSeparator  = "*"
	subElementSep     = ":"
)

// =============================================================================
// X12 270 Encoder
// =============================================================================

// Encode270 generates a complete X12 270 eligibility inquiry transaction set
// from the given EligibilityInquiry270 struct.
func Encode270(inq *EligibilityInquiry270) ([]byte, error) {
	var buf bytes.Buffer

	now := time.Now()
	dateStr := now.Format("060102") // YYMMDD
	timeStr := now.Format("1504")   // HHMM

	// ISA — Interchange Control Header
	isaSender := padRight(inq.SenderID, 15)
	isaReceiver := padRight(inq.ReceiverID, 15)
	buf.WriteString(fmt.Sprintf("ISA*00*          *00*          *ZZ*%s*ZZ*%s*%s*%s*U*00401*000000001*0*P*:%s",
		isaSender, isaReceiver, dateStr, timeStr, segmentTerminator))

	// GS — Functional Group Header
	buf.WriteString(fmt.Sprintf("GS*HS*%s*%s*%s*%s*1*X*004010X092A1%s",
		inq.SenderID, inq.ReceiverID, inq.TransactionDate, timeStr, segmentTerminator))

	// ST — Transaction Set Header
	buf.WriteString(fmt.Sprintf("ST*270*%s%s", "0001", segmentTerminator))

	// BHT — Beginning of Hierarchical Transaction
	buf.WriteString(fmt.Sprintf("BHT*0022*13*%s*%s*%s%s",
		inq.TransactionNo, inq.TransactionDate, timeStr, segmentTerminator))

	// =========================================================================
	// Loop 2000A — Information Source (Payer)
	// =========================================================================
	buf.WriteString(fmt.Sprintf("HL*1**20*1%s", segmentTerminator)) // HL01=1, HL03=20(Information Source), HL04=1

	// NM1 — Information Source Name
	src := inq.InfoSource
	buf.WriteString(fmt.Sprintf("NM1*PR*2*%s*****%s*%s%s",
		src.Name, src.EntityIDQual, src.EntityID, segmentTerminator))

	// =========================================================================
	// Loop 2000B — Information Receiver (Provider)
	// =========================================================================
	buf.WriteString(fmt.Sprintf("HL*2*1*21*1%s", segmentTerminator)) // HL04=1(additional subordinate data)

	// NM1 — Information Receiver Name
	rcv := inq.InfoReceiver
	buf.WriteString(fmt.Sprintf("NM1*1P*2*%s*****%s*%s%s",
		rcv.Name, rcv.EntityIDQual, rcv.EntityID, segmentTerminator))

	// REF — Receiver Reference (optional)
	if rcv.EntityIDQual == "XX" {
		buf.WriteString(fmt.Sprintf("REF*TJ*%s%s", rcv.EntityID, segmentTerminator)) // TJ = Federal Taxpayer ID
	}

	// =========================================================================
	// Loop 2000C — Subscriber
	// =========================================================================
	subLevel := "3"
	subParent := "2"
	if inq.Dependent != nil {
		buf.WriteString(fmt.Sprintf("HL*3*2*22*0%s", segmentTerminator)) // HL04=0 (has dependent)
		subLevel = "4"
		subParent = "3"
	} else {
		buf.WriteString(fmt.Sprintf("HL*3*2*22*1%s", segmentTerminator)) // HL04=1 (no dependent)
	}

	sub := inq.Subscriber

	// NM1 — Subscriber Name
	buf.WriteString(fmt.Sprintf("NM1*IL*1*%s*%s*%s***%s*%s%s",
		sub.LastName, sub.FirstName, sub.MiddleName, sub.IDQualifier, sub.MemberID, segmentTerminator))

	// REF — Subscriber Group/Policy Number
	if sub.GroupNumber != "" {
		buf.WriteString(fmt.Sprintf("REF*6P*%s%s", sub.GroupNumber, segmentTerminator))
	}

	// N3/N4 — Subscriber Address
	if sub.Address1 != "" {
		if sub.Address2 != "" {
			buf.WriteString(fmt.Sprintf("N3*%s*%s%s", sub.Address1, sub.Address2, segmentTerminator))
		} else {
			buf.WriteString(fmt.Sprintf("N3*%s%s", sub.Address1, segmentTerminator))
		}
		buf.WriteString(fmt.Sprintf("N4*%s*%s*%s%s", sub.City, sub.State, sub.Zip, segmentTerminator))
	}

	// DMG — Subscriber Demographic
	buf.WriteString(fmt.Sprintf("DMG*D8*%s*%s%s", sub.DOB, sub.Gender, segmentTerminator))

	// TRN — Trace Number (for matching response)
	traceNo := fmt.Sprintf("%d", now.Unix())
	buf.WriteString(fmt.Sprintf("TRN*1*%s%s", traceNo, segmentTerminator))

	// =========================================================================
	// Loop 2110C — Subscriber Eligibility Query
	// =========================================================================
	buf.WriteString(fmt.Sprintf("EQ*%s%s", strings.Join(inq.ServiceTypes, elementSeparator), segmentTerminator))

	// =========================================================================
	// Loop 2000D — Dependent (if applicable)
	// =========================================================================
	if inq.Dependent != nil {
		dep := inq.Dependent
		buf.WriteString(fmt.Sprintf("HL*%s*%s*23*0%s", subLevel, subParent, segmentTerminator))

		// NM1 — Dependent Name
		buf.WriteString(fmt.Sprintf("NM1*QC*1*%s*%s****%s%s",
			dep.LastName, dep.FirstName, sub.IDQualifier, segmentTerminator))

		// INS — Relationship
		buf.WriteString(fmt.Sprintf("INS*Y*18*%s***A%s", dep.Relationship, segmentTerminator))

		// DMG — Dependent Demographic
		buf.WriteString(fmt.Sprintf("DMG*D8*%s*%s%s", dep.DOB, dep.Gender, segmentTerminator))

		// TRN — Trace Number
		buf.WriteString(fmt.Sprintf("TRN*1*%s%s", traceNo, segmentTerminator))

		// Loop 2110D — Dependent Eligibility Query
		buf.WriteString(fmt.Sprintf("EQ*%s%s", strings.Join(inq.ServiceTypes, elementSeparator), segmentTerminator))
	}

	// =========================================================================
	// Transaction Set Trailer
	// =========================================================================
	hlCount := 3
	if inq.Dependent != nil {
		hlCount = 4
	}
	buf.WriteString(fmt.Sprintf("SE*%d*0001%s", hlCount+6, segmentTerminator)) // SE01 = segment count (approximate)

	// GE — Functional Group Trailer
	buf.WriteString(fmt.Sprintf("GE*1*1%s", segmentTerminator))

	// IEA — Interchange Control Trailer
	buf.WriteString(fmt.Sprintf("IEA*1*000000001%s", segmentTerminator))

	return buf.Bytes(), nil
}

// padRight pads a string to the given length with spaces.
func padRight(s string, length int) string {
	if len(s) >= length {
		return s[:length]
	}
	return s + strings.Repeat(" ", length-len(s))
}
