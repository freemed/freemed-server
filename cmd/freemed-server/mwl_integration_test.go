//go:build integration

package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/freemed/freemed-server/pkg/dicomnet"
)

// TestMwlEndToEndAgainstDatabase drives the real database backed worklist
// source through the real C-FIND SCP over TCP. It requires a running MySQL
// (docker compose up -d db) with the freemed schema and at least one
// appointment in August 2026, or it is skipped.
//
// Run with: go test -tags=integration -run TestMwlEndToEnd -v ./...
func TestMwlEndToEndAgainstDatabase(t *testing.T) {
	db, err := sql.Open("mysql", "freemed:freemed@tcp(127.0.0.1:3306)/freemed?parseTime=true")
	if err != nil {
		t.Skipf("cannot open the freemed database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Skipf("freemed database is not reachable: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// A window that contains the seeded appointments.
	queryDS := dicomnet.DataSet{
		dicomnet.NewStringElement(dicomnet.TagPatientName, "PN", ""),
		dicomnet.NewStringElement(dicomnet.TagPatientID, "LO", ""),
		dicomnet.NewStringElement(dicomnet.TagStudyInstanceUID, "UI", ""),
		dicomnet.NewStringElement(dicomnet.TagAccessionNumber, "SH", ""),
		dicomnet.NewStringElement(dicomnet.TagRequestedProcedureID, "SH", ""),
		dicomnet.NewSequenceElement(dicomnet.TagScheduledProcedureStepSeq, dicomnet.DataSet{
			dicomnet.NewStringElement(dicomnet.TagModality, "CS", ""),
			dicomnet.NewStringElement(dicomnet.TagScheduledStationAETitle, "AE", ""),
			dicomnet.NewStringElement(dicomnet.TagSPSStartDate, "DA", "20260801-20260831"),
			dicomnet.NewStringElement(dicomnet.TagSPSStartTime, "TM", ""),
			dicomnet.NewStringElement(dicomnet.TagScheduledPerformingPhysName, "PN", ""),
			dicomnet.NewStringElement(dicomnet.TagSPSID, "SH", ""),
			dicomnet.NewStringElement(dicomnet.TagSPSStatus, "CS", ""),
		}),
	}
	q, err := dicomnet.ParseMWLQuery(queryDS)
	if err != nil {
		t.Fatalf("ParseMWLQuery: %v", err)
	}

	src := newMwlSource(db, "FREEMED")
	items, err := src.Find(ctx, q)
	if err != nil {
		t.Fatalf("mwlSource.Find against the real database: %v", err)
	}
	if len(items) == 0 {
		t.Skip("no appointments in August 2026 in the development database; nothing to verify")
	}
	t.Logf("worklist source returned %d appointment(s) for 20260801-20260831", len(items))
	for _, it := range items {
		t.Logf("  %s | %s | %s %s | %s | %s",
			it.PatientName, it.PatientID, it.ScheduledProcedureStepStartDate,
			it.ScheduledProcedureStepStartTime, it.Modality, it.ScheduledProcedureStepStatus)
		if it.PatientName == "" {
			t.Error("an item has an empty PatientName")
		}
		if it.PatientID == "" {
			t.Error("an item has an empty PatientID")
		}
		if it.StudyInstanceUID == "" {
			t.Error("an item has an empty StudyInstanceUID")
		}
		if it.ScheduledProcedureStepStartDate < "20260801" || it.ScheduledProcedureStepStartDate > "20260831" {
			t.Errorf("item outside the requested window: %s", it.ScheduledProcedureStepStartDate)
		}
		if it.ScheduledStationAETitle != "FREEMED" {
			t.Errorf("ScheduledStationAETitle = %q", it.ScheduledStationAETitle)
		}
	}

	// Now serve it for real: start the SCP with the database backed source and
	// perform a C-FIND as a client.
	srv, err := dicomnet.NewServer(dicomnet.Config{
		AETitle:          "FREEMED",
		MaxAssociations:  2,
		HandshakeTimeout: 5 * time.Second,
		IdleTimeout:      10 * time.Second,
		Source:           src,
		Logf:             t.Logf,
	})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srvCtx, srvCancel := context.WithCancel(context.Background())
	defer srvCancel()
	go func() { _ = srv.Serve(srvCtx, ln) }()
	defer func() {
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		_ = srv.Shutdown(shutCtx)
	}()

	conn, err := net.DialTimeout("tcp", ln.Addr().String(), 5*time.Second)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))

	rq := dicomnet.NewAssociateRQ("FREEMED", "INTEGRATION", "", []dicomnet.PresentationContext{{
		ID:               1,
		AbstractSyntax:   dicomnet.UIDModalityWorklistFIND,
		TransferSyntaxes: []string{dicomnet.ImplicitVRLittleEndian},
	}}, 16384, "", "")
	rqBytes, err := rq.Encode()
	if err != nil {
		t.Fatalf("encode A-ASSOCIATE-RQ: %v", err)
	}
	if _, err := conn.Write(rqBytes); err != nil {
		t.Fatalf("write A-ASSOCIATE-RQ: %v", err)
	}
	ac, err := dicomnet.ReadPDU(conn, dicomnet.MaxPDUBodySize)
	if err != nil {
		t.Fatalf("read A-ASSOCIATE-AC: %v", err)
	}
	if ac.Type != dicomnet.PDUAssociateAC {
		t.Fatalf("association rejected: PDU type 0x%02X", ac.Type)
	}
	accepted := false
	for _, pc := range ac.PresentationContexts {
		if pc.ID == 1 && pc.Result == dicomnet.PresContextAcceptance {
			accepted = true
		}
	}
	if !accepted {
		t.Fatalf("the MWL presentation context was not accepted: %+v", ac.PresentationContexts)
	}

	identifier, err := dicomnet.EncodeDataSet(queryDS, false)
	if err != nil {
		t.Fatalf("encode identifier: %v", err)
	}
	cmdBytes, err := dicomnet.NewCFindRequest(1, dicomnet.UIDModalityWorklistFIND, 0).Encode()
	if err != nil {
		t.Fatalf("encode C-FIND-RQ: %v", err)
	}
	pdvs, err := dicomnet.Fragment(1, cmdBytes, identifier, ac.MaxPDULength)
	if err != nil {
		t.Fatalf("fragment: %v", err)
	}
	for _, pdv := range pdvs {
		b, err := dicomnet.NewDataTF(pdv).Encode()
		if err != nil {
			t.Fatalf("encode P-DATA-TF: %v", err)
		}
		if _, err := conn.Write(b); err != nil {
			t.Fatalf("write P-DATA-TF: %v", err)
		}
	}

	acc := dicomnet.NewMessageAccumulator()
	pending := 0
	var firstResponse dicomnet.DataSet
	for {
		pdu, err := dicomnet.ReadPDU(conn, dicomnet.MaxPDUBodySize)
		if err != nil {
			t.Fatalf("read C-FIND response: %v", err)
		}
		if pdu.Type != dicomnet.PDUDataTF {
			t.Fatalf("expected P-DATA-TF, got PDU type 0x%02X", pdu.Type)
		}
		for _, pdv := range pdu.PDVs {
			if err := acc.Feed(pdv); err != nil {
				t.Fatalf("assembling response: %v", err)
			}
		}
		if !acc.Complete() {
			continue
		}
		cs := acc.Command()
		if cs == nil {
			t.Fatal("response command set did not parse")
		}
		status, _ := cs.Status()
		switch status {
		case dicomnet.StatusPending:
			pending++
			ds, err := dicomnet.DecodeDataSet(acc.DataSetBytes(), false)
			if err != nil {
				t.Fatalf("decoding response identifier: %v", err)
			}
			if pending == 1 {
				firstResponse = ds
			}
			acc = dicomnet.NewMessageAccumulator()
			continue
		case dicomnet.StatusSuccess:
			if pending == 0 {
				t.Fatal("the SCP returned no matching worklist items")
			}
			fmt.Printf("integration: %d worklist item(s) returned over the wire\n", pending)
			if v := firstResponse.GetString(dicomnet.TagPatientID); v == "" {
				t.Error("the first response has an empty PatientID")
			}
			if sps, ok := firstResponse.GetSequence(dicomnet.TagScheduledProcedureStepSeq); ok {
				if v := sps.GetString(dicomnet.TagSPSStartDate); v < "20260801" || v > "20260831" {
					t.Errorf("response SPS start date outside the window: %q", v)
				}
			} else {
				t.Error("the first response has no ScheduledProcedureStepSequence")
			}
			return
		default:
			t.Fatalf("unexpected C-FIND status 0x%04X", status)
		}
	}
}
