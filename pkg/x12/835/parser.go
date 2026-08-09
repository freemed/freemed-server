// Package x12_835 provides an X12 835 (ERA - Electronic Remittance Advice) parser.
package x12_835

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// ---------- Types ----------

// ERA835 represents a parsed 835 remission advice document.
type ERA835 struct {
	PayerName           string
	PayerID             string
	PaymentDate         string // YYYYMMDD
	PaymentAmount       float64
	PaymentMethod       string // CHK, ACH, NON
	TraceNumber         string // Check/EFT trace
	Claims              []ERA835Claim
	ProviderAdjustments []ERA835PLB
}

// ERA835Claim represents a single claim within an 835.
type ERA835Claim struct {
	PatientControlNumber string         // CLP01 - claim ID (matches procvoucher)
	ClaimStatusCode      string         // CLP02 - 1=processed as primary, 2=secondary, etc.
	ClaimCharge          float64        // CLP03
	ClaimPayment         float64        // CLP04
	PatientResponsibility float64        // CLP05
	PayerClaimID         string         // CLP07
	ServiceLines         []ERA835ServiceLine
	CASAdjustments       []ERA835CAS // Claim-level adjustments
}

// ERA835ServiceLine represents a service line within a claim.
type ERA835ServiceLine struct {
	ProcedureCode     string // SVC01-2 (HC:CPT)
	ProcedureModifier1 string
	ProcedureModifier2 string
	ProcedureModifier3 string
	ProcedureModifier4 string
	LineCharge        float64 // SVC02
	LinePayment       float64 // SVC03
	CASAdjustments    []ERA835CAS
}

// ERA835CAS represents a Claim Adjustment Segment.
type ERA835CAS struct {
	GroupCode  string // CO, PR, OA, PI
	ReasonCode string
	Amount     float64
	Quantity   int
}

// ERA835PLB represents a Provider-Level Balance adjustment.
type ERA835PLB struct {
	ProviderID       string
	AdjustmentDate   string
	AdjustmentReason string
	Amount           float64
}

// ---------- Segment helpers ----------

const (
	segmentTerminator = '~'
	elementSeparator  = '*'
)

