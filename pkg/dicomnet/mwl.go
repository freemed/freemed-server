package dicomnet

import (
	"errors"
	"fmt"
	"strings"
)

// DICOM tags used by the Modality Worklist Information Model (PS3.4 K.6.1)
// and by the surrounding patient/study attributes.
const (
	TagSpecificCharacterSet          uint32 = 0x00080005
	TagSOPClassUID                   uint32 = 0x00080016
	TagSOPInstanceUID                uint32 = 0x00080018
	TagStudyDate                     uint32 = 0x00080020
	TagSeriesDate                    uint32 = 0x00080021
	TagAccessionNumber               uint32 = 0x00080050
	TagModality                      uint32 = 0x00080060
	TagInstitutionName               uint32 = 0x00080080
	TagReferringPhysicianName        uint32 = 0x00080090
	TagStudyDescription              uint32 = 0x00081030
	TagPatientName                   uint32 = 0x00100010
	TagPatientID                     uint32 = 0x00100020
	TagPatientBirthDate              uint32 = 0x00100030
	TagPatientSex                    uint32 = 0x00100040
	TagStudyInstanceUID              uint32 = 0x0020000D
	TagRequestedProcedureDescription uint32 = 0x00321060
	TagRequestedContrastAgent        uint32 = 0x00321070
	TagScheduledProcedureStepSeq     uint32 = 0x00400100
	TagScheduledStationAETitle       uint32 = 0x00400001
	TagSPSStartDate                  uint32 = 0x00400002
	TagSPSStartTime                  uint32 = 0x00400003
	TagSPSEndDate                    uint32 = 0x00400004
	TagSPSEndTime                    uint32 = 0x00400005
	TagScheduledPerformingPhysName   uint32 = 0x00400006
	TagSPSDescription                uint32 = 0x00400007
	TagSPSID                         uint32 = 0x00400009
	TagSPSLocation                   uint32 = 0x00400011
	TagSPSStatus                     uint32 = 0x00400020
	TagSPSScheduledComments          uint32 = 0x00400400
	TagRequestedProcedureID          uint32 = 0x00401001
	TagRequestedProcedurePriority    uint32 = 0x00401004
)

// Errors returned while interpreting a C-FIND identifier.
var (
	ErrInvalidDateRange = errors.New("dicomnet: invalid date range")
	ErrInvalidTimeRange = errors.New("dicomnet: invalid time range")
	ErrNilQuery         = errors.New("dicomnet: nil MWL query")
)

// MWLItem is a single Modality Worklist entry — one scheduled procedure step
// for one patient. All dates are DICOM DA (YYYYMMDD) and all times are DICOM
// TM (HHMMSS) strings, empty when unknown.
type MWLItem struct {
	PatientName      string
	PatientID        string
	PatientBirthDate string
	PatientSex       string

	AccessionNumber        string
	StudyInstanceUID       string
	StudyDescription       string
	InstitutionName        string
	ReferringPhysicianName string

	RequestedProcedureID          string
	RequestedProcedureDescription string

	Modality                          string
	ScheduledStationAETitle           string
	ScheduledProcedureStepStartDate   string
	ScheduledProcedureStepStartTime   string
	ScheduledProcedureStepEndDate     string
	ScheduledProcedureStepEndTime     string
	ScheduledPerformingPhysicianName  string
	ScheduledProcedureStepID          string
	ScheduledProcedureStepDescription string
	ScheduledProcedureStepLocation    string
	ScheduledProcedureStepStatus      string
	RequestedContrastAgent            string
}

// Value returns the top-level attribute value for a tag. It reports false for
// attributes the item does not carry.
func (it MWLItem) Value(tag uint32) (string, bool) {
	switch tag {
	case TagPatientName:
		return it.PatientName, true
	case TagPatientID:
		return it.PatientID, true
	case TagPatientBirthDate:
		return it.PatientBirthDate, true
	case TagPatientSex:
		return it.PatientSex, true
	case TagAccessionNumber:
		return it.AccessionNumber, true
	case TagStudyInstanceUID:
		return it.StudyInstanceUID, true
	case TagStudyDescription:
		return it.StudyDescription, true
	case TagInstitutionName:
		return it.InstitutionName, true
	case TagReferringPhysicianName:
		return it.ReferringPhysicianName, true
	case TagRequestedProcedureID:
		return it.RequestedProcedureID, true
	case TagRequestedProcedureDescription:
		return it.RequestedProcedureDescription, true
	}
	return "", false
}

