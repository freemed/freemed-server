package x12_270

import (
	"fmt"
	"strconv"
	"strings"
)

// =============================================================================
// X12 271 Parser
// =============================================================================

// Parse271 parses a raw X12 271 eligibility response into a structured object.
func Parse271(data []byte) (*EligibilityResponse271, error) {
	segments := splitSegments(data)

	resp := &EligibilityResponse271{}
	var (
		currentSubscriber  *EligibilitySubscriber
		currentDependent   *EligibilitySubscriber
		currentBenefit     *EligibilityBenefit
		inSubscriber       bool
		inDependent        bool
	)

	for _, seg := range segments {
		tag, elements := segmentElements(seg)

		switch tag {
		case "ISA":
			// Skip interchange header in response — handled by transport
		case "GS":
			// Skip functional group header
		case "ST":
			// Verify 271
			if len(elements) > 0 && elements[0] != "271" {
				// Not a 271 — may be wrapped, continue
			}
		case "BHT":
			if len(elements) > 1 {
				resp.TraceNumber = elements[1]
			}

		// Loop 2000A — Information Source (Payer)
		// NM101=entity type, NM103=last name, NM104=first name,
		// NM108=ID qualifier (second-to-last), NM109=ID code (last element)
		case "NM1":
			entityType := getElement(elements, 0)
			// Minimum: entityType + entityQual + lastName = 3 elements
			if len(elements) < 3 {
				continue
			}
					// NM109 is always the last element
					idCode := elements[len(elements)-1]
			switch entityType {
			case "PR": // Payer
				resp.PayerName = getElement(elements, 2)
				resp.PayerID = idCode
			case "1P": // Provider
				// Skip
			case "IL": // Subscriber
				currentSubscriber = &EligibilitySubscriber{
					LastName:  getElement(elements, 2),
					FirstName: getElement(elements, 3),
					MemberID:  idCode,
				}
				inSubscriber = true
				inDependent = false
			case "QC": // Dependent
				currentDependent = &EligibilitySubscriber{
					LastName:  getElement(elements, 2),
					FirstName: getElement(elements, 3),
					MemberID:  idCode,
				}
				inDependent = true
				inSubscriber = false
			}

		// HL — Hierarchy Level
		case "HL":
			if len(elements) >= 3 {
				hlCode := getElement(elements, 2)
				switch hlCode {
				case "22": // Subscriber
					inSubscriber = true
					inDependent = false
				case "23": // Dependent
					inDependent = true
					inSubscriber = false
				default:
					inSubscriber = false
					inDependent = false
				}
			}

		// DTP — Date/Time Period
		case "DTP":
			if len(elements) >= 3 {
				periodQual := getElement(elements, 0)
				periodDate := getElement(elements, 2)
				if inSubscriber && currentSubscriber != nil && currentSubscriber.EligibilityDate == "" {
					currentSubscriber.EligibilityDate = periodDate
				}
				if inDependent && currentDependent != nil && currentDependent.EligibilityDate == "" {
					currentDependent.EligibilityDate = periodDate
				}
				if currentBenefit != nil {
					currentBenefit.TimePeriodQual = periodQual
					currentBenefit.TimePeriod = periodDate
				}
			}

		// EB — Eligibility/Benefit Information
		case "EB":
			currentBenefit = &EligibilityBenefit{}
			if len(elements) > 0 {
				eb01 := getElement(elements, 0)
				currentBenefit.CoverageLevel = eb01
			}
			if len(elements) > 2 {
				currentBenefit.ServiceTypeCode = getElement(elements, 2)
				currentBenefit.ServiceTypeDesc = serviceTypeDescription(currentBenefit.ServiceTypeCode)
			}
			if len(elements) > 3 {
				currentBenefit.PlanCoverage = getElement(elements, 3)
			}
			if len(elements) > 6 {
				if amt, err := strconv.ParseFloat(getElement(elements, 6), 64); err == nil {
					currentBenefit.Amount = amt
				}
			}
			if len(elements) > 8 {
				if qty, err := strconv.ParseFloat(getElement(elements, 8), 64); err == nil {
					currentBenefit.Quantity = qty
				}
			}
			resp.Benefits = append(resp.Benefits, *currentBenefit)

		// MSG — Message Text (contains network info)
		case "MSG":
			if currentBenefit != nil && len(elements) > 0 {
				msgText := getElement(elements, 0)
				if strings.Contains(strings.ToLower(msgText), "network") ||
					strings.Contains(strings.ToLower(msgText), "in network") ||
					strings.Contains(strings.ToLower(msgText), "out of network") {
					currentBenefit.InPlanNetwork = msgText
				}
			}

		// REF — Reference
		case "REF":
			if len(elements) >= 2 {
				refQual := getElement(elements, 0)
				refID := getElement(elements, 1)
				if currentBenefit != nil {
					currentBenefit.RefIDQual = refQual
					currentBenefit.RefID = refID
				}
			}

		// TRN — Trace Number
		case "TRN":
			if len(elements) >= 2 {
				resp.TraceNumber = getElement(elements, 1)
			}

		// INS — Insured/Subscriber Relationship
		case "INS":
			// INS01 = Y/N (subscriber indicator)
			// INS02 = relationship code

		// BPR — Financial Information (contains response date)
		case "BPR":
			if len(elements) > 15 {
				resp.ResponseDate = getElement(elements, 15)
			}

		// SE, GE, IEA — Trailers (ignore)
		case "SE", "GE", "IEA":
			// End of segments — finalize
		}
	}

	// Assign subscriber/dependent to response
	if inSubscriber && currentSubscriber != nil {
		resp.Subscriber = *currentSubscriber
	}
	if currentDependent != nil {
		resp.Dependent = currentDependent
		if currentSubscriber != nil {
			resp.Subscriber = *currentSubscriber
		}
	}
	if currentSubscriber != nil && resp.Subscriber.MemberID == "" {
		resp.Subscriber = *currentSubscriber
	}

	return resp, nil
}

