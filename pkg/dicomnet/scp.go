package dicomnet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// Implementation identification sent in the user information item. The root
// is a placeholder sample root; replace it with the site's registered UID
// before deploying to production (it is only used for association diagnostics).
const (
	ImplementationClassUID    = "1.2.826.0.1.3680043.8.498.1"
	ImplementationVersionName = "FREEMED_MWL_1"
)

// Defaults applied by Config.applyDefaults.
const (
	DefaultAETitle          = "FREEMED"
	DefaultMaxAssociations  = 16
	DefaultIdleTimeout      = 5 * time.Minute
	DefaultHandshakeTimeout = 30 * time.Second
	DefaultShutdownGrace    = 5 * time.Second
)

// Errors returned by the SCP.
var (
	ErrNoWorklistSource = errors.New("dicomnet: no worklist source configured")
	ErrServerClosed     = errors.New("dicomnet: server closed")
	ErrBusy             = errors.New("dicomnet: maximum number of associations reached")
)

// WorklistSource supplies Modality Worklist entries for a parsed C-FIND query.
// Implementations must be safe for concurrent use and must honour ctx.
type WorklistSource interface {
	Find(ctx context.Context, q *MWLQuery) ([]MWLItem, error)
}

// Config configures a Modality Worklist C-FIND SCP.
type Config struct {
	// ListenAddr is the TCP address to listen on, e.g. "0.0.0.0:11112".
	ListenAddr string

	// AETitle is the AE title this SCP answers to. Requests addressed to any
	// other AE title are rejected with A-ASSOCIATE-RJ.
	AETitle string

	// MaxAssociations bounds the number of simultaneously served
	// associations. Excess connections are closed immediately.
	MaxAssociations int

	// MaxPDULength is the maximum PDU length advertised to peers.
	MaxPDULength uint32

	// IdleTimeout bounds the wait for the next PDU on an established
	// association.
	IdleTimeout time.Duration

	// HandshakeTimeout bounds the association negotiation exchange.
	HandshakeTimeout time.Duration

	// Source supplies the worklist contents.
	Source WorklistSource

	// Logf, when set, receives diagnostic messages. Defaults to log.Printf.
	Logf func(format string, args ...any)
}

func (c *Config) applyDefaults() {
	if c.AETitle == "" {
		c.AETitle = DefaultAETitle
	}
	if c.MaxAssociations <= 0 {
		c.MaxAssociations = DefaultMaxAssociations
	}
	if c.MaxPDULength == 0 || c.MaxPDULength > MaxPDUBodySize {
		c.MaxPDULength = DefaultMaxPDULength
	}
	if c.IdleTimeout <= 0 {
		c.IdleTimeout = DefaultIdleTimeout
	}
	if c.HandshakeTimeout <= 0 {
		c.HandshakeTimeout = DefaultHandshakeTimeout
	}
	if c.Logf == nil {
		c.Logf = log.Printf
	}
}

// ValidateAETitle checks an AE title: 1..16 printable characters, no leading
// or trailing spaces, no control characters and no path separators. The AE
// title is never used to build a SQL statement or a filesystem path.
func ValidateAETitle(s string) error {
	if s == "" {
		return fmt.Errorf("%w: empty", ErrInvalidAETitle)
	}
	if len(s) > MaxAETitleLength {
		return fmt.Errorf("%w: %q is %d characters (max %d)", ErrInvalidAETitle, s, len(s), MaxAETitleLength)
	}
	if strings.TrimSpace(s) != s {
		return fmt.Errorf("%w: %q has leading or trailing whitespace", ErrInvalidAETitle, s)
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c > 0x7E {
			return fmt.Errorf("%w: %q contains a non-printable character", ErrInvalidAETitle, s)
		}
		switch c {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return fmt.Errorf("%w: %q contains a reserved character %q", ErrInvalidAETitle, s, string(c))
		}
	}
	return nil
}