// split835Segments splits raw 835 data into individual segment strings,
// stripping newlines and handling the segment terminator.
func split835Segments(data []byte) []string {
	// Strip newlines (EDI files may have CR/LF for readability)
	cleaned := bytes.ReplaceAll(data, []byte("\r\n"), nil)
	cleaned = bytes.ReplaceAll(cleaned, []byte("\r"), nil)
	cleaned = bytes.ReplaceAll(cleaned, []byte("\n"), nil)

	raw := string(cleaned)
	segs := strings.Split(raw, string(segmentTerminator))
	var result []string
	for _, s := range segs {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// segmentElements splits a segment string into its tag and elements.
func segmentElements(seg string) (tag string, elements []string) {
	parts := strings.Split(seg, string(elementSeparator))
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

// parseFloat safely converts a string to float64.
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// parseInt safely converts a string to int.
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

// ---------- Parser ----------

// Parse835 parses a raw X12 835 (ERA) byte stream into an ERA835 struct.
func Parse835(data []byte) (*ERA835, error) {
	segments := split835Segments(data)
	if len(segments) == 0 {
		return nil, fmt.Errorf("x12_835: empty file")
	}

	era := &ERA835{}
	var currentClaim *ERA835Claim
	var currentLine *ERA835ServiceLine
	inClaimLoop := false

	for _, seg := range segments {
		tag, elems := segmentElements(seg)

		switch tag {
		case "BPR":
			// BPR*I*AMOUNT*C*PAYMENT_METHOD**DATE*TRACE_NUMBER*PAYER_ID**PAYER_NAME~
			if len(elems) >= 2 {
				era.PaymentAmount = parseFloat(elems[1])
			}
			if len(elems) >= 4 {
				era.PaymentMethod = elems[3]
			}
			if len(elems) >= 6 {
				// BPR06 is the payment effective date (YYMMDD or CCYYMMDD)
				era.PaymentDate = normalize835Date(elems[5])
			}
			// BPR07 is trace number, BPR08 is payer ID, BPR10 is payer name
			// BPR format: BPR*I*AMT*C*METHOD**DATE*TRACE*PAYERID**PAYERNAME~
			if len(elems) >= 9 {
				era.PayerID = elems[7]
			}
			if len(elems) >= 10 {
				era.PayerName = elems[9]
			}
			// Trace number can also be in position 6/7 depending on format
			if len(elems) >= 7 && era.TraceNumber == "" {
				era.TraceNumber = elems[6]
			}

		case "TRN":
			// TRN*1*TRACE_NUMBER*PAYER_ID~
			if len(elems) >= 2 {
				era.TraceNumber = elems[1]
			}

		case "N1":
			// N1*PR*PAYER_NAME~
			if len(elems) >= 3 && elems[0] == "PR" {
				era.PayerName = elems[2]
			}

		case "REF":
			// REF*F8*REFERENCE_NUMBER - patient control number reference
			// REF*1L*PAYER_CLAIM_ID - payer claim ID
			if len(elems) >= 3 && currentClaim != nil {
				switch elems[0] {
				case "F8":
					// Original reference number - leave as-is, it's just additional info
				case "1L":
					if currentClaim.PayerClaimID == "" {
						currentClaim.PayerClaimID = elems[1]
					}
				}
			}

		case "DTM":
			// DTM*050*DATE - received date
			// DTM*232*DATE - claim statement period start
			// DTM*233*DATE - claim statement period end
			if len(elems) >= 2 && elems[0] == "050" && era.PaymentDate == "" {
				era.PaymentDate = normalize835Date(elems[1])
			}

		case "CLP":
			// CLP*PATIENT_CONTROL*CLAIM_STATUS*CHARGE*PAYMENT*PATIENT_RESP**PAYER_CLAIM_ID~
			// Finalize any previous claim (flush current line first) and start a new one
			if currentClaim != nil {
				if currentLine != nil {
					currentClaim.ServiceLines = append(currentClaim.ServiceLines, *currentLine)
					currentLine = nil
				}
				era.Claims = append(era.Claims, *currentClaim)
			}
			currentClaim = &ERA835Claim{}
			currentLine = nil
			inClaimLoop = true

			if len(elems) >= 1 {
				currentClaim.PatientControlNumber = elems[0]
			}
			if len(elems) >= 2 {
				currentClaim.ClaimStatusCode = elems[1]
			}
			if len(elems) >= 3 {
				currentClaim.ClaimCharge = parseFloat(elems[2])
			}
			if len(elems) >= 4 {
				currentClaim.ClaimPayment = parseFloat(elems[3])
			}
			if len(elems) >= 5 {
				currentClaim.PatientResponsibility = parseFloat(elems[4])
			}
			if len(elems) >= 7 {
				currentClaim.PayerClaimID = elems[6]
			}

		case "CAS":
			// CAS*GROUP*REASON*AMOUNT*QUANTITY~
			// CAS can repeat - multiple adjustments per level
			cas := parseCAS(elems)
			if cas != nil {
				if currentLine != nil {
					currentLine.CASAdjustments = append(currentLine.CASAdjustments, *cas)
				} else if currentClaim != nil {
					currentClaim.CASAdjustments = append(currentClaim.CASAdjustments, *cas)
				}
			}

		case "SVC":
			// SVC*HC:CPT*LINE_CHARGE*LINE_PAYMENT~
			// Finalize any previous service line
			if currentClaim != nil {
				if currentLine != nil {
					currentClaim.ServiceLines = append(currentClaim.ServiceLines, *currentLine)
				}
				currentLine = &ERA835ServiceLine{}
			}

			if len(elems) >= 1 && currentLine != nil {
				// SVC01-1 through SVC01-6 are composite: HC:CPT:MOD1:MOD2:MOD3:MOD4
				svcComposite := strings.Split(elems[0], ":")
				if len(svcComposite) > 1 {
					currentLine.ProcedureCode = svcComposite[1]
				}
				if len(svcComposite) > 2 {
					currentLine.ProcedureModifier1 = svcComposite[2]
				}
				if len(svcComposite) > 3 {
					currentLine.ProcedureModifier2 = svcComposite[3]
				}
				if len(svcComposite) > 4 {
					currentLine.ProcedureModifier3 = svcComposite[4]
				}
				if len(svcComposite) > 5 {
					currentLine.ProcedureModifier4 = svcComposite[5]
				}
			}
			if len(elems) >= 2 && currentLine != nil {
				currentLine.LineCharge = parseFloat(elems[1])
			}
			if len(elems) >= 3 && currentLine != nil {
				currentLine.LinePayment = parseFloat(elems[2])
			}

		case "PLB":
			// PLB*PROVIDER_ID*YYYYMMDD*ADJUSTMENT_REASON*AMOUNT~
			plb := ERA835PLB{}
			if len(elems) >= 1 {
				plb.ProviderID = elems[0]
			}
			if len(elems) >= 2 {
				plb.AdjustmentDate = normalize835Date(elems[1])
			}
			if len(elems) >= 3 {
				var reasonParts []string
				var amtParts []string
				for i := 2; i < len(elems); i++ {
					// PLB adjustments come in pairs: reason, amount
					// Or sometimes just reason if there's a single adjustment
					if i > 2 && i < len(elems) && isNumeric(elems[i]) {
						// Check if this looks like an amount
						_, err := strconv.ParseFloat(elems[i], 64)
						if err == nil && len(reasonParts) > 0 {
							amtParts = append(amtParts, elems[i])
							continue
						}
					}
					reasonParts = append(reasonParts, elems[i])
				}
				plb.AdjustmentReason = strings.Join(reasonParts, ":")
				if len(amtParts) > 0 {
					plb.Amount = parseFloat(amtParts[0])
				}
			}
			era.ProviderAdjustments = append(era.ProviderAdjustments, plb)

		case "SE", "GE", "IEA":
			// Trailer segments — finalize any open claim
			if currentClaim != nil {
				if currentLine != nil {
					currentClaim.ServiceLines = append(currentClaim.ServiceLines, *currentLine)
					currentLine = nil
				}
				era.Claims = append(era.Claims, *currentClaim)
				currentClaim = nil
			}
			inClaimLoop = false
		}
	}

	// Finalize any remaining open claim
	if currentClaim != nil {
		if currentLine != nil {
			currentClaim.ServiceLines = append(currentClaim.ServiceLines, *currentLine)
		}
		era.Claims = append(era.Claims, *currentClaim)
	}

	if len(era.Claims) == 0 && !inClaimLoop {
		return era, nil
	}

	return era, nil
}

// parseCAS parses a CAS segment elements into an ERA835CAS.
func parseCAS(elems []string) *ERA835CAS {
	if len(elems) < 3 {
		return nil
	}
	cas := &ERA835CAS{
		GroupCode:  elems[0],
		ReasonCode: elems[1],
		Amount:     parseFloat(elems[2]),
	}
	if len(elems) >= 4 {
		cas.Quantity = parseInt(elems[3])
	}
	return cas
}

// normalize835Date converts various date formats found in 835s to CCYYMMDD.
// Accepts CCYYMMDD (8 digits) or YYMMDD (6 digits).
func normalize835Date(s string) string {
	s = strings.TrimSpace(s)
	if len(s) == 8 {
		return s
	}
	if len(s) == 6 {
		// Assume 20xx for 6-digit dates
		return "20" + s
	}
	return s
}

// isNumeric checks if a string looks like a number (all digits, optional sign and decimal).
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	// Allow negative sign at start
	start := 0
	if s[0] == '-' {
		start = 1
	}
	hasDecimal := false
	for i := start; i < len(s); i++ {
		if s[i] == '.' {
			if hasDecimal {
				return false
			}
			hasDecimal = true
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
