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
	"strings"
	"testing"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/server/utils"
)

// issueTestMessage builds the minimal valid HL7 message the issue suite uses.
func issueTestMessage(t *testing.T, controlID string) *builder.Message {
	t.Helper()
	if controlID == "" {
		controlID = "CONTROL_ID"
	}
	m, err := builder.NewMessage(builder.MessageOptions{
		Text: "MSH|^~\\&|SENDER|FAC|RECEIVER|RFAC|20240101000000||ADT^A01|" + controlID + "|D|2.7",
	})
	if err != nil {
		t.Fatalf("build message: %v", err)
	}
	return m
}

// TestPR134InboundRequestSocket checks that InboundRequest exposes the
// underlying socket, and GetSocket panics when none was provided.
func TestPR134InboundRequestSocket(t *testing.T) {
	t.Run("getSocket returns the socket passed in props", func(t *testing.T) {
		fake, real := net.Pipe()
		defer func() { _ = fake.Close() }()
		defer func() { _ = real.Close() }()
		req := NewInboundRequest(issueTestMessage(t, ""), InboundRequestProps{Socket: fake, Type: "message"})
		if req.GetSocket() != fake {
			t.Fatalf("GetSocket() did not return the provided socket")
		}
	})

	t.Run("getSocket panics when no socket was provided", func(t *testing.T) {
		defer func() {
			r := recover()
			err, ok := r.(error)
			if !ok || err.Error() != "Socket is not defined." {
				t.Fatalf("recover = %v, want HL7ListenerError 'Socket is not defined.'", r)
			}
			var listenerErr *utils.HL7ListenerError
			if !errors.As(err, &listenerErr) {
				t.Fatalf("recover type = %T, want *HL7ListenerError", err)
			}
		}()
		req := NewInboundRequest(issueTestMessage(t, ""), InboundRequestProps{Type: "message"})
		req.GetSocket()
	})
}

// TestRejectUnparseableReadableMSH covers issue #49: when the first segment of
// an unparseable message is a readable MSH, the reject ACK echoes its control
// id, AR in original mode and CR in enhanced mode (MSH-15 set).
func TestRejectUnparseableReadableMSH(t *testing.T) {
	cases := []struct{ name, text, want string }{
		{"original", "MSH|^~\\&|A|B|C|D|20240101||ADT^A01|CID1|P|2.5.1\rEVN|A01", "MSA|AR|CID1"},
		{"enhanced", "MSH|^~\\&|A|B|C|D|20240101||ADT^A01|CID2|P|2.5.1|||AL\rEVN|A01", "MSA|CR|CID2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, srvSide := net.Pipe()
			defer func() { _ = client.Close() }()
			defer func() { _ = srvSide.Close() }()
			got := make(chan string, 1)
			go func() {
				buf := make([]byte, 4096)
				n, _ := client.Read(buf)
				got <- string(buf[:n])
			}()
			(&Inbound{}).rejectUnparseable(srvSide, tc.text)
			if ack := <-got; !strings.Contains(ack, tc.want) {
				t.Fatalf("ack %q does not contain %q", ack, tc.want)
			}
		})
	}
}
