package x12

import "strings"

// ---------- Loop 2400 – Service Line ----------

// Loop2400 holds a single service line on an 837 Professional claim.
type Loop2400 struct {
	LX      Segment // LX*1~
	SV1     Segment // SV1*HC:PROC:MOD1:MOD2:MOD3:MOD4*CHARGE*UN*UNITS***DIAG_PTRS~
	DTP472  Segment // DTP*472*D8*CCYYMMDD~  (Service Date)
}

// ---------- Helper: build SV1 segment ----------

// BuildSV1 constructs an SV1 (Professional Service) segment.
//
//	procedureCode: CPT or HCPCS code (e.g. "99213")
//	modifiers: up to 4 modifier codes (e.g. "25", "59")
//	charge: dollar amount for the line
//	units: number of service units
//	diagPointers: 1-based indices into the claim's diagnosis codes (concatenated as string, e.g. "12")
func BuildSV1(procedureCode string, modifiers []string, charge float64, units int, diagPointers []int) Segment {
	// Modifier sub-elements (MOD1:MOD2:MOD3:MOD4)
	modParts := make([]string, 4)
	for i := 0; i < 4; i++ {
		if i < len(modifiers) {
			modParts[i] = modifiers[i]
		}
	}

	// Build composite medical procedure identifier: HC:PROC:MOD1:MOD2:MOD3:MOD4
	procID := "HC" + SubElementSep + procedureCode
	for _, m := range modParts {
		procID += SubElementSep + m
	}
	// Trim trailing empty sub-elements
	procID = strings.TrimRight(procID, SubElementSep)

	// Build diagnosis pointers string (e.g. "12" for pointers 1 and 2)
	diagPtrStr := ""
	for i, p := range diagPointers {
		if i > 0 {
			diagPtrStr += SubElementSep
		}
		diagPtrStr += intToStr(p)
	}

	return NewSegment("SV1",
		procID,
		formatDollars(charge),
		"UN", // unit/basis for measurement code
		intToStr(units),
		"",   // SV105 reserved
		"",   // SV106 reserved
		diagPtrStr, // SV107 diagnosis code pointers
	)
}

// BuildLX constructs an LX (Service Line Number) segment.
func BuildLX(lineNumber int) Segment {
	return NewSegment("LX", intToStr(lineNumber))
}