// SPSValue returns the Scheduled Procedure Step attribute value for a tag.
func (it MWLItem) SPSValue(tag uint32) (string, bool) {
	switch tag {
	case TagModality:
		return it.Modality, true
	case TagScheduledStationAETitle:
		return it.ScheduledStationAETitle, true
	case TagSPSStartDate:
		return it.ScheduledProcedureStepStartDate, true
	case TagSPSStartTime:
		return it.ScheduledProcedureStepStartTime, true
	case TagSPSEndDate:
		return it.ScheduledProcedureStepEndDate, true
	case TagSPSEndTime:
		return it.ScheduledProcedureStepEndTime, true
	case TagScheduledPerformingPhysName:
		return it.ScheduledPerformingPhysicianName, true
	case TagSPSID:
		return it.ScheduledProcedureStepID, true
	case TagSPSDescription:
		return it.ScheduledProcedureStepDescription, true
	case TagSPSLocation:
		return it.ScheduledProcedureStepLocation, true
	case TagSPSStatus:
		return it.ScheduledProcedureStepStatus, true
	case TagRequestedContrastAgent:
		return it.RequestedContrastAgent, true
	case TagStudyInstanceUID:
		return it.StudyInstanceUID, true
	}
	return "", false
}

// QueryKey is one attribute from a C-FIND request identifier. A zero-length
// Value marks a return key (universal matching); a non-empty Value marks a
// matching key.
type QueryKey struct {
	Tag   uint32
	VR    string
	Value string
	InSPS bool
	Items []DataSet // nested items when this key is a sequence
}

// IsReturnKey reports whether the key was requested with zero length.
func (k QueryKey) IsReturnKey() bool { return k.Value == "" }

// MWLQuery is a parsed Modality Worklist C-FIND request identifier.
type MWLQuery struct {
	// Keys are the attributes from the request identifier, in order.
	Keys []QueryKey

	// RequestedSOPClassUID is copied from the (0000,0002) command element when
	// the caller sets it.
	RequestedSOPClassUID string

	// UnsupportedMatchingKeys lists tags with a non-empty matching value that
	// the worklist model does not know how to match. They are returned with
	// zero length and ignored for filtering.
	UnsupportedMatchingKeys []uint32

	values  map[uint32]string
	spsVals map[uint32]string
}

// ParseMWLQuery interprets a C-FIND identifier data set as a Modality Worklist
// query. Malformed identifiers (which this function never trusts) are reported
// as errors; it never panics.
func ParseMWLQuery(ds DataSet) (*MWLQuery, error) {
	q := &MWLQuery{
		values:  make(map[uint32]string),
		spsVals: make(map[uint32]string),
	}
	for _, e := range ds {
		if e.Tag == 0 || e.Group() == 0xFFFE {
			return nil, fmt.Errorf("%w: 0x%08X", ErrInvalidTag, e.Tag)
		}
		vr := e.VR
		if vr == "" {
			vr = VRForTag(e.Tag)
		}
		if e.Tag == TagScheduledProcedureStepSeq && len(e.Items) > 0 {
			for _, item := range e.Items {
				for _, ie := range item {
					ivr := ie.VR
					if ivr == "" {
						ivr = VRForTag(ie.Tag)
					}
					q.Keys = append(q.Keys, QueryKey{
						Tag:   ie.Tag,
						VR:    ivr,
						Value: ie.String(),
						InSPS: true,
						Items: ie.Items,
					})
					if v := ie.String(); v != "" {
						q.spsVals[ie.Tag] = v
					}
				}
			}
			q.Keys = append(q.Keys, QueryKey{Tag: e.Tag, VR: "SQ", Value: "", InSPS: false, Items: e.Items})
			continue
		}
		q.Keys = append(q.Keys, QueryKey{Tag: e.Tag, VR: vr, Value: e.String(), Items: e.Items})
		if v := e.String(); v != "" {
			if vr == "SQ" {
				continue
			}
			if _, known := mwlMatchSupport(e.Tag, false); known {
				q.values[e.Tag] = v
			} else {
				q.UnsupportedMatchingKeys = append(q.UnsupportedMatchingKeys, e.Tag)
			}
		}
	}

	for tag := range q.spsVals {
		if _, known := mwlMatchSupport(tag, true); !known {
			q.UnsupportedMatchingKeys = append(q.UnsupportedMatchingKeys, tag)
		}
	}
	return q, nil
}