// association is the per-connection negotiated state.
type association struct {
	conn      net.Conn
	callingAE string
	calledAE  string
	peerMax   uint32
	// accepted maps presentation context ID -> transfer syntax UID.
	accepted map[byte]string
}

// Server is a Modality Worklist C-FIND SCP.
type Server struct {
	cfg Config

	mu       sync.Mutex
	ln       net.Listener
	conns    map[net.Conn]struct{}
	sem      chan struct{}
	closed   bool
	done     chan struct{}
	wg       sync.WaitGroup
	rejected uint64

	// baseCtx is cancelled by Shutdown so that in-flight worklist queries
	// stop instead of blocking the graceful shutdown.
	baseCtx context.Context
	cancel  context.CancelFunc
}

// NewServer validates cfg and returns a server ready to ListenAndServe.
func NewServer(cfg Config) (*Server, error) {
	if err := ValidateAETitle(cfg.AETitle); err != nil {
		return nil, err
	}
	if cfg.Source == nil {
		return nil, ErrNoWorklistSource
	}
	cfg.applyDefaults()
	baseCtx, cancel := context.WithCancel(context.Background())
	return &Server{
		cfg:     cfg,
		conns:   make(map[net.Conn]struct{}),
		sem:     make(chan struct{}, cfg.MaxAssociations),
		done:    make(chan struct{}),
		baseCtx: baseCtx,
		cancel:  cancel,
	}, nil
}

// Addr returns the address the server is listening on, or nil before
// ListenAndServe has bound a listener.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ln == nil {
		return nil
	}
	return s.ln.Addr()
}

// AETitle returns the configured AE title.
func (s *Server) AETitle() string { return s.cfg.AETitle }

// ListenAndServe binds the configured address and serves until ctx is
// cancelled or Shutdown is called.
func (s *Server) ListenAndServe(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return err
	}
	return s.Serve(ctx, ln)
}

// Serve serves an already bound listener.
func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		_ = ln.Close()
		return ErrServerClosed
	}
	s.ln = ln
	s.mu.Unlock()

	s.cfg.Logf("dicomnet: MWL SCP listening on %s as AE %q", ln.Addr(), s.cfg.AETitle)

	go func() {
		select {
		case <-ctx.Done():
			shutCtx, cancel := context.WithTimeout(context.Background(), DefaultShutdownGrace)
			defer cancel()
			_ = s.Shutdown(shutCtx)
		case <-s.doneChan():
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			s.mu.Lock()
			closed := s.closed
			s.mu.Unlock()
			if closed || errors.Is(err, net.ErrClosed) {
				return nil
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			return err
		}
		select {
		case s.sem <- struct{}{}:
		default:
			// Connection bound reached: refuse without spawning anything.
			s.mu.Lock()
			s.rejected++
			s.mu.Unlock()
			s.cfg.Logf("dicomnet: refusing association from %s: %d associations already active", conn.RemoteAddr(), s.cfg.MaxAssociations)
			_ = conn.Close()
			continue
		}

		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			<-s.sem
			_ = conn.Close()
			return nil
		}
		s.conns[conn] = struct{}{}
		s.mu.Unlock()

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			defer func() {
				<-s.sem
				s.mu.Lock()
				delete(s.conns, conn)
				s.mu.Unlock()
				_ = conn.Close()
			}()
			s.serveConn(ctx, conn)
		}()
	}
}

// doneChan is closed once Shutdown has been requested.
func (s *Server) doneChan() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.done
}

