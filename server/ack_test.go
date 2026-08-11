package server

/*
MIT License

Copyright (c) 2026 Shane

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
*/

import (
	"errors"
	"net"
	"testing"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/server/utils"
)

// ackMessage builds an inbound message with the given version and MSH-15/MSH-16
// values. Passing "" for either field leaves it empty (original mode when both
// are empty).
func ackMessage(t *testing.T, version, msh15, msh16 string) *builder.Message {
	t.Helper()
	text := "MSH|^~\\&|SENDER|FAC|RECEIVER|RFAC|20240101000000||ADT^A01|CONTROL_ID|D|" + version + "|||" + msh15 + "|" + msh16
	m, err := builder.NewMessage(builder.MessageOptions{Text: text})
	if err != nil {
		t.Fatalf("build message: %v", err)
	}
	return m
}

// ackResponse wires a SendResponse to an in-memory socket whose peer end is
// drained, so the codec write in SendResponse never blocks.
func ackResponse(t *testing.T, message *builder.Message) *SendResponse {
	t.Helper()
	peer, sock := net.Pipe()
	go func() {
		buf := make([]byte, 4096)
		for {
			if _, err := peer.Read(buf); err != nil {
				return
			}
		}
	}()
	t.Cleanup(func() {
		_ = peer.Close()
		_ = sock.Close()
	})
	return NewSendResponse(sock, message, nil)
}

// TestNegotiateAckCode covers issue #24: code-family selection for both ACK
// classes across every outcome.
func TestNegotiateAckCode(t *testing.T) {
	cases := []struct {
		class   AckClass
		outcome AckOutcome
		want    string
	}{
		{AckAccept, OutcomeSuccess, "CA"},
		{AckAccept, OutcomeError, "CE"},
		{AckAccept, OutcomeReject, "CR"},
		{AckApplication, OutcomeSuccess, "AA"},
		{AckApplication, OutcomeError, "AE"},
		{AckApplication, OutcomeReject, "AR"},
	}
	for _, c := range cases {
		// "AL" always sends, so send is expected true here.
		got, send := NegotiateAck("AL", c.class, c.outcome)
		if got != c.want || !send {
			t.Fatalf("NegotiateAck(AL, %v, %v) = (%q, %v), want (%q, true)", c.class, c.outcome, got, send, c.want)
		}
	}
}

// TestNegotiateAckSendCondition covers issue #24: the HL7 table 0155 send rule
// (AL/NE/ER/SU + empty) across outcomes.
func TestNegotiateAckSendCondition(t *testing.T) {
	cases := []struct {
		field   string
		outcome AckOutcome
		want    bool
	}{
		{"", OutcomeSuccess, true}, // empty -> always
		{"", OutcomeError, true},
		{"AL", OutcomeSuccess, true},
		{"AL", OutcomeReject, true},
		{"NE", OutcomeSuccess, false}, // never
		{"NE", OutcomeError, false},
		{"ER", OutcomeSuccess, false}, // error/reject only
		{"ER", OutcomeError, true},
		{"ER", OutcomeReject, true},
		{"SU", OutcomeSuccess, true}, // success only
		{"SU", OutcomeError, false},
		{"SU", OutcomeReject, false},
		{"??", OutcomeError, true}, // unknown -> conservative always
	}
	for _, c := range cases {
		if _, send := NegotiateAck(c.field, AckApplication, c.outcome); send != c.want {
			t.Fatalf("NegotiateAck(%q, App, %v) send = %v, want %v", c.field, c.outcome, send, c.want)
		}
	}
}

// TestSendAckOriginalMode covers issue #24: with MSH-15 and MSH-16 empty the
// server returns a single application ACK (AA/AE/AR), always sent - the
// pre-#24 behavior.
func TestSendAckOriginalMode(t *testing.T) {
	cases := []struct {
		outcome AckOutcome
		want    string
	}{
		{OutcomeSuccess, "AA"},
		{OutcomeError, "AE"},
		{OutcomeReject, "AR"},
	}
	for _, c := range cases {
		res := ackResponse(t, ackMessage(t, "2.7", "", ""))
		sent, err := res.SendAck(c.outcome)
		if err != nil {
			t.Fatalf("SendAck(%v): %v", c.outcome, err)
		}
		if !sent {
			t.Fatalf("SendAck(%v) sent = false, want true (original mode always acks)", c.outcome)
		}
		if got := res.GetAckMessage().Get("MSA.1").String(); got != c.want {
			t.Fatalf("original-mode MSA.1 = %q, want %q", got, c.want)
		}
	}
}

