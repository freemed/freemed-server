package dicomnet

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

type fakeSource struct {
	mu      sync.Mutex
	items   []MWLItem
	err     error
	queries int
	last    *MWLQuery
	block   chan struct{}
}

func (f *fakeSource) Find(ctx context.Context, q *MWLQuery) ([]MWLItem, error) {
	f.mu.Lock()
	f.queries++
	f.last = q
	block := f.block
	f.mu.Unlock()
	if block != nil {
		select {
		case <-block:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func (f *fakeSource) queryCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.queries
}

// ---------------------------------------------------------------------------
// Hand-built SCU: reference PDU construction and reference response parsing.
// Deliberately independent of the production encoder/decoder so that a
// symmetric bug cannot make the test pass.
// ---------------------------------------------------------------------------

type refPC struct {
	id               byte
	abstract         string
	transferSyntaxes []string
}

func buildAssociateRQ(called, calling string, pcs []refPC, maxPDU uint32) []byte {
	var body []byte
	body = append(body, 0x00, 0x01, 0x00, 0x00) // protocol version 1 + reserved
	body = append(body, refAETitle(called)...)
	body = append(body, refAETitle(calling)...)
	body = append(body, make([]byte, 32)...)
	body = append(body, refItem(itemApplicationContext, []byte(UIDApplicationContext))...)
	for _, pc := range pcs {
		var p []byte
		p = append(p, pc.id, 0x00, 0x00, 0x00)
		p = append(p, refItem(itemAbstractSyntax, []byte(pc.abstract))...)
		for _, ts := range pc.transferSyntaxes {
			p = append(p, refItem(itemTransferSyntax, []byte(ts))...)
		}
		body = append(body, refItem(itemPresentationContextRQ, p)...)
	}
	var ui []byte
	ml := make([]byte, 4)
	binary.BigEndian.PutUint32(ml, maxPDU)
	ui = append(ui, refItem(itemMaxLength, ml)...)
	ui = append(ui, refItem(itemImplementationClassUID, []byte(ImplementationClassUID))...)
	ui = append(ui, refItem(itemImplementationVersion, []byte(ImplementationVersionName))...)
	body = append(body, refItem(itemUserInformation, ui)...)
	return refPDUBytes(PDUAssociateRQ, body)
}

// buildDataTF wraps a command set and an optional data set in one last-fragment
// PDV each.
func buildDataTF(contextID byte, command, data []byte) []byte {
	var body []byte
	add := func(v []byte, isCommand bool) {
		mch := MCHLastBit
		if isCommand {
			mch |= MCHCommandBit
		}
		hdr := make([]byte, 6)
		binary.BigEndian.PutUint32(hdr[0:4], uint32(len(v)+2))
		hdr[4] = contextID
		hdr[5] = mch
		body = append(body, hdr...)
		body = append(body, v...)
	}
	add(command, true)
	if len(data) > 0 {
		add(data, false)
	}
	return refPDUBytes(PDUDataTF, body)
}

func buildReleaseRQ() []byte { return refPDUBytes(PDUReleaseRQ, make([]byte, 4)) }
func buildReleaseRP() []byte { return refPDUBytes(PDUReleaseRP, make([]byte, 4)) }
func buildAbort() []byte     { return refPDUBytes(PDUAbort, []byte{0x00, 0x00, 0x00, 0x00}) }

func readRawPDU(t *testing.T, conn net.Conn) (byte, []byte) {
	t.Helper()
	hdr := make([]byte, 6)
	if _, err := io.ReadFull(conn, hdr); err != nil {
		t.Fatalf("reading PDU header: %v", err)
	}
	length := binary.BigEndian.Uint32(hdr[2:6])
	if length > MaxPDUBodySize {
		t.Fatalf("peer announced a %d byte PDU body", length)
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(conn, body); err != nil {
		t.Fatalf("reading PDU body (%d bytes): %v", length, err)
	}
	return hdr[0], body
}

type acInfo struct {
	appContext string
	maxPDU     uint32
	results    map[byte]byte
	transfer   map[byte]string
}

func parseAC(t *testing.T, body []byte) acInfo {
	t.Helper()
	const fixed = 68 // version(2) + reserved(2) + called AE(16) + calling AE(16) + reserved(32)
	if len(body) < fixed {
		t.Fatalf("A-ASSOCIATE-AC body is %d bytes", len(body))
	}
	info := acInfo{results: map[byte]byte{}, transfer: map[byte]string{}}
	pos := fixed
	for pos < len(body) {
		if pos+4 > len(body) {
			t.Fatalf("truncated AC item header at %d", pos)
		}
		typ := body[pos]
		l := int(binary.BigEndian.Uint16(body[pos+2 : pos+4]))
		if pos+4+l > len(body) {
			t.Fatalf("AC item 0x%02X overruns the PDU", typ)
		}
		item := body[pos+4 : pos+4+l]
		switch typ {
		case itemApplicationContext:
			info.appContext = string(item)
		case itemPresentationContextAC:
			if len(item) < 4 {
				t.Fatalf("short presentation context AC item")
			}
			id, result := item[0], item[2]
			info.results[id] = result
			for s := 4; s < len(item); {
				st := item[s]
				sl := int(binary.BigEndian.Uint16(item[s+2 : s+4]))
				sub := item[s+4 : s+4+sl]
				if st == itemTransferSyntax {
					info.transfer[id] = strings.TrimRight(string(sub), " \x00")
				}
				s += 4 + sl
			}
		case itemUserInformation:
			for s := 0; s < len(item); {
				st := item[s]
				sl := int(binary.BigEndian.Uint16(item[s+2 : s+4]))
				sub := item[s+4 : s+4+sl]
				if st == itemMaxLength && len(sub) >= 4 {
					info.maxPDU = binary.BigEndian.Uint32(sub)
				}
				s += 4 + sl
			}
		}
		pos += 4 + l
	}
	return info
}

// refParseElements parses an Implicit VR Little Endian stream by hand.
func refParseElements(t *testing.T, data []byte) map[uint32][]byte {
	t.Helper()
	out := map[uint32][]byte{}
	for pos := 0; pos < len(data); {
		if pos+8 > len(data) {
			t.Fatalf("truncated element header at %d (len %d)", pos, len(data))
		}
		tag := uint32(binary.LittleEndian.Uint16(data[pos:pos+2]))<<16 |
			uint32(binary.LittleEndian.Uint16(data[pos+2:pos+4]))
		l := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		if pos+8+l > len(data) {
			t.Fatalf("element 0x%08X claims %d bytes, only %d remain", tag, l, len(data)-pos-8)
		}
		v := make([]byte, l)
		copy(v, data[pos+8:pos+8+l])
		out[tag] = v
		pos += 8 + l
	}
	return out
}

func refUS(t *testing.T, elems map[uint32][]byte, tag uint32) uint16 {
	t.Helper()
	v, ok := elems[tag]
	if !ok || len(v) < 2 {
		t.Fatalf("command element 0x%08X missing", tag)
	}
	return binary.LittleEndian.Uint16(v)
}

// refCommandHasDataSet reports whether the reference command set announces a
// following data set.
func refCommandHasDataSet(t *testing.T, cmd []byte) bool {
	t.Helper()
	elems := refParseElements(t, cmd)
	v, ok := elems[TagCommandDataSetType]
	if !ok {
		return false
	}
	if len(v) < 2 {
		t.Fatalf("malformed (0000,0800)")
	}
	return binary.LittleEndian.Uint16(v) != DataSetTypeNone
}

// readMessage reassembles one DIMSE message using the wire layout directly.
func readMessage(t *testing.T, conn net.Conn) (command, data []byte) {
	t.Helper()
	cmdDone := false
	for {
		typ, body := readRawPDU(t, conn)
		if typ != PDUDataTF {
			t.Fatalf("expected a P-DATA-TF PDU, got type 0x%02X", typ)
		}
		for pos := 0; pos < len(body); {
			if pos+6 > len(body) {
				t.Fatalf("truncated PDV at %d", pos)
			}
			plen := int(binary.BigEndian.Uint32(body[pos : pos+4]))
			if plen < 2 || pos+4+plen > len(body) {
				t.Fatalf("PDV length %d is out of range", plen)
			}
			mch := body[pos+5]
			chunk := body[pos+6 : pos+4+plen]
			isCommand := mch&MCHCommandBit != 0
			isLast := mch&MCHLastBit != 0
			if isCommand {
				command = append(command, chunk...)
			} else {
				data = append(data, chunk...)
			}
			pos += 4 + plen
			if isLast && isCommand {
				cmdDone = true
				if !refCommandHasDataSet(t, command) {
					return command, nil
				}
			}
			if isLast && !isCommand && cmdDone {
				return command, data
			}
		}
	}
}

// readMessageWithMaxPDU behaves like readMessage but also records the largest
// PDU body seen, so fragmentation against the peer's maximum length can be
// asserted.
func readMessageWithMaxPDU(t *testing.T, conn net.Conn) (command, data []byte, maxBody int, pdus int) {
	t.Helper()
	cmdDone := false
	for {
		hdr := make([]byte, 6)
		if _, err := io.ReadFull(conn, hdr); err != nil {
			t.Fatalf("reading PDU header: %v", err)
		}
		length := int(binary.BigEndian.Uint32(hdr[2:6]))
		body := make([]byte, length)
		if _, err := io.ReadFull(conn, body); err != nil {
			t.Fatalf("reading PDU body: %v", err)
		}
		if hdr[0] != PDUDataTF {
			t.Fatalf("expected a P-DATA-TF PDU, got type 0x%02X", hdr[0])
		}
		pdus++
		if length > maxBody {
			maxBody = length
		}
		for pos := 0; pos < len(body); {
			plen := int(binary.BigEndian.Uint32(body[pos : pos+4]))
			if plen < 2 || pos+4+plen > len(body) {
				t.Fatalf("PDV length %d is out of range", plen)
			}
			mch := body[pos+5]
			chunk := body[pos+6 : pos+4+plen]
			isCommand := mch&MCHCommandBit != 0
			isLast := mch&MCHLastBit != 0
			if isCommand {
				command = append(command, chunk...)
			} else {
				data = append(data, chunk...)
			}
			pos += 4 + plen
			if isLast && isCommand {
				cmdDone = true
				if !refCommandHasDataSet(t, command) {
					return command, nil, maxBody, pdus
				}
			}
			if isLast && !isCommand && cmdDone {
				return command, data, maxBody, pdus
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Client helpers
// ---------------------------------------------------------------------------

func dial(t *testing.T, addr string) net.Conn {
	t.Helper()
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatalf("dial %s: %v", addr, err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		t.Fatalf("SetDeadline: %v", err)
	}
	return conn
}

// associate performs an association and returns the parsed AC, the context ID
// chosen for the MWL SOP class, and the negotiated transfer syntax.
func associate(t *testing.T, conn net.Conn, called string, pcs []refPC, maxPDU uint32) (acInfo, byte, string) {
	t.Helper()
	if _, err := conn.Write(buildAssociateRQ(called, "SCU_TEST", pcs, maxPDU)); err != nil {
		t.Fatalf("writing A-ASSOCIATE-RQ: %v", err)
	}
	typ, body := readRawPDU(t, conn)
	if typ != PDUAssociateAC {
		t.Fatalf("expected A-ASSOCIATE-AC (0x02), got PDU type 0x%02X", typ)
	}
	info := parseAC(t, body)
	var ctxID byte
	for _, pc := range pcs {
		if pc.abstract == UIDModalityWorklistFIND && info.results[pc.id] == PresContextAcceptance {
			ctxID = pc.id
			break
		}
	}
	if ctxID == 0 {
		t.Fatalf("no MWL presentation context accepted: %+v", info.results)
	}
	return info, ctxID, info.transfer[ctxID]
}

// readCFindResponses reads C-FIND responses until a final (non pending) status
// arrives, returning the statuses and the response identifiers.
func readCFindResponses(t *testing.T, conn net.Conn) ([]uint16, [][]byte) {
	t.Helper()
	var statuses []uint16
	var identifiers [][]byte
	for i := 0; i < 64; i++ {
		command, data := readMessage(t, conn)
		elems := refParseElements(t, command)
		if got := refUS(t, elems, TagCommandField); got != CmdCFindRSP {
			t.Fatalf("command field = 0x%04X, want 0x%04X", got, CmdCFindRSP)
		}
		status := refUS(t, elems, TagStatus)
		if v := refUS(t, elems, TagMessageIDBeingRespondedTo); v != 1 {
			t.Fatalf("message ID being responded to = %d, want 1", v)
		}
		if _, ok := elems[TagAffectedSOPClassUID]; ok {
			if sop := strings.TrimRight(string(elems[TagAffectedSOPClassUID]), " \x00"); sop != UIDModalityWorklistFIND {
				t.Fatalf("affected SOP class = %q", sop)
			}
		}
		statuses = append(statuses, status)
		identifiers = append(identifiers, data)
		if status != StatusPending {
			return statuses, identifiers
		}
	}
	t.Fatal("too many pending C-FIND responses")
	return nil, nil
}

// standardCFindIdentifier encodes the identifier used by the client.
func standardCFindIdentifier(t *testing.T, explicit bool) []byte {
	t.Helper()
	raw, err := EncodeDataSet(standardIdentifier(), explicit)
	if err != nil {
		t.Fatalf("EncodeDataSet: %v", err)
	}
	return raw
}

// ---------------------------------------------------------------------------
// Server harness
// ---------------------------------------------------------------------------

type testServer struct {
	srv    *Server
	addr   string
	cancel context.CancelFunc
	errCh  chan error
}

func startTestServer(t *testing.T, cfg Config) *testServer {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	srv, err := NewServer(cfg)
	if err != nil {
		cancel()
		t.Fatalf("NewServer: %v", err)
	}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ctx, ln) }()
	ts := &testServer{srv: srv, addr: ln.Addr().String(), cancel: cancel, errCh: errCh}
	t.Cleanup(func() {
		cancel()
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		_ = srv.Shutdown(shutCtx)
		select {
		case <-errCh:
		case <-time.After(5 * time.Second):
		}
	})
	return ts
}

func mwlConfig(src WorklistSource) Config {
	return Config{
		AETitle:          "FREEMED",
		MaxAssociations:  4,
		IdleTimeout:      10 * time.Second,
		HandshakeTimeout: 5 * time.Second,
		Logf:             func(string, ...any) {}, // keep test output quiet
		Source:           src,
	}
}

func sourceItems() []MWLItem {
	base := sampleItem()
	second := sampleItem()
	second.PatientName = "SMITH^JANE"
	second.RequestedProcedureID = "RP1002"
	second.ScheduledProcedureStepID = "SPS1002"
	other := sampleItem()
	other.Modality = "MR"
	other.PatientName = "JONES^MARY"
	outOfRange := sampleItem()
	outOfRange.PatientName = "SMITH^ALICE"
	outOfRange.ScheduledProcedureStepStartDate = "20240220"
	return []MWLItem{base, second, other, outOfRange}
}

// ---------------------------------------------------------------------------
// End-to-end association
// ---------------------------------------------------------------------------

func TestSCPAssociationAndCFind(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	ts := startTestServer(t, mwlConfig(src))

	conn := dial(t, ts.addr)
	pcs := []refPC{{
		id:               1,
		abstract:         UIDModalityWorklistFIND,
		transferSyntaxes: []string{ImplicitVRLittleEndian, ExplicitVRLittleEndian},
	}}
	info, ctxID, tsUID := associate(t, conn, "FREEMED", pcs, 16384)

	if info.results[1] != PresContextAcceptance {
		t.Fatalf("presentation context 1 result = 0x%02X", info.results[1])
	}
	if info.appContext != UIDApplicationContext {
		t.Errorf("application context = %q", info.appContext)
	}
	if info.maxPDU == 0 || info.maxPDU > MaxPDUBodySize {
		t.Errorf("advertised maximum PDU length = %d", info.maxPDU)
	}
	if tsUID != ImplicitVRLittleEndian {
		t.Errorf("negotiated transfer syntax = %q, want %q", tsUID, ImplicitVRLittleEndian)
	}

	// C-FIND-RQ with a real MWL identifier.
	identifier := standardCFindIdentifier(t, tsUID == ExplicitVRLittleEndian)
	request := NewCFindRequest(1, UIDModalityWorklistFIND, 0)
	requestBytes, err := request.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := conn.Write(buildDataTF(ctxID, requestBytes, identifier)); err != nil {
		t.Fatalf("writing C-FIND-RQ: %v", err)
	}

	statuses, identifiers := readCFindResponses(t, conn)
	if len(statuses) < 2 {
		t.Fatalf("statuses = %v, want at least one pending plus a final success", statuses)
	}
	pending := 0
	for i, st := range statuses {
		if i == len(statuses)-1 {
			if st != StatusSuccess {
				t.Fatalf("final status = 0x%04X, want 0x0000", st)
			}
			if len(identifiers[i]) != 0 {
				t.Errorf("final response carried a %d byte identifier", len(identifiers[i]))
			}
			continue
		}
		if st != StatusPending {
			t.Fatalf("status %d = 0x%04X, want 0xFF00 (pending)", i, st)
		}
		if len(identifiers[i]) == 0 {
			t.Fatalf("pending response %d carried an empty identifier", i)
		}
		pending++
	}
	if pending != 2 {
		t.Fatalf("pending responses = %d, want 2 (the two matching CT items)", pending)
	}

	// Independent byte-level check: the response identifier must carry the
	// PatientName tag followed by the ASCII value from the worklist source.
	first := identifiers[0]
	wantTag := []byte{0x10, 0x00, 0x10, 0x00} // (0010,0010) group first
	idx := bytes.Index(first, wantTag)
	if idx < 0 {
		t.Fatalf("PatientName tag not found in the response identifier: %x", first)
	}
	if !bytes.Contains(first, []byte("SMITH^")) {
		t.Errorf("response identifier does not contain the patient name: %x", first)
	}
	if !bytes.Contains(first, []byte("2.25.")) {
		t.Errorf("response identifier does not contain the study instance UID: %x", first)
	}

	// Decode the identifiers with the production decoder and check the
	// requested attributes came back with the item's values.
	for i := 0; i < pending; i++ {
		ds, err := DecodeDataSet(identifiers[i], tsUID == ExplicitVRLittleEndian)
		if err != nil {
			t.Fatalf("DecodeDataSet(response %d): %v", i, err)
		}
		if v := ds.GetString(TagPatientName); !strings.HasPrefix(v, "SMITH^") {
			t.Errorf("response %d PatientName = %q", i, v)
		}
		if v := ds.GetString(TagPatientID); v != "MRN0001234" {
			t.Errorf("response %d PatientID = %q", i, v)
		}
		sps, ok := ds.GetSequence(TagScheduledProcedureStepSeq)
		if !ok {
			t.Fatalf("response %d has no ScheduledProcedureStepSequence", i)
		}
		if v := sps.GetString(TagModality); v != "CT" {
			t.Errorf("response %d Modality = %q", i, v)
		}
		if v := sps.GetString(TagSPSStartDate); v != "20240115" {
			t.Errorf("response %d SPS start date = %q", i, v)
		}
		if v := sps.GetString(TagSPSID); v == "" {
			t.Errorf("response %d SPS ID is empty", i)
		}
	}
	if got := src.queryCount(); got != 1 {
		t.Errorf("worklist source queried %d times, want 1", got)
	}

	// A-RELEASE-RQ must be answered with A-RELEASE-RP.
	if _, err := conn.Write(buildReleaseRQ()); err != nil {
		t.Fatalf("writing A-RELEASE-RQ: %v", err)
	}
	typ, _ := readRawPDU(t, conn)
	if typ != PDUReleaseRP {
		t.Fatalf("expected A-RELEASE-RP (0x06), got 0x%02X", typ)
	}
	// The server must then close the association.
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Read(make([]byte, 1)); err == nil {
		t.Error("association still open after A-RELEASE-RP")
	}
}

func TestSCPExplicitVRLittleEndianAssociation(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	ts := startTestServer(t, mwlConfig(src))

	conn := dial(t, ts.addr)
	pcs := []refPC{{
		id:               3,
		abstract:         UIDModalityWorklistFIND,
		transferSyntaxes: []string{ExplicitVRLittleEndian},
	}}
	_, ctxID, tsUID := associate(t, conn, "FREEMED", pcs, 16384)
	if tsUID != ExplicitVRLittleEndian {
		t.Fatalf("negotiated transfer syntax = %q, want explicit VR LE", tsUID)
	}

	identifier := standardCFindIdentifier(t, true)
	requestBytes, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := conn.Write(buildDataTF(ctxID, requestBytes, identifier)); err != nil {
		t.Fatalf("writing C-FIND-RQ: %v", err)
	}
	statuses, identifiers := readCFindResponses(t, conn)
	if statuses[len(statuses)-1] != StatusSuccess {
		t.Fatalf("final status = 0x%04X", statuses[len(statuses)-1])
	}
	// The identifier must be explicit VR LE: decoding it as implicit VR must
	// fail, which proves the SCP honoured the negotiated transfer syntax.
	if _, err := DecodeDataSet(identifiers[0], false); err == nil {
		t.Error("response identifier decoded as implicit VR although explicit VR LE was negotiated")
	}
	ds, err := DecodeDataSet(identifiers[0], true)
	if err != nil {
		t.Fatalf("DecodeDataSet(explicit): %v", err)
	}
	if v := ds.GetString(TagPatientName); !strings.HasPrefix(v, "SMITH^") {
		t.Errorf("PatientName = %q", v)
	}
}

func TestSCPFragmentsResponsesToPeerMaxPDU(t *testing.T) {
	// Long attribute values force a response identifier larger than the
	// 512 byte PDU the client advertises.
	big := sampleItem()
	big.PatientName = "SMITH^JOHN^" + strings.Repeat("X", 250)
	big.InstitutionName = strings.Repeat("INSTITUTION", 20)
	big.StudyDescription = strings.Repeat("DESCRIPTION", 20)
	src := &fakeSource{items: []MWLItem{big}}
	ts := startTestServer(t, mwlConfig(src))

	conn := dial(t, ts.addr)
	pcs := []refPC{{
		id:               1,
		abstract:         UIDModalityWorklistFIND,
		transferSyntaxes: []string{ImplicitVRLittleEndian},
	}}
	_, ctxID, _ := associate(t, conn, "FREEMED", pcs, 512)

	// Request every attribute the model can return.
	sps := DataSet{
		NewStringElement(TagModality, "CS", ""),
		NewStringElement(TagScheduledStationAETitle, "AE", ""),
		NewStringElement(TagSPSStartDate, "DA", ""),
		NewStringElement(TagSPSStartTime, "TM", ""),
		NewStringElement(TagScheduledPerformingPhysName, "PN", ""),
		NewStringElement(TagSPSID, "SH", ""),
		NewStringElement(TagSPSDescription, "LO", ""),
		NewStringElement(TagSPSLocation, "SH", ""),
		NewStringElement(TagSPSStatus, "CS", ""),
	}
	idDS := DataSet{
		NewStringElement(TagPatientName, "PN", ""),
		NewStringElement(TagPatientID, "LO", ""),
		NewStringElement(TagPatientBirthDate, "DA", ""),
		NewStringElement(TagPatientSex, "CS", ""),
		NewStringElement(TagAccessionNumber, "SH", ""),
		NewStringElement(TagStudyInstanceUID, "UI", ""),
		NewStringElement(TagStudyDescription, "LO", ""),
		NewStringElement(TagInstitutionName, "LO", ""),
		NewStringElement(TagReferringPhysicianName, "PN", ""),
		NewStringElement(TagRequestedProcedureID, "SH", ""),
		NewSequenceElement(TagScheduledProcedureStepSeq, sps),
	}
	identifier, err := EncodeDataSet(idDS, false)
	if err != nil {
		t.Fatalf("EncodeDataSet: %v", err)
	}
	requestBytes, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := conn.Write(buildDataTF(ctxID, requestBytes, identifier)); err != nil {
		t.Fatalf("writing C-FIND-RQ: %v", err)
	}

	command, data, maxBody, pdus := readMessageWithMaxPDU(t, conn)
	if len(data) == 0 {
		t.Fatalf("pending response carried no identifier")
	}
	if got := refUS(t, refParseElements(t, command), TagStatus); got != StatusPending {
		t.Fatalf("status = 0x%04X, want pending", got)
	}
	if maxBody > 512 {
		t.Fatalf("largest PDU body = %d bytes, exceeds the negotiated 512 byte maximum", maxBody)
	}
	if pdus < 2 {
		t.Fatalf("response arrived in %d P-DATA-TF PDU(s); expected the server to fragment", pdus)
	}
	// The reassembled identifier must be complete and correct.
	ds, err := DecodeDataSet(data, false)
	if err != nil {
		t.Fatalf("DecodeDataSet: %v", err)
	}
	if v := ds.GetString(TagPatientName); v != big.PatientName {
		t.Errorf("PatientName length = %d, want %d", len(v), len(big.PatientName))
	}
	if v := ds.GetString(TagInstitutionName); v != big.InstitutionName {
		t.Errorf("InstitutionName did not survive fragmentation")
	}
	spsOut, ok := ds.GetSequence(TagScheduledProcedureStepSeq)
	if !ok {
		t.Fatal("SPS sequence missing after reassembly")
	}
	if v := spsOut.GetString(TagSPSStatus); v != "SCHEDULED" {
		t.Errorf("SPS status = %q", v)
	}

	// Drain the final response and release.
	statuses, _ := readCFindResponses(t, conn)
	if statuses[len(statuses)-1] != StatusSuccess {
		t.Fatalf("final status = 0x%04X", statuses[len(statuses)-1])
	}
	if _, err := conn.Write(buildReleaseRQ()); err != nil {
		t.Fatalf("writing A-RELEASE-RQ: %v", err)
	}
	if typ, _ := readRawPDU(t, conn); typ != PDUReleaseRP {
		t.Fatalf("expected A-RELEASE-RP, got 0x%02X", typ)
	}
}

func TestSCPMixedPresentationContextNegotiation(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	ts := startTestServer(t, mwlConfig(src))

	conn := dial(t, ts.addr)
	pcs := []refPC{
		{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}},
		{id: 3, abstract: "1.2.840.10008.5.1.4.1.1.2", transferSyntaxes: []string{ExplicitVRLittleEndian}}, // CT storage
		{id: 5, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{"1.2.840.10008.1.2.4.50"}},   // JPEG baseline
		{id: 7, abstract: UIDVerification, transferSyntaxes: []string{ImplicitVRLittleEndian}},
	}
	info, ctxID, _ := associate(t, conn, "FREEMED", pcs, 16384)

	if info.results[1] != PresContextAcceptance {
		t.Errorf("MWL context result = 0x%02X, want acceptance", info.results[1])
	}
	if info.results[3] != PresContextAbstractSyntaxUnsup {
		t.Errorf("unsupported SOP class result = 0x%02X, want 0x03", info.results[3])
	}
	if info.results[5] != PresContextTransferSyntaxUnsup {
		t.Errorf("unsupported transfer syntax result = 0x%02X, want 0x04", info.results[5])
	}
	if info.results[7] != PresContextAcceptance {
		t.Errorf("verification context result = 0x%02X, want acceptance", info.results[7])
	}
	if ctxID != 1 {
		t.Errorf("MWL context ID = %d, want 1", ctxID)
	}

	// C-ECHO on the verification context must succeed.
	echo, err := NewCFindResponse(0, "", 0, false).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	_ = echo
	req := NewCommandSet()
	req.SetUS(TagCommandField, CmdCEchoRQ)
	req.SetUS(TagMessageID, 99)
	req.SetUI(TagRequestedSOPClassUID, UIDVerification)
	req.SetUS(TagCommandDataSetType, DataSetTypeNone)
	reqBytes, err := req.Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := conn.Write(buildDataTF(7, reqBytes, nil)); err != nil {
		t.Fatalf("writing C-ECHO-RQ: %v", err)
	}
	command, _ := readMessage(t, conn)
	elems := refParseElements(t, command)
	if got := refUS(t, elems, TagCommandField); got != CmdCEchoRSP {
		t.Fatalf("command field = 0x%04X, want C-ECHO-RSP", got)
	}
	if got := refUS(t, elems, TagStatus); got != StatusSuccess {
		t.Fatalf("C-ECHO status = 0x%04X, want success", got)
	}
}

func TestSCPUnsupportedMatchingKeyStillReturnsItems(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	ts := startTestServer(t, mwlConfig(src))
	conn := dial(t, ts.addr)
	pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
	_, ctxID, _ := associate(t, conn, "FREEMED", pcs, 16384)

	// A matching key the model does not support must not empty the worklist.
	idDS := DataSet{
		NewStringElement(TagRequestedProcedurePriority, "SH", "HIGH"),
		NewStringElement(TagPatientName, "PN", "SMITH*"),
	}
	identifier, err := EncodeDataSet(idDS, false)
	if err != nil {
		t.Fatalf("EncodeDataSet: %v", err)
	}
	requestBytes, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := conn.Write(buildDataTF(ctxID, requestBytes, identifier)); err != nil {
		t.Fatalf("writing C-FIND-RQ: %v", err)
	}
	statuses, identifiers := readCFindResponses(t, conn)
	if statuses[len(statuses)-1] != StatusSuccess {
		t.Fatalf("final status = 0x%04X", statuses[len(statuses)-1])
	}
	// Three SMITH items match (the unsupported priority key must not filter
	// them out) plus the final success response.
	if len(statuses) != 4 {
		t.Fatalf("responses = %d, want 3 pending plus a final success", len(statuses))
	}
	pending := 0
	for i, st := range statuses[:len(statuses)-1] {
		if st != StatusPending || len(identifiers[i]) == 0 {
			t.Fatalf("response %d = 0x%04X with a %d byte identifier", i, st, len(identifiers[i]))
		}
		pending++
	}
	if pending != 3 {
		t.Fatalf("pending responses = %d, want 3", pending)
	}
}

func TestSCPSourceErrorReturnsFailureStatus(t *testing.T) {
	src := &fakeSource{err: errors.New("database is on fire")}
	ts := startTestServer(t, mwlConfig(src))
	conn := dial(t, ts.addr)
	pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
	_, ctxID, _ := associate(t, conn, "FREEMED", pcs, 16384)

	identifier := standardCFindIdentifier(t, false)
	requestBytes, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := conn.Write(buildDataTF(ctxID, requestBytes, identifier)); err != nil {
		t.Fatalf("writing C-FIND-RQ: %v", err)
	}
	statuses, _ := readCFindResponses(t, conn)
	if len(statuses) != 1 || statuses[0] != StatusRefusedResources {
		t.Fatalf("statuses = %v, want a single 0xA700", statuses)
	}
}

func TestSCPMalformedIdentifierReturnsCannotUnderstand(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	ts := startTestServer(t, mwlConfig(src))
	conn := dial(t, ts.addr)
	pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
	_, ctxID, _ := associate(t, conn, "FREEMED", pcs, 16384)

	// A truncated element header: not a valid data set.
	requestBytes, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	bad := []byte{0x10, 0x00, 0x10, 0x00, 0x40, 0x00, 0x00, 0x00, 'X'}
	if _, err := conn.Write(buildDataTF(ctxID, requestBytes, bad)); err != nil {
		t.Fatalf("writing C-FIND-RQ: %v", err)
	}
	statuses, _ := readCFindResponses(t, conn)
	if len(statuses) != 1 || statuses[0] != StatusCannotUnderstand {
		t.Fatalf("statuses = %v, want a single 0xC000", statuses)
	}
	if src.queryCount() != 0 {
		t.Error("the worklist source was queried for a malformed identifier")
	}
}

// ---------------------------------------------------------------------------
// Association rejection
// ---------------------------------------------------------------------------

func TestSCPRejectsUnknownAETitle(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	ts := startTestServer(t, mwlConfig(src))

	conn := dial(t, ts.addr)
	pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
	if _, err := conn.Write(buildAssociateRQ("NOT_FREEMED", "SCU_TEST", pcs, 16384)); err != nil {
		t.Fatalf("writing A-ASSOCIATE-RQ: %v", err)
	}
	typ, body := readRawPDU(t, conn)
	if typ != PDUAssociateRJ {
		t.Fatalf("expected A-ASSOCIATE-RJ (0x03), got PDU type 0x%02X", typ)
	}
	if len(body) != 4 {
		t.Fatalf("A-ASSOCIATE-RJ body is %d bytes, want 4", len(body))
	}
	if body[1] != RJResultPermanent {
		t.Errorf("result = 0x%02X, want permanent rejection", body[1])
	}
	if body[2] != RJSourceServiceUser {
		t.Errorf("source = 0x%02X, want service user", body[2])
	}
	if body[3] != RJReasonCalledAENotRecog {
		t.Errorf("reason = 0x%02X, want called-AE-title-not-recognized", body[3])
	}
	// The association must not be usable, and the connection must be closed.
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, err := conn.Read(make([]byte, 1)); err == nil {
		t.Error("connection stayed open after A-ASSOCIATE-RJ")
	}
}

func TestSCPRejectsUnsupportedSOPClass(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	ts := startTestServer(t, mwlConfig(src))

	conn := dial(t, ts.addr)
	pcs := []refPC{{id: 1, abstract: "1.2.840.10008.5.1.4.1.1.2", transferSyntaxes: []string{ExplicitVRLittleEndian}}}
	if _, err := conn.Write(buildAssociateRQ("FREEMED", "SCU_TEST", pcs, 16384)); err != nil {
		t.Fatalf("writing A-ASSOCIATE-RQ: %v", err)
	}
	typ, body := readRawPDU(t, conn)
	if typ != PDUAssociateRJ {
		t.Fatalf("expected A-ASSOCIATE-RJ for a proposal with no acceptable presentation context, got 0x%02X", typ)
	}
	if body[1] != RJResultPermanent {
		t.Errorf("result = 0x%02X, want permanent rejection", body[1])
	}
}

func TestSCPRejectsUnsupportedTransferSyntaxOnly(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	ts := startTestServer(t, mwlConfig(src))

	conn := dial(t, ts.addr)
	pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{"1.2.840.10008.1.2.4.50"}}}
	if _, err := conn.Write(buildAssociateRQ("FREEMED", "SCU_TEST", pcs, 16384)); err != nil {
		t.Fatalf("writing A-ASSOCIATE-RQ: %v", err)
	}
	typ, _ := readRawPDU(t, conn)
	if typ != PDUAssociateRJ {
		t.Fatalf("expected A-ASSOCIATE-RJ, got PDU type 0x%02X", typ)
	}
}

func TestSCPRejectsWrongProtocolVersionAndApplicationContext(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	ts := startTestServer(t, mwlConfig(src))

	pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}

	t.Run("protocol version 4", func(t *testing.T) {
		raw := buildAssociateRQ("FREEMED", "SCU_TEST", pcs, 16384)
		raw[6], raw[7] = 0x00, 0x04
		conn := dial(t, ts.addr)
		if _, err := conn.Write(raw); err != nil {
			t.Fatalf("write: %v", err)
		}
		typ, body := readRawPDU(t, conn)
		if typ != PDUAssociateRJ || body[3] != RJReasonProtocolVersionNotSup {
			t.Fatalf("PDU type 0x%02X reason 0x%02X, want RJ with reason 0x06", typ, body[3])
		}
	})

	t.Run("unknown application context", func(t *testing.T) {
		raw := buildAssociateRQ("FREEMED", "SCU_TEST", pcs, 16384)
		idx := bytes.Index(raw, []byte(UIDApplicationContext))
		if idx < 0 {
			t.Fatal("application context not found in the reference request")
		}
		copy(raw[idx:idx+len(UIDApplicationContext)], []byte("1.2.840.10008.3.1.1.9"))

		conn := dial(t, ts.addr)
		if _, err := conn.Write(raw); err != nil {
			t.Fatalf("write: %v", err)
		}
		typ, body := readRawPDU(t, conn)
		if typ != PDUAssociateRJ || body[3] != RJReasonAppContextNotSup {
			t.Fatalf("PDU type 0x%02X reason 0x%02X, want RJ with reason 0x02", typ, body[3])
		}
	})
}

// ---------------------------------------------------------------------------
// Robustness
// ---------------------------------------------------------------------------

func TestSCPSurvivesHostileInput(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	ts := startTestServer(t, mwlConfig(src))

	t.Run("garbage instead of an association request", func(t *testing.T) {
		conn := dial(t, ts.addr)
		if _, err := conn.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xDE, 0xAD}); err != nil {
			t.Fatalf("write: %v", err)
		}
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		buf := make([]byte, 16)
		n, err := conn.Read(buf)
		if err == nil && n > 0 && buf[0] != PDUAbort {
			t.Errorf("expected an A-ABORT PDU, got type 0x%02X", buf[0])
		}
	})

	t.Run("absurd PDU length", func(t *testing.T) {
		conn := dial(t, ts.addr)
		hdr := []byte{PDUAssociateRQ, 0x00, 0x7F, 0xFF, 0xFF, 0xFF}
		if _, err := conn.Write(hdr); err != nil {
			t.Fatalf("write: %v", err)
		}
		// The server must reject the length without allocating 2 GiB: it
		// either aborts or closes the connection. It must never wait for a
		// body that large.
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		buf := make([]byte, 8)
		n, err := conn.Read(buf)
		if err == nil && n > 0 && buf[0] != PDUAbort {
			t.Errorf("expected an A-ABORT or a closed connection, got PDU type 0x%02X", buf[0])
		}
	})

	t.Run("truncated association request", func(t *testing.T) {
		conn := dial(t, ts.addr)
		// A well formed header announcing 8 bytes of association PDU.
		if _, err := conn.Write([]byte{PDUAssociateRQ, 0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x01}); err != nil {
			t.Fatalf("write: %v", err)
		}
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, _ = conn.Read(make([]byte, 16))
	})

	t.Run("invalid command set on an established association", func(t *testing.T) {
		conn := dial(t, ts.addr)
		pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
		_, ctxID, _ := associate(t, conn, "FREEMED", pcs, 16384)
		// A PDV whose "command set" is a single byte.
		if _, err := conn.Write(buildDataTF(ctxID, []byte{0x01}, nil)); err != nil {
			t.Fatalf("write: %v", err)
		}
		typ, _ := readRawPDU(t, conn)
		if typ != PDUAbort {
			t.Errorf("expected A-ABORT after a malformed command set, got 0x%02X", typ)
		}
	})

	t.Run("PDV on an unaccepted presentation context", func(t *testing.T) {
		conn := dial(t, ts.addr)
		pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
		_, _, _ = associate(t, conn, "FREEMED", pcs, 16384)
		cmd, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
		if err != nil {
			t.Fatalf("Encode: %v", err)
		}
		// Context 9 was never accepted.
		if _, err := conn.Write(buildDataTF(9, cmd, nil)); err != nil {
			t.Fatalf("write: %v", err)
		}
		typ, _ := readRawPDU(t, conn)
		if typ != PDUAbort {
			t.Errorf("expected A-ABORT for a PDV on an unaccepted context, got 0x%02X", typ)
		}
	})

	t.Run("peer abort", func(t *testing.T) {
		conn := dial(t, ts.addr)
		pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
		_, _, _ = associate(t, conn, "FREEMED", pcs, 16384)
		if _, err := conn.Write(buildAbort()); err != nil {
			t.Fatalf("write: %v", err)
		}
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, _ = conn.Read(make([]byte, 1))
	})

	// After all of that the server must still serve a normal association.
	t.Run("server still healthy", func(t *testing.T) {
		conn := dial(t, ts.addr)
		pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
		_, ctxID, _ := associate(t, conn, "FREEMED", pcs, 16384)
		requestBytes, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
		if err != nil {
			t.Fatalf("Encode: %v", err)
		}
		if _, err := conn.Write(buildDataTF(ctxID, requestBytes, standardCFindIdentifier(t, false))); err != nil {
			t.Fatalf("write: %v", err)
		}
		statuses, _ := readCFindResponses(t, conn)
		if statuses[len(statuses)-1] != StatusSuccess {
			t.Fatalf("final status = 0x%04X", statuses[len(statuses)-1])
		}
	})
}