// Shutdown stops accepting new associations, closes active connections and
// waits for the handlers to return, honouring ctx.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	if s.done == nil {
		s.done = make(chan struct{})
	}
	close(s.done)
	ln := s.ln
	conns := make([]net.Conn, 0, len(s.conns))
	for c := range s.conns {
		conns = append(conns, c)
	}
	s.mu.Unlock()

	// Stop in-flight worklist queries so handlers cannot block the shutdown.
	s.cancel()

	if ln != nil {
		_ = ln.Close()
	}
	for _, c := range conns {
		_ = c.Close()
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// serveConn negotiates and then services one association.
func (s *Server) serveConn(ctx context.Context, conn net.Conn) {
	if err := conn.SetDeadline(time.Now().Add(s.cfg.HandshakeTimeout)); err != nil {
		return
	}

	pdu, err := ReadPDU(conn, MaxPDUBodySize)
	if err != nil {
		s.cfg.Logf("dicomnet: %s: reading association request: %v", conn.RemoteAddr(), err)
		_ = s.sendPDU(conn, NewAbort(AbortSourceServiceProvider, 0x00), time.Now().Add(5*time.Second))
		return
	}
	if pdu.Type != PDUAssociateRQ {
		s.cfg.Logf("dicomnet: %s: expected A-ASSOCIATE-RQ, got PDU type 0x%02X", conn.RemoteAddr(), pdu.Type)
		_ = s.sendPDU(conn, NewAbort(AbortSourceServiceProvider, 0x02), time.Now().Add(5*time.Second))
		return
	}

	assoc, err := s.negotiate(conn, pdu)
	if err != nil {
		s.cfg.Logf("dicomnet: %s: association rejected: %v", conn.RemoteAddr(), err)
		return
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		return
	}
	s.cfg.Logf("dicomnet: association accepted from AE %q (%s), %d presentation context(s)", assoc.callingAE, conn.RemoteAddr(), len(assoc.accepted))
	s.serviceAssociation(ctx, assoc)
}

// negotiate validates the association request and answers with an AC or RJ.
func (s *Server) negotiate(conn net.Conn, rq *PDU) (*association, error) {
	if rq.ProtocolVersion != 1 {
		return nil, s.reject(conn, RJResultPermanent, RJSourceServiceUser, RJReasonProtocolVersionNotSup,
			fmt.Sprintf("protocol version %d not supported", rq.ProtocolVersion))
	}
	if rq.ApplicationContext != "" && rq.ApplicationContext != UIDApplicationContext {
		return nil, s.reject(conn, RJResultPermanent, RJSourceServiceUser, RJReasonAppContextNotSup,
			fmt.Sprintf("application context %q not supported", rq.ApplicationContext))
	}
	if rq.CalledAETitle == "" {
		return nil, s.reject(conn, RJResultPermanent, RJSourceServiceUser, RJReasonCalledAENotRecog, "empty called AE title")
	}
	if err := ValidateAETitle(rq.CalledAETitle); err != nil {
		return nil, s.reject(conn, RJResultPermanent, RJSourceServiceUser, RJReasonCalledAENotRecog, err.Error())
	}
	if !strings.EqualFold(strings.TrimSpace(rq.CalledAETitle), s.cfg.AETitle) {
		return nil, s.reject(conn, RJResultPermanent, RJSourceServiceUser, RJReasonCalledAENotRecog,
			fmt.Sprintf("called AE title %q is not %q", rq.CalledAETitle, s.cfg.AETitle))
	}

	accepted := make(map[byte]string)
	var acPCs []PresentationContext
	for _, pc := range rq.PresentationContexts {
		out := PresentationContext{ID: pc.ID}
		switch pc.AbstractSyntax {
		case UIDModalityWorklistFIND, UIDVerification:
			if ts, ok := chooseTransferSyntax(pc.TransferSyntaxes, s.cfg.MaxPDULength); ok {
				out.Result = PresContextAcceptance
				out.TransferSyntaxes = []string{ts}
				accepted[pc.ID] = ts
			} else {
				out.Result = PresContextTransferSyntaxUnsup
				out.TransferSyntaxes = []string{ImplicitVRLittleEndian}
			}
		default:
			out.Result = PresContextAbstractSyntaxUnsup
			out.TransferSyntaxes = []string{ImplicitVRLittleEndian}
		}
		acPCs = append(acPCs, out)
	}

	if len(accepted) == 0 {
		return nil, s.reject(conn, RJResultPermanent, RJSourceServiceUser, RJReasonNoReason,
			"no acceptable presentation context proposed")
	}

	ac := &PDU{
		Type:                      PDUAssociateAC,
		ProtocolVersion:           1,
		CalledAETitle:             rq.CalledAETitle,
		CallingAETitle:            rq.CallingAETitle,
		ApplicationContext:        UIDApplicationContext,
		PresentationContexts:      acPCs,
		MaxPDULength:              s.cfg.MaxPDULength,
		ImplementationClassUID:    ImplementationClassUID,
		ImplementationVersionName: ImplementationVersionName,
	}
	if err := s.sendPDU(conn, ac, time.Now().Add(s.cfg.HandshakeTimeout)); err != nil {
		return nil, err
	}
	return &association{
		conn:      conn,
		callingAE: rq.CallingAETitle,
		calledAE:  rq.CalledAETitle,
		peerMax:   rq.MaxPDULength,
		accepted:  accepted,
	}, nil
}

// reject sends an A-ASSOCIATE-RJ PDU and returns the reason as an error.
func (s *Server) reject(conn net.Conn, result, source, reason byte, msg string) error {
	_ = s.sendPDU(conn, NewAssociateRJ(result, source, reason), time.Now().Add(5*time.Second))
	return fmt.Errorf("%s (result 0x%02X source 0x%02X reason 0x%02X)", msg, result, source, reason)
}

// chooseTransferSyntax picks a supported transfer syntax from those proposed.
// Implicit VR Little Endian is preferred, then Explicit VR Little Endian.
func chooseTransferSyntax(proposed []string, _ uint32) (string, bool) {
	var haveImplicit, haveExplicit bool
	for _, ts := range proposed {
		switch strings.TrimSpace(ts) {
		case ImplicitVRLittleEndian:
			haveImplicit = true
		case ExplicitVRLittleEndian:
			haveExplicit = true
		}
	}
	switch {
	case haveImplicit:
		return ImplicitVRLittleEndian, true
	case haveExplicit:
		return ExplicitVRLittleEndian, true
	}
	return "", false
}

// serviceAssociation reads PDUs and dispatches DIMSE messages until the peer
// releases, aborts, or a protocol error occurs.
func (s *Server) serviceAssociation(parent context.Context, a *association) {
	// The association context is cancelled when either the caller's context
	// ends (the Serve watcher then calls Shutdown) or Shutdown itself runs, so
	// a C-FIND in progress is abandoned instead of blocking shutdown.
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	go func() {
		select {
		case <-s.baseCtx.Done():
			cancel()
		case <-ctx.Done():
		}
	}()

	acc := NewMessageAccumulator()
	for {
		if err := a.conn.SetReadDeadline(time.Now().Add(s.cfg.IdleTimeout)); err != nil {
			return
		}
		pdu, err := ReadPDU(a.conn, MaxPDUBodySize)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				s.cfg.Logf("dicomnet: %s: %v", a.conn.RemoteAddr(), err)
			}
			return
		}

		switch pdu.Type {
		case PDUDataTF:
			for _, pdv := range pdu.PDVs {
				if _, ok := a.accepted[pdv.ContextID]; !ok {
					s.cfg.Logf("dicomnet: %s: PDV on unaccepted presentation context %d", a.conn.RemoteAddr(), pdv.ContextID)
					_ = s.sendPDU(a.conn, NewAbort(AbortSourceServiceProvider, 0x02), time.Now().Add(5*time.Second))
					return
				}
				if err := acc.Feed(pdv); err != nil {
					s.cfg.Logf("dicomnet: %s: message assembly: %v", a.conn.RemoteAddr(), err)
					_ = s.sendPDU(a.conn, NewAbort(AbortSourceServiceProvider, 0x02), time.Now().Add(5*time.Second))
					return
				}
			}
			if !acc.Complete() {
				continue
			}
			cmd := acc.Command()
			if cmd == nil {
				_ = s.sendPDU(a.conn, NewAbort(AbortSourceServiceProvider, 0x02), time.Now().Add(5*time.Second))
				return
			}
			if err := s.dispatch(ctx, a, cmd, acc.DataSetBytes()); err != nil {
				s.cfg.Logf("dicomnet: %s: %v", a.conn.RemoteAddr(), err)
				return
			}
			acc = NewMessageAccumulator()

		case PDUReleaseRQ:
			if err := s.sendPDU(a.conn, NewReleaseRP(), time.Now().Add(5*time.Second)); err != nil {
				s.cfg.Logf("dicomnet: %s: sending release response: %v", a.conn.RemoteAddr(), err)
			}
			return

		case PDUAbort:
			s.cfg.Logf("dicomnet: %s: peer aborted the association", a.conn.RemoteAddr())
			return

		case PDUAssociateRQ:
			s.cfg.Logf("dicomnet: %s: unexpected A-ASSOCIATE-RQ on an established association", a.conn.RemoteAddr())
			_ = s.sendPDU(a.conn, NewAbort(AbortSourceServiceProvider, 0x02), time.Now().Add(5*time.Second))
			return

		default:
			s.cfg.Logf("dicomnet: %s: unexpected PDU type 0x%02X", a.conn.RemoteAddr(), pdu.Type)
			_ = s.sendPDU(a.conn, NewAbort(AbortSourceServiceProvider, 0x02), time.Now().Add(5*time.Second))
			return
		}
	}
}

