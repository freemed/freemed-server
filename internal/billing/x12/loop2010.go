package x12

// ---------- Loop 2010AA – Billing Provider Name ----------

// Loop2010AA holds the billing provider's demographic information.
type Loop2010AA struct {
	NM1 Segment // NM1*85*2*CLINIC NAME*****XX*TAX_ID~
	N3  Segment // N3*123 MAIN ST~
	N4  Segment // N4*CITY*ST*ZIP~
	REF Segment // REF*EI*TAX_ID~
	PER Segment // PER*IC*CONTACT*TE*PHONE~
}

// ---------- Loop 2010AB – Pay-to Provider Name ----------

// Loop2010AB holds the pay-to provider information (when different from billing).
type Loop2010AB struct {
	NM1 Segment // NM1*87*2*PAYTO NAME*****XX*TAX_ID~
	N3  Segment // N3*456 OAK AVE~
	N4  Segment // N4*CITY*ST*ZIP~
}

// ---------- Loop 2010BA – Subscriber Name ----------

// Loop2010BA holds the subscriber's demographic information.
type Loop2010BA struct {
	NM1 Segment // NM1*IL*1*LAST*FIRST*M****MI*MEMBER_ID~
	N3  Segment // N3*789 PINE ST~
	N4  Segment // N4*CITY*ST*ZIP~
	DMG Segment // DMG*D8*19800101~
	REF Segment // REF*SY*SSN~
}

// ---------- Loop 2010BB – Payer Name ----------

// Loop2010BB holds payer information.
type Loop2010BB struct {
	NM1 Segment // NM1*PR*2*PAYER NAME*****PI*PAYER_ID~
	N3  Segment // N3*PO BOX 12345~
	N4  Segment // N4*CITY*ST*ZIP~
	REF Segment // REF*2U*PAYER_ID~
}

// ---------- Loop 2010CA – Patient Name ----------

// Loop2010CA holds the patient's demographic information (when patient ≠ subscriber).
type Loop2010CA struct {
	NM1 Segment // NM1*QC*1*LAST*FIRST~
	N3  Segment // N3*123 MAIN ST~
	N4  Segment // N4*CITY*ST*ZIP~
	DMG Segment // DMG*D8*19900101~
	REF Segment // REF*SY*SSN~
}
