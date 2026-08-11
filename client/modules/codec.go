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

	"golang.org/x/text/encoding"

	"github.com/Bugs5382/go-hl7/client/helpers"
)

// MLLPCodec frames HL7 messages over the Minimal Lower Layer Protocol: a
// message body is wrapped <VT>body<FS><CR> (0x0B body 0x1C 0x0D). It buffers
// split frames, accumulating bytes until a complete <FS><CR> end marker
// arrives.
//
// The body is encoded on send and decoded on receive using a configurable
// charset (issue #26), the analog of node-hl7's BufferEncoding. HL7 declares
// its charset in MSH-18; pass that value (an MSH-18 code such as "8859/1", or
// an IANA/WHATWG name) to NewMLLPCodecWithCharset. The default is UTF-8, so an
// empty charset preserves the original UTF-8 byte-slice behavior.
type MLLPCodec struct {
	// returnCharacter joins the decoded message parts; it defaults to "\r".
	returnCharacter string
	// charset is the configured MSH-18 / IANA charset name, or "" for UTF-8.
	// It is retained for reporting; encoding drives the conversions.
	charset string
	// encoding converts between the UTF-8 Go string and the wire charset. A nil
	// encoding means UTF-8 (or an ASCII subset), for which the bytes pass through
	// unchanged.
	encoding encoding.Encoding
	// dataBuffer accumulates incoming bytes across ReceiveData calls until a
	// full frame is present.
	dataBuffer []byte
	// lastMessage holds the most recently decoded message, or nil when none has
	// been decoded yet.
	lastMessage *string
}

// NewMLLPCodec constructs a UTF-8 MLLPCodec. Pass "" for returnCharacter to use
// the "\r" default. It is equivalent to NewMLLPCodecWithCharset with an empty
// charset and never fails, so the historical single-argument form is preserved.
func NewMLLPCodec(returnCharacter string) *MLLPCodec {
	// An empty charset resolves to UTF-8 and cannot error, so the returned error
	// is safely discarded to keep this constructor's original signature.
	c, _ := NewMLLPCodecWithCharset(returnCharacter, "")
	return c
}

// NewMLLPCodecWithCharset constructs an MLLPCodec that encodes on send and
// decodes on receive using charset. charset accepts an HL7 MSH-18 code (for
// example "8859/1" or "UNICODE UTF-8") or an IANA/WHATWG name (for example
// "iso-8859-1"). An empty charset, or any UTF-8/ASCII name, keeps the default
// UTF-8 behavior. An unknown or unsupported charset returns an error and no
// codec. Pass "" for returnCharacter to use the "\r" default.
func NewMLLPCodecWithCharset(returnCharacter, charset string) (*MLLPCodec, error) {
	if returnCharacter == "" {
		returnCharacter = "\r"
	}
	enc, err := ResolveCharset(charset)
	if err != nil {
		return nil, err
	}
	return &MLLPCodec{returnCharacter: returnCharacter, charset: charset, encoding: enc}, nil
}

// Charset reports the configured charset name, or "" when the codec uses UTF-8.
func (c *MLLPCodec) Charset() string { return c.charset }

// GetLastMessage returns the last decoded message, or nil when none has been
// decoded.
func (c *MLLPCodec) GetLastMessage() *string {
	return c.lastMessage
}

// ReceiveData appends incoming bytes and processes a complete frame when the
// buffer holds both the end byte (FS, 0x1C) and footer byte (CR, 0x0D). It
// returns true when a message was processed, or false while still waiting for
// the rest of a split frame.
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

// SendMessage frames message as <VT>message<FS><CR> and writes it to w, where w
// may be any net.Conn or buffer. The body is encoded into the configured charset
// first (a no-op for the UTF-8 default). A nil writer is a no-op. An encoding
// failure is returned without writing.
func (c *MLLPCodec) SendMessage(w io.Writer, message string) error {
	if w == nil {
		return nil
	}
	body, err := c.encodeBody(message)
	if err != nil {
		return err
	}
	buf := make([]byte, 0, len(body)+3)
	buf = append(buf, helpers.ProtocolMLLPHeader)
	buf = append(buf, body...)
	buf = append(buf, helpers.ProtocolMLLPEnd)
	buf = append(buf, helpers.ProtocolMLLPFooter)
	_, err = w.Write(buf)
	return err
}

// encodeBody converts the UTF-8 message string into the configured charset's
// bytes. A nil encoding (UTF-8/ASCII) returns the raw UTF-8 bytes unchanged.
func (c *MLLPCodec) encodeBody(message string) ([]byte, error) {
	if c.encoding == nil {
		return []byte(message), nil
	}
	out, err := c.encoding.NewEncoder().Bytes([]byte(message))
	if err != nil {
		return nil, helpers.NewHL7FatalErrorf("charset encode failed for %q: %v", c.charset, err)
	}
	return out, nil
}

// decodeBody converts charset-encoded body bytes back into a UTF-8 string. A nil
// encoding (UTF-8/ASCII) returns the bytes as-is.
func (c *MLLPCodec) decodeBody(body []byte) string {
	if c.encoding == nil {
		return string(body)
	}
	out, err := c.encoding.NewDecoder().Bytes(body)
	if err != nil {
		// Supported single-byte charsets decode every byte without error; on the
		// rare failure fall back to the raw bytes so no data is silently dropped.
		return string(body)
	}
	return string(out)
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
	marker := []byte{helpers.ProtocolMLLPEnd, helpers.ProtocolMLLPFooter}
	end := bytes.LastIndex(c.dataBuffer, marker)
	if end < 0 {
		return
	}
	complete := c.dataBuffer[:end+2]
	remainder := append([]byte(nil), c.dataBuffer[end+2:]...)

	// Work on bytes rather than a UTF-8 string: the body may be in a non-UTF-8
	// charset (issue #26), so it must be decoded from its original bytes. The MLLP
	// framing (VT/FS/CR) and surrounding whitespace are ASCII and unaffected.
	messages := make([]string, 0)
	for _, part := range bytes.Split(complete, marker) {
		trimmed := bytes.TrimSpace(part)
		if len(trimmed) == 0 {
			continue
		}
		messages = append(messages, c.decodeBody(stripMLLPCharacters(trimmed)))
	}

	joined := strings.Join(messages, c.returnCharacter)
	c.lastMessage = &joined

	c.dataBuffer = remainder
}

// stripMLLPCharacters removes the VT (0x0B) and FS (0x1C) control characters
// from a message body, leaving any charset-encoded payload bytes intact.
func stripMLLPCharacters(message []byte) []byte {
	message = bytes.ReplaceAll(message, []byte{0x0b}, nil)
	message = bytes.ReplaceAll(message, []byte{0x1c}, nil)
	return message
}