// dispatch routes one complete DIMSE message.
func (s *Server) dispatch(ctx context.Context, a *association, cmd *CommandSet, data []byte) error {
	field, ok := cmd.CommandField()
	if !ok {
		return ErrMissingAttr
	}
	switch field {
	case CmdCEchoRQ:
		id, _ := cmd.MessageID()
		sop, _ := cmd.UI(TagRequestedSOPClassUID)
		return s.sendCommand(a, 0, NewCEchoResponse(id, sop, StatusSuccess), nil)

	case CmdCFindRQ:
		return s.handleCFind(ctx, a, cmd, data)

	case CmdCCancelRQ:
		// A cancel that arrives between queries has nothing to cancel.
		return nil

	default:
		id, _ := cmd.MessageID()
		_ = id
		s.cfg.Logf("dicomnet: %s: unsupported DIMSE command 0x%04X", a.conn.RemoteAddr(), field)
		resp := NewCommandSet()
		resp.SetUS(TagCommandField, field|CmdResponseBit)
		if mid, ok := cmd.MessageID(); ok {
			resp.SetUS(TagMessageIDBeingRespondedTo, mid)
		}
		if sop, ok := cmd.UI(TagRequestedSOPClassUID); ok {
			resp.SetUI(TagAffectedSOPClassUID, sop)
		}
		resp.SetUS(TagCommandDataSetType, DataSetTypeNone)
		resp.SetUS(TagStatus, StatusSOPClassUnsup)
		return s.sendCommand(a, firstContextID(a.accepted), resp, nil)
	}
}