// getElement safely returns an element at the given index, or empty string.
func getElement(elements []string, idx int) string {
	if idx < len(elements) {
		return elements[idx]
	}
	return ""
}

// serviceTypeDescription returns a human-readable description for a service type code.
func serviceTypeDescription(code string) string {
	descriptions := map[string]string{
		"1":   "Medical Care",
		"2":   "Surgical",
		"3":   "Consultation",
		"4":   "Diagnostic X-Ray",
		"5":   "Diagnostic Lab",
		"6":   "Radiation Therapy",
		"7":   "Anesthesia",
		"8":   "Surgical Assistance",
		"12":  "Durable Medical Equipment Purchase",
		"13":  "Durable Medical Equipment Rental",
		"18":  "Dental Care",
		"30":  "Health Benefit Plan Coverage",
		"33":  "Chiropractic",
		"35":  "Dental Care",
		"47":  "Hospital",
		"48":  "Hospital — Inpatient",
		"50":  "Hospital — Outpatient",
		"51":  "Hospital — Emergency Accident",
		"52":  "Hospital — Emergency Medical",
		"53":  "Hospital Ambulatory Surgical",
		"54":  "Long Term Care",
		"86":  "Emergency Services",
		"88":  "Pharmacy",
		"98":  "Professional (Physician)",
		"AL":  "Vision (Optometry)",
		"MH":  "Mental Health",
		"UC":  "Urgent Care",
	}
	if desc, ok := descriptions[code]; ok {
		return desc
	}
	return fmt.Sprintf("Service Type %s", code)
}

// splitSegments is a copy of the 835 segment splitter for reuse in 271.
func splitSegments(data []byte) []string {
	// Strip newlines
	s := strings.ReplaceAll(string(data), "\r\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")

	segs := strings.Split(s, "~")
	var result []string
	for _, seg := range segs {
		seg = strings.TrimSpace(seg)
		if seg != "" {
			result = append(result, seg)
		}
	}
	return result
}

// segmentElements splits a segment string into tag and elements.
func segmentElements(seg string) (tag string, elements []string) {
	parts := strings.Split(seg, "*")
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}
