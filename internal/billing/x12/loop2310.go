package x12

// ---------- Loop 2310A – Referring Provider Name ----------

// Loop2310A holds the referring provider information.
type Loop2310A struct {
	NM1 Segment // NM1*DN*1*LAST*FIRST*M****XX*NPI~
	REF Segment // REF*1C*NPI~
}

// ---------- Loop 2310B – Rendering Provider Name ----------

// Loop2310B holds the rendering provider information.
type Loop2310B struct {
	NM1 Segment // NM1*82*1*LAST*FIRST*M****XX*NPI~
	REF Segment // REF*1C*NPI~
	PRV Segment // PRV*PE*PXC*207Q00000X~
}