// handleCFind executes a Modality Worklist C-FIND and streams the responses.
func (s *Server) handleCFind(ctx context.Context, a *association, cmd *CommandSet, identifier []byte) error {
	messageID, ok := cmd.MessageID()
	if !ok {
		return fmt.Errorf("%w: (0000,0110) message ID", ErrMissingAttr)
	}
	sop, _ := cmd.UI(TagRequestedSOPClassUID)
	if sop == "" {
		sop = UIDModalityWorklistFIND
	}

	// The identifier uses the negotiated transfer syntax.
	explicit := false
	for _, ts := range a.accepted {
		if ts == ExplicitVRLittleEndian {
			explicit = true
		}
		break
	}

	// Identify the context the request arrived on: the first accepted MWL
	// context is the only one a conformant SCU uses, but reply on whichever
	// context is accepted to stay symmetric.
	contextID := firstContextID(a.accepted)

	if len(identifier) == 0 {
		return s.sendCommand(a, contextID, NewCFindResponse(messageID, sop, StatusCannotUnderstand, false), nil)
	}
	ds, err := DecodeDataSet(identifier, explicit)
	if err != nil {
		s.cfg.Logf("dicomnet: %s: malformed C-FIND identifier: %v", a.conn.RemoteAddr(), err)
		return s.sendCommand(a, contextID, NewCFindResponse(messageID, sop, StatusCannotUnderstand, false), nil)
	}
	q, err := ParseMWLQuery(ds)
	if err != nil {
		s.cfg.Logf("dicomnet: %s: unusable C-FIND identifier: %v", a.conn.RemoteAddr(), err)
		return s.sendCommand(a, contextID, NewCFindResponse(messageID, sop, StatusDataSetMismatch, false), nil)
	}
	q.RequestedSOPClassUID = sop
	if len(q.UnsupportedMatchingKeys) > 0 {
		s.cfg.Logf("dicomnet: %s: %d unsupported matching key(s) ignored", a.conn.RemoteAddr(), len(q.UnsupportedMatchingKeys))
	}

	items, err := s.cfg.Source.Find(ctx, q)
	if err != nil {
		s.cfg.Logf("dicomnet: worklist query failed: %v", err)
		return s.sendCommand(a, contextID, NewCFindResponse(messageID, sop, StatusRefusedResources, false), nil)
	}

	matched := 0
	for _, item := range items {
		if !q.MatchItem(item) {
			continue
		}
		respID, err := q.BuildResponseIdentifier(item)
		if err != nil {
			s.cfg.Logf("dicomnet: building response identifier: %v", err)
			return s.sendCommand(a, contextID, NewCFindResponse(messageID, sop, StatusProcessingFail, false), nil)
		}
		encoded, err := EncodeDataSet(respID, explicit)
		if err != nil {
			s.cfg.Logf("dicomnet: encoding response identifier: %v", err)
			return s.sendCommand(a, contextID, NewCFindResponse(messageID, sop, StatusProcessingFail, false), nil)
		}
		if err := s.sendCommand(a, contextID, NewCFindResponse(messageID, sop, StatusPending, true), encoded); err != nil {
			return err
		}
		matched++
	}

	s.cfg.Logf("dicomnet: %s: C-FIND (MWL) message %d returned %d matching item(s)", a.conn.RemoteAddr(), messageID, matched)
	return s.sendCommand(a, contextID, NewCFindResponse(messageID, sop, StatusSuccess, false), nil)
}