func TestSCPBoundsConcurrentAssociations(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	cfg := mwlConfig(src)
	cfg.MaxAssociations = 1
	ts := startTestServer(t, cfg)

	pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}

	first := dial(t, ts.addr)
	_, _, _ = associate(t, first, "FREEMED", pcs, 16384)

	// A second association must be refused (connection closed) while the first
	// is still active.
	second := dial(t, ts.addr)
	if _, err := second.Write(buildAssociateRQ("FREEMED", "SCU_TEST", pcs, 16384)); err != nil {
		t.Fatalf("write: %v", err)
	}
	_ = second.SetReadDeadline(time.Now().Add(3 * time.Second))
	if _, err := second.Read(make([]byte, 1)); err == nil {
		t.Error("the server served a second association although MaxAssociations is 1")
	}

	// Releasing the first association frees the slot.
	if _, err := first.Write(buildReleaseRQ()); err != nil {
		t.Fatalf("write: %v", err)
	}
	if typ, _ := readRawPDU(t, first); typ != PDUReleaseRP {
		t.Fatalf("expected A-RELEASE-RP, got 0x%02X", typ)
	}
	_ = first.Close()

	deadline := time.Now().Add(5 * time.Second)
	for {
		third := dial(t, ts.addr)
		if _, err := third.Write(buildAssociateRQ("FREEMED", "SCU_TEST", pcs, 16384)); err != nil {
			t.Fatalf("write: %v", err)
		}
		if typ, _ := readRawPDU(t, third); typ == PDUAssociateAC {
			return
		}
		_ = third.Close()
		if time.Now().After(deadline) {
			t.Fatal("the association slot was never released")
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func TestSCPShutdownOnContextCancel(t *testing.T) {
	baseline := runtime.NumGoroutine()

	src := &fakeSource{items: sourceItems()}
	cfg := mwlConfig(src)
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ctx, ln) }()
	addr := ln.Addr().String()

	pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
	for i := 0; i < 5; i++ {
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
		if _, err := conn.Write(buildAssociateRQ("FREEMED", "SCU_TEST", pcs, 16384)); err != nil {
			t.Fatalf("write RQ: %v", err)
		}
		typ, _ := readRawPDU(t, conn)
		if typ != PDUAssociateAC {
			t.Fatalf("association %d rejected (PDU type 0x%02X)", i, typ)
		}
		requestBytes, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
		if err != nil {
			t.Fatalf("Encode: %v", err)
		}
		if _, err := conn.Write(buildDataTF(1, requestBytes, standardCFindIdentifier(t, false))); err != nil {
			t.Fatalf("write C-FIND-RQ: %v", err)
		}
		statuses, _ := readCFindResponses(t, conn)
		if statuses[len(statuses)-1] != StatusSuccess {
			t.Fatalf("association %d final status = 0x%04X", i, statuses[len(statuses)-1])
		}
		if _, err := conn.Write(buildReleaseRQ()); err != nil {
			t.Fatalf("write release: %v", err)
		}
		if typ, _ := readRawPDU(t, conn); typ != PDUReleaseRP {
			t.Fatalf("association %d: expected A-RELEASE-RP, got 0x%02X", i, typ)
		}
		_ = conn.Close()
	}

	// Cancelling the context must stop the server.
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Serve returned %v after context cancellation, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return after the context was cancelled")
	}

	// Shutdown is idempotent and must not hang.
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}

	// No goroutine leak: the count must settle back to the baseline.
	deadline := time.Now().Add(5 * time.Second)
	for {
		runtime.GC()
		now := runtime.NumGoroutine()
		if now <= baseline+2 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("goroutines after shutdown = %d, baseline = %d", now, baseline)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestSCPShutdownClosesActiveAssociations(t *testing.T) {
	src := &fakeSource{items: sourceItems(), block: make(chan struct{})}
	ts := startTestServer(t, mwlConfig(src))

	conn := dial(t, ts.addr)
	pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
	_, ctxID, _ := associate(t, conn, "FREEMED", pcs, 16384)

	requestBytes, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := conn.Write(buildDataTF(ctxID, requestBytes, standardCFindIdentifier(t, false))); err != nil {
		t.Fatalf("write C-FIND-RQ: %v", err)
	}
	// Wait until the handler is actually inside the worklist query.
	deadline := time.Now().Add(5 * time.Second)
	for src.queryCount() == 0 {
		if time.Now().After(deadline) {
			t.Fatal("the SCP never queried the worklist source")
		}
		time.Sleep(10 * time.Millisecond)
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := ts.srv.Shutdown(shutCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	close(src.block)

	// Shutdown must close the active association. The handler may have raced a
	// final failure response onto the wire before the connection was closed,
	// so drain until the server actually closes the socket: a read timeout
	// would mean the connection was left open.
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 4096)
	for {
		if _, err := conn.Read(buf); err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				t.Fatal("the active association was left open by Shutdown")
			}
			return // EOF or connection reset: the server closed it
		}
	}
}

