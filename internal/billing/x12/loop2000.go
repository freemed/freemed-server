package x12

// ---------- Loop 2000A – Billing Provider Hierarchical Level ----------

// Loop2000A represents the billing provider HL segment.
type Loop2000A struct {
	HL  Segment // HL*1**20*1~
	PRV Segment // PRV*BI*PXC*207Q00000X~
	CUR Segment // CUR*85*USD~
}

// ---------- Loop 2000B – Subscriber Hierarchical Level ----------

// Loop2000B represents the subscriber HL segment.
type Loop2000B struct {
	HL  Segment // HL*2*1*22*0~
	SBR Segment // SBR*P*18*MEMBER_ID******CI~
	PAT Segment // PAT*01~
}

// ---------- Loop 2000C – Patient Hierarchical Level ----------

// Loop2000C represents the patient HL segment (used when patient ≠ subscriber).
type Loop2000C struct {
	HL  Segment // HL*3*2*23*0~
	PAT Segment // PAT*01~
	DMG Segment // DMG*D8*19900101~
	NM1 Segment // NM1*QC*1*PATIENT LAST*PATIENT FIRST~
}
