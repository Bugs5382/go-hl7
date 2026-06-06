package modules

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
	"io"
	"strings"

	"github.com/Bugs5382/go-hl7/client/helpers"
)

// MLLPCodec frames HL7 messages over the Minimal Lower Layer Protocol: a
// message body is wrapped <VT>body<FS><CR> (0x0B body 0x1C 0x0D). It mirrors
// the modules/codec.ts MLLPCodec, including the split-frame buffering:
// receiveData accumulates bytes until a complete <FS><CR> end marker arrives.
type MLLPCodec struct {
	// returnCharacter joins the decoded message parts. The spec defaults to "\r".
	returnCharacter string
	// dataBuffer accumulates incoming bytes across receiveData calls until a
	// full frame is present (the dataBuffer Buffer).
	dataBuffer []byte
	// lastMessage holds the most recently decoded message, or nil when none has
	// been decoded yet (the lastMessage: string | undefined).
	lastMessage *string
}

// NewMLLPCodec constructs an MLLPCodec. The spec takes (encoding = "utf8",
// returnCharacter = "\r"); Go strings are already UTF-8 byte slices so the
// encoding argument is unnecessary here. Pass "" for returnCharacter to use the
// "\r" default.
func NewMLLPCodec(returnCharacter string) *MLLPCodec {
	if returnCharacter == "" {
		returnCharacter = "\r"
	}
	return &MLLPCodec{returnCharacter: returnCharacter}
}

// GetLastMessage returns the last decoded message, or nil when none has been
// decoded (the getLastMessage(): null | string).
func (c *MLLPCodec) GetLastMessage() *string {
	return c.lastMessage
}

// ReceiveData appends incoming bytes and processes a complete frame when the
// buffer holds both the end byte (FS, 0x1C) and footer byte (CR, 0x0D). It
// returns true when a message was processed, or false while still waiting for
// the rest of a split frame. Mirrors the receiveData.
func (c *MLLPCodec) ReceiveData(data []byte) bool {
	c.dataBuffer = append(c.dataBuffer, data...)

	// Only process once the buffer contains the end and footer protocol bytes.
	if bytes.IndexByte(c.dataBuffer, helpers.ProtocolMLLPEnd) >= 0 &&
		bytes.IndexByte(c.dataBuffer, helpers.ProtocolMLLPFooter) >= 0 {
		c.processMessage()
		return true
	}

	// Still waiting for more of the message to come over.
	return false
}

// ReceiveString is the string convenience form of ReceiveData (accepts
// `Buffer | string`).
func (c *MLLPCodec) ReceiveString(data string) bool {
	return c.ReceiveData([]byte(data))
}

// SendMessage frames message as <VT>message<FS><CR> and writes it to w. The spec
// takes (socket, message, encoding); Go uses an io.Writer so any net.Conn or
// buffer works, and the body is already a UTF-8 string. A nil writer is a
// no-op, mirroring the optional `socket?.write`.
func (c *MLLPCodec) SendMessage(w io.Writer, message string) error {
	if w == nil {
		return nil
	}
	buf := make([]byte, 0, len(message)+3)
	buf = append(buf, helpers.ProtocolMLLPHeader)
	buf = append(buf, message...)
	buf = append(buf, helpers.ProtocolMLLPEnd)
	buf = append(buf, helpers.ProtocolMLLPFooter)
	_, err := w.Write(buf)
	return err
}

// processMessage decodes the buffered frames: split on the <FS><CR> end marker,
// trim and strip the MLLP control characters from each non-empty part, and join
// the complete frames with returnCharacter.
//
// The source splits the whole buffer and then clears it. Go's stream reads
// coalesce many small frames into one read far more readily than Node's
// per-write data events, so a buffer routinely ends mid-frame; clearing the
// whole buffer would silently drop that trailing partial frame. To stay correct
// under that coalescing we consume only through the last complete <FS><CR> and
// retain any trailing partial frame for the next call. With a single complete
// frame this is byte-for-byte identical to the source. See QUESTIONS for this
// adaptation.
func (c *MLLPCodec) processMessage() {
	end := bytes.LastIndex(c.dataBuffer, []byte{helpers.ProtocolMLLPEnd, helpers.ProtocolMLLPFooter})
	if end < 0 {
		return
	}
	complete := string(c.dataBuffer[:end+2])
	remainder := append([]byte(nil), c.dataBuffer[end+2:]...)

	messages := make([]string, 0)
	parts := strings.Split(complete, string([]byte{helpers.ProtocolMLLPEnd, helpers.ProtocolMLLPFooter}))
	for _, part := range parts {
		if strings.TrimSpace(part) != "" {
			messages = append(messages, stripMLLPCharacters(strings.TrimSpace(part)))
		}
	}

	joined := strings.Join(messages, c.returnCharacter)
	c.lastMessage = &joined

	c.dataBuffer = remainder
}

// stripMLLPCharacters removes the VT (0x0B, "") and FS (0x1C, "")
// control characters from a message body. Mirrors the stripMLLPCharacters.
func stripMLLPCharacters(message string) string {
	message = strings.ReplaceAll(message, "\x0b", "")
	message = strings.ReplaceAll(message, "\x1c", "")
	return message
}