// ---------------------------------------------------------------------------
// Configuration validation
// ---------------------------------------------------------------------------

func TestValidateAETitle(t *testing.T) {
	valid := []string{"FREEMED", "CT1", "A", "ABCDEFGHIJKLMNOP", "A B C"}
	for _, s := range valid {
		if err := ValidateAETitle(s); err != nil {
			t.Errorf("ValidateAETitle(%q) = %v, want nil", s, err)
		}
	}
	invalid := []string{
		"",
		"ABCDEFGHIJKLMNOPQ", // 17 characters
		strings.Repeat("A", 100),
		" FREEMED",
		"FREEMED ",
		"FREE\nMED",
		"FREE/MED",
		"FREE\\MED",
		"FREE:MED",
		"FREE*MED",
		"FREE?MED",
		"FREE\x00MED",
		"FREE\x7fMED",
	}
	for _, s := range invalid {
		if err := ValidateAETitle(s); !errors.Is(err, ErrInvalidAETitle) {
			t.Errorf("ValidateAETitle(%q) = %v, want ErrInvalidAETitle", s, err)
		}
	}
}

func TestNewServerValidationAndDefaults(t *testing.T) {
	src := &fakeSource{}
	if _, err := NewServer(Config{AETitle: "", Source: src}); !errors.Is(err, ErrInvalidAETitle) {
		t.Errorf("empty AE title: err = %v", err)
	}
	if _, err := NewServer(Config{AETitle: strings.Repeat("X", 17), Source: src}); !errors.Is(err, ErrInvalidAETitle) {
		t.Errorf("oversized AE title: err = %v", err)
	}
	if _, err := NewServer(Config{AETitle: "FREEMED"}); !errors.Is(err, ErrNoWorklistSource) {
		t.Errorf("nil source: err = %v", err)
	}

	srv, err := NewServer(Config{AETitle: "FREEMED", Source: src})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	if srv.cfg.MaxAssociations != DefaultMaxAssociations {
		t.Errorf("MaxAssociations = %d, want %d", srv.cfg.MaxAssociations, DefaultMaxAssociations)
	}
	if srv.cfg.MaxPDULength != DefaultMaxPDULength {
		t.Errorf("MaxPDULength = %d, want %d", srv.cfg.MaxPDULength, DefaultMaxPDULength)
	}
	if srv.cfg.IdleTimeout != DefaultIdleTimeout {
		t.Errorf("IdleTimeout = %v", srv.cfg.IdleTimeout)
	}
	if srv.cfg.Logf == nil {
		t.Error("Logf default not applied")
	}
	if srv.AETitle() != "FREEMED" {
		t.Errorf("AETitle() = %q", srv.AETitle())
	}
	// Addr is nil before serving, and Shutdown before serving is a no-op.
	if srv.Addr() != nil {
		t.Errorf("Addr() = %v before Serve", srv.Addr())
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown before Serve: %v", err)
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Errorf("second Shutdown: %v", err)
	}
}

