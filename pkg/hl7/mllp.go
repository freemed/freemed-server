package hl7

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// MLLP frame markers.
const (
	MLLPStartBlock = 0x0B // <VT> vertical tab
	MLLPEndBlock   = 0x1C // <FS> file separator
	MLLPEndMsg     = 0x0D // <CR> carriage return
)

// MLLP frame delimiters as byte slices.
var (
	mllpStart  = []byte{MLLPStartBlock}
	mllpEnd    = []byte{MLLPEndBlock, MLLPEndMsg}
)

// StartMLLPServer starts a TCP listener that accepts HL7v2 messages over MLLP.
//
// addr is the TCP address to bind (e.g. ":2575"). The default HL7 port is 2575.
// handler is called for each parsed message. If it returns an error, a negative
// acknowledgement (MSA|AR) is sent; otherwise a positive acknowledgement (MSA|AA).
//
// Returns immediately; runs in the background. Errors during accept are logged.
func StartMLLPServer(addr string, handler func(*Message) error) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("mllp: failed to listen on %s: %w", addr, err)
	}

	log.Printf("MLLP server listening on %s", addr)

	go func() {
		defer ln.Close()
		for {
			conn, err := ln.Accept()
			if err != nil {
				// Check if the listener was closed
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					log.Printf("mllp: temporary accept error: %v", err)
					continue
				}
				log.Printf("mllp: accept error: %v", err)
				return
			}
			go handleMLLPConnection(conn, handler)
		}
	}()

	return nil
}

// handleMLLPConnection reads MLLP-framed messages from a connection.
func handleMLLPConnection(conn net.Conn, handler func(*Message) error) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		// Set a read deadline (no data for 30s = timeout)
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// Read until the start-of-block marker (0x0B)
		_, err := reader.ReadBytes(MLLPStartBlock)
		if err != nil {
			if err != io.EOF {
				log.Printf("mllp: error reading start block: %v", err)
			}
			return
		}

		// Read the message body until end-of-block + carriage return (0x1C 0x0D)
		msgBytes, err := reader.ReadBytes(MLLPEndBlock)
		if err != nil {
			log.Printf("mllp: error reading message: %v", err)
			return
		}

		// Strip the end markers (last two bytes should be 0x1C and the 0x0D)
		// The message ends with: ...<FS><CR>
		// After ReadBytes(MLLPEndBlock), msgBytes ends with <FS>, and the <CR>
		// follows. Read the trailing CR.
		if len(msgBytes) > 0 && msgBytes[len(msgBytes)-1] == MLLPEndBlock {
			msgBytes = msgBytes[:len(msgBytes)-1] // strip FS
		}

		// Read the CR that follows the FS
		crByte, err := reader.ReadByte()
		if err != nil {
			log.Printf("mllp: error reading trailing CR: %v", err)
			return
		}
		if crByte != MLLPEndMsg {
			log.Printf("mllp: expected trailing CR (0x0D), got 0x%02X", crByte)
		}

		// Reset read deadline (no timeout for processing)
		conn.SetReadDeadline(time.Time{})

		// Parse the HL7 message
		msg, err := Unmarshal(msgBytes)
		if err != nil {
			log.Printf("mllp: parse error: %v", err)
			sendMLLPAck(conn, "AE", "Failed to parse message: "+err.Error())
			continue
		}

		// Call the handler
		handlerErr := handler(msg)
		if handlerErr != nil {
			log.Printf("mllp: handler error: %v", handlerErr)
			sendMLLPAck(conn, "AR", "Application error: "+handlerErr.Error())
		} else {
			sendMLLPAck(conn, "AA", "Message accepted")
		}
	}
}

// sendMLLPAck sends an HL7 acknowledgement framed in MLLP.
func sendMLLPAck(conn net.Conn, ackCode, ackText string) {
	now := time.Now().Format("20060102150405")

	ack := fmt.Sprintf(
		"MSH|^~\\&|FREEMED|FREEMED_FACILITY|DEST_APP|DEST_FACILITY|%s||ACK|||P|2.5\r"+
			"MSA|%s||%s",
		now, ackCode, ackText,
	)

	// Frame with MLLP
	var frame []byte
	frame = append(frame, MLLPStartBlock)
	frame = append(frame, []byte(ack)...)
	frame = append(frame, MLLPEndBlock, MLLPEndMsg)

	conn.Write(frame)
}