// MatchingValue returns the matching value supplied for a top-level tag.
func (q *MWLQuery) MatchingValue(tag uint32) (string, bool) {
	if q == nil {
		return "", false
	}
	v, ok := q.values[tag]
	return v, ok
}

// SPSMatchingValue returns the matching value supplied for a Scheduled
// Procedure Step tag.
func (q *MWLQuery) SPSMatchingValue(tag uint32) (string, bool) {
	if q == nil {
		return "", false
	}
	v, ok := q.spsVals[tag]
	return v, ok
}

// DateRange returns the requested Scheduled Procedure Step start date range.
func (q *MWLQuery) DateRange() (DateRange, error) {
	v, ok := q.SPSMatchingValue(TagSPSStartDate)
	if !ok {
		return DateRange{Universal: true}, nil
	}
	return ParseDateRange(v)
}

// TimeRange returns the requested Scheduled Procedure Step start time range.
func (q *MWLQuery) TimeRange() (TimeRange, error) {
	v, ok := q.SPSMatchingValue(TagSPSStartTime)
	if !ok {
		return TimeRange{Universal: true}, nil
	}
	return ParseTimeRange(v)
}

// MatchItem reports whether an item satisfies every matching key in the query.
// Matching keys the worklist model does not understand never filter items out
// (they are reported in UnsupportedMatchingKeys instead).
func (q *MWLQuery) MatchItem(it MWLItem) bool {
	if q == nil {
		return true
	}
	for tag, pattern := range q.values {
		kind, known := mwlMatchSupport(tag, false)
		if !known {
			continue
		}
		value, _ := it.Value(tag)
		switch kind {
		case matchDate:
			r, err := ParseDateRange(pattern)
			if err != nil {
				return false
			}
			if !r.Matches(value) {
				return false
			}
		case matchTime:
			r, err := ParseTimeRange(pattern)
			if err != nil {
				return false
			}
			if !r.Matches(value) {
				return false
			}
		default:
			if !MatchMultiValue(pattern, value) {
				return false
			}
		}
	}
	for tag, pattern := range q.spsVals {
		kind, known := mwlMatchSupport(tag, true)
		if !known {
			continue
		}
		value, _ := it.SPSValue(tag)
		switch kind {
		case matchDate:
			r, err := ParseDateRange(pattern)
			if err != nil {
				return false
			}
			if !r.Matches(value) {
				return false
			}
		case matchTime:
			r, err := ParseTimeRange(pattern)
			if err != nil {
				return false
			}
			if !r.Matches(value) {
				return false
			}
		default:
			if !MatchMultiValue(pattern, value) {
				return false
			}
		}
	}
	return true
}

type matchKind int

const (
	matchText matchKind = iota
	matchDate
	matchTime
)

// mwlMatchSupport reports how a tag is matched and whether the worklist model
// supports matching on it at all.
func mwlMatchSupport(tag uint32, inSPS bool) (matchKind, bool) {
	if inSPS {
		switch tag {
		case TagModality, TagScheduledStationAETitle, TagScheduledPerformingPhysName,
			TagSPSID, TagSPSDescription, TagSPSLocation, TagSPSStatus, TagRequestedContrastAgent,
			TagStudyInstanceUID:
			return matchText, true
		case TagSPSStartDate, TagSPSEndDate:
			return matchDate, true
		case TagSPSStartTime, TagSPSEndTime:
			return matchTime, true
		}
		return matchText, false
	}
	switch tag {
	case TagPatientName, TagPatientID, TagAccessionNumber, TagStudyInstanceUID,
		TagRequestedProcedureID, TagRequestedProcedureDescription, TagStudyDescription,
		TagInstitutionName, TagReferringPhysicianName:
		return matchText, true
	case TagPatientBirthDate:
		return matchDate, true
	}
	return matchText, false
}