func TestServerListenAndServeAndShutdown(t *testing.T) {
	src := &fakeSource{items: sourceItems()}
	cfg := mwlConfig(src)
	cfg.ListenAddr = "127.0.0.1:0"
	srv, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe(ctx) }()

	var addr string
	deadline := time.Now().Add(5 * time.Second)
	for srv.Addr() == nil {
		if time.Now().After(deadline) {
			t.Fatal("ListenAndServe never bound a listener")
		}
		time.Sleep(10 * time.Millisecond)
	}
	addr = srv.Addr().String()

	conn := dial(t, addr)
	pcs := []refPC{{id: 1, abstract: UIDModalityWorklistFIND, transferSyntaxes: []string{ImplicitVRLittleEndian}}}
	_, ctxID, _ := associate(t, conn, "FREEMED", pcs, 16384)
	requestBytes, err := NewCFindRequest(1, UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if _, err := conn.Write(buildDataTF(ctxID, requestBytes, standardCFindIdentifier(t, false))); err != nil {
		t.Fatalf("write: %v", err)
	}
	statuses, _ := readCFindResponses(t, conn)
	if statuses[len(statuses)-1] != StatusSuccess {
		t.Fatalf("final status = 0x%04X", statuses[len(statuses)-1])
	}

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("ListenAndServe returned %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ListenAndServe did not return after Shutdown")
	}

	// New connections must be refused after shutdown.
	if _, err := net.DialTimeout("tcp", addr, time.Second); err == nil {
		t.Error("the listener still accepts connections after Shutdown")
	}
}