// TestSendAckEnhancedAcceptCode covers issue #24: enhanced mode returns the
// accept (commit) ACK CA/CE/CR governed by the outcome.
func TestSendAckEnhancedAcceptCode(t *testing.T) {
	cases := []struct {
		outcome AckOutcome
		want    string
	}{
		{OutcomeSuccess, "CA"},
		{OutcomeError, "CE"},
		{OutcomeReject, "CR"},
	}
	for _, c := range cases {
		// MSH-15 = AL -> always send the accept ACK.
		res := ackResponse(t, ackMessage(t, "2.7", "AL", ""))
		sent, err := res.SendAck(c.outcome)
		if err != nil {
			t.Fatalf("SendAck(%v): %v", c.outcome, err)
		}
		if !sent {
			t.Fatalf("SendAck(%v) sent = false, want true", c.outcome)
		}
		if got := res.GetAckMessage().Get("MSA.1").String(); got != c.want {
			t.Fatalf("enhanced accept MSA.1 = %q, want %q", got, c.want)
		}
	}
}

// TestSendAckSuppression covers issue #24: MSH-15's send condition suppresses
// the accept ACK (nothing written) per NE/ER/SU.
func TestSendAckSuppression(t *testing.T) {
	cases := []struct {
		name     string
		msh15    string
		outcome  AckOutcome
		wantSent bool
		wantCode string
	}{
		{"never suppresses success", "NE", OutcomeSuccess, false, ""},
		{"never suppresses error", "NE", OutcomeError, false, ""},
		{"error-only suppresses success", "ER", OutcomeSuccess, false, ""},
		{"error-only sends on error", "ER", OutcomeError, true, "CE"},
		{"success-only sends on success", "SU", OutcomeSuccess, true, "CA"},
		{"success-only suppresses error", "SU", OutcomeError, false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res := ackResponse(t, ackMessage(t, "2.7", c.msh15, ""))
			sent, err := res.SendAck(c.outcome)
			if err != nil {
				t.Fatalf("SendAck: %v", err)
			}
			if sent != c.wantSent {
				t.Fatalf("sent = %v, want %v", sent, c.wantSent)
			}
			if !c.wantSent {
				if res.GetAckMessage() != nil {
					t.Fatalf("expected no ACK, got %q", res.GetAckMessage().String())
				}
				return
			}
			if got := res.GetAckMessage().Get("MSA.1").String(); got != c.wantCode {
				t.Fatalf("MSA.1 = %q, want %q", got, c.wantCode)
			}
		})
	}
}

// TestSendApplicationAck covers issue #24: the MSH-16 application-ACK path -
// code family AA/AE/AR plus MSH-16 send-condition suppression.
func TestSendApplicationAck(t *testing.T) {
	cases := []struct {
		name     string
		msh16    string
		outcome  AckOutcome
		wantSent bool
		wantCode string
	}{
		{"always on success", "AL", OutcomeSuccess, true, "AA"},
		{"always on error", "AL", OutcomeError, true, "AE"},
		{"always on reject", "AL", OutcomeReject, true, "AR"},
		{"never suppresses", "NE", OutcomeSuccess, false, ""},
		{"error-only sends on error", "ER", OutcomeError, true, "AE"},
		{"error-only suppresses success", "ER", OutcomeSuccess, false, ""},
		{"success-only sends on success", "SU", OutcomeSuccess, true, "AA"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// MSH-15 populated so the message is in enhanced mode; MSH-16 drives
			// the application ACK under test.
			res := ackResponse(t, ackMessage(t, "2.7", "AL", c.msh16))
			sent, err := res.SendApplicationAck(c.outcome)
			if err != nil {
				t.Fatalf("SendApplicationAck: %v", err)
			}
			if sent != c.wantSent {
				t.Fatalf("sent = %v, want %v", sent, c.wantSent)
			}
			if !c.wantSent {
				return
			}
			if got := res.GetAckMessage().Get("MSA.1").String(); got != c.wantCode {
				t.Fatalf("MSA.1 = %q, want %q", got, c.wantCode)
			}
		})
	}
}

// TestSendAckValidatesAgainstVersion covers issue #24: the accept ACK code is
// validated against the inbound MSH-12 - CA/CE/CR are rejected on HL7 2.1.
func TestSendAckValidatesAgainstVersion(t *testing.T) {
	res := ackResponse(t, ackMessage(t, "2.1", "AL", ""))
	sent, err := res.SendAck(OutcomeSuccess)
	if err == nil {
		t.Fatalf("expected an MSA.1 validation error for CA on HL7 2.1, got sent=%v", sent)
	}
	if sent {
		t.Fatalf("sent = true, want false on validation failure")
	}
	var serverErr *utils.HL7ServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("error type = %T, want *HL7ServerError", err)
	}
	if res.GetAckMessage() != nil {
		t.Fatalf("expected no ACK stored on validation failure, got %q", res.GetAckMessage().String())
	}
}
