package modules_test

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
	"bytes"
	"testing"

	"github.com/Bugs5382/go-hl7/client/modules"
)

// These tests mirror the MLLPCodec behavior (modules/codec.ts): VT/FS/CR
// framing on send, accumulate-and-decode on receive, and the split-frame
// buffering the spec server-level issue test exercises end-to-end.

const (
	vt = "\x0b"
	fs = "\x1c"
	cr = "\r"
)

func TestMLLPCodecSendMessageFraming(t *testing.T) {
	c := modules.NewMLLPCodec("")
	var buf bytes.Buffer
	if err := c.SendMessage(&buf, "MSH|^~\\&|APP"); err != nil {
		t.Fatalf("SendMessage error: %v", err)
	}
	want := vt + "MSH|^~\\&|APP" + fs + cr
	if buf.String() != want {
		t.Fatalf("framing mismatch: got %q want %q", buf.String(), want)
	}
}

func TestMLLPCodecSendMessageNilWriter(t *testing.T) {
	c := modules.NewMLLPCodec("")
	if err := c.SendMessage(nil, "MSH|whatever"); err != nil {
		t.Fatalf("nil writer should be a no-op, got error: %v", err)
	}
}

func TestMLLPCodecGetLastMessageNilInitially(t *testing.T) {
	c := modules.NewMLLPCodec("")
	if c.GetLastMessage() != nil {
		t.Fatalf("expected nil last message before any decode")
	}
}

func TestMLLPCodecReceiveCompleteFrame(t *testing.T) {
	c := modules.NewMLLPCodec("")
	frame := vt + "MSH|^~\\&|APP" + fs + cr
	if !c.ReceiveString(frame) {
		t.Fatalf("expected ReceiveData to return true on a complete frame")
	}
	got := c.GetLastMessage()
	if got == nil {
		t.Fatalf("expected a decoded message")
	}
	if *got != "MSH|^~\\&|APP" {
		t.Fatalf("decoded mismatch: got %q", *got)
	}
}

func TestMLLPCodecSplitFrame(t *testing.T) {
	c := modules.NewMLLPCodec("")
	// First chunk: start byte + partial body, no end marker yet.
	if c.ReceiveString(vt + "MSH|^~\\&|") {
		t.Fatalf("expected false while waiting for the rest of the frame")
	}
	if c.GetLastMessage() != nil {
		t.Fatalf("no message should be decoded from a partial frame")
	}
	// Second chunk: remainder of body + end marker.
	if !c.ReceiveString("APP|FAC" + fs + cr) {
		t.Fatalf("expected true once the end marker arrives")
	}
	got := c.GetLastMessage()
	if got == nil || *got != "MSH|^~\\&|APP|FAC" {
		t.Fatalf("decoded split frame mismatch: %v", got)
	}
}

func TestMLLPCodecMultipleFramesJoined(t *testing.T) {
	c := modules.NewMLLPCodec("")
	data := vt + "MSG_A" + fs + cr + vt + "MSG_B" + fs + cr
	if !c.ReceiveString(data) {
		t.Fatalf("expected true for buffered frames")
	}
	got := c.GetLastMessage()
	if got == nil || *got != "MSG_A\rMSG_B" {
		t.Fatalf("expected joined messages, got %v", got)
	}
}

func TestMLLPCodecCustomReturnCharacter(t *testing.T) {
	c := modules.NewMLLPCodec("\n")
	data := vt + "MSG_A" + fs + cr + vt + "MSG_B" + fs + cr
	c.ReceiveString(data)
	got := c.GetLastMessage()
	if got == nil || *got != "MSG_A\nMSG_B" {
		t.Fatalf("expected newline-joined messages, got %v", got)
	}
}

func TestMLLPCodecRoundTrip(t *testing.T) {
	enc := modules.NewMLLPCodec("")
	dec := modules.NewMLLPCodec("")
	var buf bytes.Buffer
	if err := enc.SendMessage(&buf, "MSH|round^trip"); err != nil {
		t.Fatalf("SendMessage error: %v", err)
	}
	if !dec.ReceiveData(buf.Bytes()) {
		t.Fatalf("expected the framed message to decode")
	}
	got := dec.GetLastMessage()
	if got == nil || *got != "MSH|round^trip" {
		t.Fatalf("round trip mismatch: %v", got)
	}
}