func TestChooseTransferSyntax(t *testing.T) {
	tests := []struct {
		proposed []string
		want     string
		ok       bool
	}{
		{[]string{ImplicitVRLittleEndian}, ImplicitVRLittleEndian, true},
		{[]string{ExplicitVRLittleEndian}, ExplicitVRLittleEndian, true},
		{[]string{ExplicitVRLittleEndian, ImplicitVRLittleEndian}, ImplicitVRLittleEndian, true},
		{[]string{"1.2.840.10008.1.2.4.50", ExplicitVRLittleEndian}, ExplicitVRLittleEndian, true},
		{[]string{"1.2.840.10008.1.2.4.50"}, "", false},
		{nil, "", false},
		{[]string{" "}, "", false},
	}
	for _, tc := range tests {
		got, ok := chooseTransferSyntax(tc.proposed, 0)
		if got != tc.want || ok != tc.ok {
			t.Errorf("chooseTransferSyntax(%v) = %q/%v, want %q/%v", tc.proposed, got, ok, tc.want, tc.ok)
		}
	}
}

func TestFirstContextID(t *testing.T) {
	if got := firstContextID(map[byte]string{3: ImplicitVRLittleEndian, 1: ImplicitVRLittleEndian, 5: ImplicitVRLittleEndian}); got != 1 {
		t.Errorf("firstContextID = %d, want 1", got)
	}
	if got := firstContextID(nil); got != 0 {
		t.Errorf("firstContextID(nil) = %d, want 0", got)
	}
}

func TestSCPErrorWhenBusy(t *testing.T) {
	// ErrBusy documents the refusal path exercised above; keep the sentinel
	// meaningful.
	if ErrBusy == nil || !strings.Contains(ErrBusy.Error(), "association") {
		t.Errorf("ErrBusy = %v", ErrBusy)
	}
	if !strings.Contains(fmt.Sprint(ErrServerClosed), "closed") {
		t.Errorf("ErrServerClosed = %v", ErrServerClosed)
	}
}