// BuildResponseIdentifier builds the C-FIND response identifier for one item.
// Only the attributes present in the request identifier are returned, matching
// keys filled in with the item's actual values (PS3.4 C.6.2.1.2).
func (q *MWLQuery) BuildResponseIdentifier(it MWLItem) (DataSet, error) {
	if q == nil {
		return nil, ErrNilQuery
	}
	var ds DataSet
	var spsKeys []QueryKey

	for _, k := range q.Keys {
		if k.InSPS {
			spsKeys = append(spsKeys, k)
			continue
		}
		if k.Tag == TagScheduledProcedureStepSeq {
			continue // emitted below, once all SPS keys are known
		}
		vr := k.VR
		if vr == "" {
			vr = VRForTag(k.Tag)
		}
		if vr == "SQ" {
			// Unsupported sequence: return it zero-length rather than lying.
			ds = append(ds, Element{Tag: k.Tag, VR: "SQ"})
			continue
		}
		v, _ := it.Value(k.Tag)
		if !mwlResponseVRs[k.Tag] {
			// Attribute we do not populate: zero length, correct VR.
			v = ""
		}
		ds = append(ds, NewStringElement(k.Tag, vr, v))
	}

	if len(spsKeys) > 0 {
		var item DataSet
		for _, k := range spsKeys {
			vr := k.VR
			if vr == "" {
				vr = VRForTag(k.Tag)
			}
			if vr == "SQ" {
				item = append(item, Element{Tag: k.Tag, VR: "SQ"})
				continue
			}
			v, _ := it.SPSValue(k.Tag)
			if !mwlResponseVRs[k.Tag] {
				v = ""
			}
			item = append(item, NewStringElement(k.Tag, vr, v))
		}
		ds = append(ds, NewSequenceElement(TagScheduledProcedureStepSeq, item))
	}
	return ds, nil
}

// mwlResponseVRs marks attributes this implementation can populate. Requested
// attributes outside the set are returned with zero length, which is legal and
// tells the SCU the SCP has no value for them.
var mwlResponseVRs = map[uint32]bool{
	TagSpecificCharacterSet:          true,
	TagPatientName:                   true,
	TagPatientID:                     true,
	TagPatientBirthDate:              true,
	TagPatientSex:                    true,
	TagAccessionNumber:               true,
	TagStudyInstanceUID:              true,
	TagStudyDescription:              true,
	TagInstitutionName:               true,
	TagReferringPhysicianName:        true,
	TagRequestedProcedureID:          true,
	TagRequestedProcedureDescription: true,

	TagModality:                    true,
	TagScheduledStationAETitle:     true,
	TagSPSStartDate:                true,
	TagSPSStartTime:                true,
	TagSPSEndDate:                  true,
	TagSPSEndTime:                  true,
	TagScheduledPerformingPhysName: true,
	TagSPSID:                       true,
	TagSPSDescription:              true,
	TagSPSLocation:                 true,
	TagSPSStatus:                   true,
	TagRequestedContrastAgent:      true,
}

// MatchWildcard performs DICOM single value matching with the wildcard
// characters '*' (zero or more characters) and '?' (exactly one character).
// The comparison is case insensitive, as required for PN/LO/CS matching
// (PS3.4 C.2.2.2). It is iterative (no recursion) and cannot be made to blow
// the stack by hostile input.
func MatchWildcard(pattern, value string) bool {
	pr := []rune(strings.ToUpper(pattern))
	vr := []rune(strings.ToUpper(value))

	pi, vi := 0, 0
	star := -1
	mark := 0
	for vi < len(vr) {
		switch {
		// A '*' in the pattern is always a wildcard: DICOM has no escaping,
		// so it is tested before any literal comparison.
		case pi < len(pr) && pr[pi] == '*':
			star = pi
			mark = vi
			pi++
		case pi < len(pr) && (pr[pi] == '?' || pr[pi] == vr[vi]):
			pi++
			vi++
		case star >= 0:
			pi = star + 1
			mark++
			vi = mark
		default:
			return false
		}
	}
	for pi < len(pr) && pr[pi] == '*' {
		pi++
	}
	return pi == len(pr)
}