// firstContextID returns the lowest accepted presentation context ID. Zero is
// returned when nothing was accepted (the association is never established in
// that case, so it is only a defensive fallback).
func firstContextID(accepted map[byte]string) byte {
	var best byte
	found := false
	for id := range accepted {
		if !found || id < best {
			best = id
			found = true
		}
	}
	if !found {
		return 0
	}
	return best
}

// sendCommand fragments a command set (and optional data set) into P-DATA-TF
// PDUs respecting the peer's maximum PDU length, then writes them.
func (s *Server) sendCommand(a *association, contextID byte, cmd *CommandSet, dataSet []byte) error {
	if contextID == 0 {
		contextID = firstContextID(a.accepted)
	}
	body, err := cmd.Encode()
	if err != nil {
		return err
	}
	pdvs, err := Fragment(contextID, body, dataSet, a.peerMax)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(s.cfg.IdleTimeout)
	for _, pdv := range pdvs {
		if err := s.sendPDU(a.conn, NewDataTF(pdv), deadline); err != nil {
			return err
		}
	}
	return nil
}

// sendPDU encodes and writes a single PDU.
func (s *Server) sendPDU(conn net.Conn, pdu *PDU, deadline time.Time) error {
	b, err := pdu.Encode()
	if err != nil {
		return err
	}
	if !deadline.IsZero() {
		if err := conn.SetWriteDeadline(deadline); err != nil {
			return err
		}
	}
	_, err = conn.Write(b)
	return err
}