// MatchMultiValue matches a (possibly multi-valued) query pattern against a
// (possibly multi-valued) attribute value. Every component of the pattern must
// match at least one component of the value.
func MatchMultiValue(pattern, value string) bool {
	patterns := SplitValues(pattern)
	if len(patterns) == 0 {
		return true // universal matching
	}
	values := SplitValues(value)
	if len(values) == 0 {
		values = []string{""}
	}
	for _, p := range patterns {
		matched := false
		for _, v := range values {
			if MatchWildcard(p, v) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

// DateRange is a DICOM DA range: "YYYYMMDD", "YYYYMMDD-YYYYMMDD",
// "YYYYMMDD-" (open ended) or "-YYYYMMDD" (open started). A single date is an
// exact match.
type DateRange struct {
	Start     string // "" means unbounded
	End       string // "" means unbounded
	Single    bool
	Universal bool
}

// ParseDateRange parses a DICOM DA matching value.
func ParseDateRange(s string) (DateRange, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "*" {
		return DateRange{Universal: true}, nil
	}
	parts := strings.SplitN(s, "-", 2)
	if len(parts) == 1 {
		if !validDate(parts[0]) {
			return DateRange{}, fmt.Errorf("%w: %q", ErrInvalidDateRange, s)
		}
		return DateRange{Start: parts[0], End: parts[0], Single: true}, nil
	}
	start, end := parts[0], parts[1]
	if start != "" && !validDate(start) {
		return DateRange{}, fmt.Errorf("%w: %q", ErrInvalidDateRange, s)
	}
	if end != "" && !validDate(end) {
		return DateRange{}, fmt.Errorf("%w: %q", ErrInvalidDateRange, s)
	}
	if start == "" && end == "" {
		return DateRange{Universal: true}, nil
	}
	return DateRange{Start: start, End: end}, nil
}

// Matches reports whether a DICOM DA value falls inside the range. An empty
// date only matches a universal range.
func (r DateRange) Matches(date string) bool {
	date = strings.TrimSpace(date)
	if r.Universal {
		return true
	}
	if date == "" {
		return false
	}
	if r.Start != "" && date < r.Start {
		return false
	}
	if r.End != "" && date > r.End {
		return false
	}
	return true
}

// TimeRange is a DICOM TM range, with the same syntax as DateRange.
type TimeRange struct {
	Start     string
	End       string
	Single    bool
	Universal bool
}

// ParseTimeRange parses a DICOM TM matching value (HHMMSS or HHMM precision).
func ParseTimeRange(s string) (TimeRange, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "*" {
		return TimeRange{Universal: true}, nil
	}
	parts := strings.SplitN(s, "-", 2)
	if len(parts) == 1 {
		if !validTime(parts[0]) {
			return TimeRange{}, fmt.Errorf("%w: %q", ErrInvalidTimeRange, s)
		}
		return TimeRange{Start: parts[0], End: parts[0], Single: true}, nil
	}
	start, end := parts[0], parts[1]
	if start != "" && !validTime(start) {
		return TimeRange{}, fmt.Errorf("%w: %q", ErrInvalidTimeRange, s)
	}
	if end != "" && !validTime(end) {
		return TimeRange{}, fmt.Errorf("%w: %q", ErrInvalidTimeRange, s)
	}
	if start == "" && end == "" {
		return TimeRange{Universal: true}, nil
	}
	return TimeRange{Start: start, End: end}, nil
}

// Matches reports whether a DICOM TM value falls inside the range. Comparison
// is prefix based, so a reduced precision value ("09" for the ninth hour)
// matches full precision values ("090000") as required by PS3.4 C.2.2.2.
func (r TimeRange) Matches(tm string) bool {
	tm = strings.TrimSpace(tm)
	if r.Universal {
		return true
	}
	if tm == "" {
		return false
	}
	if r.Start != "" {
		n := min(len(tm), len(r.Start))
		if tm[:n] < r.Start[:n] {
			return false
		}
	}
	if r.End != "" {
		n := min(len(tm), len(r.End))
		if tm[:n] > r.End[:n] {
			return false
		}
	}
	return true
}

// validDate reports whether s is a plausible DICOM DA (YYYYMMDD, digits only).
func validDate(s string) bool {
	if len(s) != 8 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// validTime reports whether s is a plausible DICOM TM (HH, HHMM or HHMMSS,
// digits only).
func validTime(s string) bool {
	if len(s) != 2 && len(s) != 4 && len(s) != 6 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
