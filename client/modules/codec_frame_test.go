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
	"strings"
	"testing"

	"github.com/Bugs5382/go-hl7/client/modules"
)

// These tests pin issue #35: a frame completes only on the adjacent <FS><CR>
// trailer. A body carries CR segment separators, so a buffer that holds an FS
// and some CR is not necessarily a complete frame.

const (
	bodyA = "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|A1|P|2.7\rEVN|A01"
	bodyB = "MSH|^~\\&|B|FAC|||20260101000000||ADT^A01|B1|P|2.7\rEVN|A01"
	bodyC = "MSH|^~\\&|C|FAC|||20260101000000||ADT^A01|C1|P|2.7\rEVN|A01"
)

// receive feeds one read to the codec and returns the message it completed, or
// "" and false when the read completed nothing.
func receive(t *testing.T, c *modules.MLLPCodec, data string) (string, bool) {
	t.Helper()
	if !c.ReceiveString(data) {
		return "", false
	}
	got := c.GetLastMessage()
	if got == nil {
		t.Fatalf("ReceiveData returned true for %q but no message was decoded", data)
	}
	return *got, true
}

func TestMLLPCodecTrailerSplitAcrossReads(t *testing.T) {
	c := modules.NewMLLPCodec("")
	if got, ok := receive(t, c, vt+bodyA+fs+cr); !ok || got != bodyA {
		t.Fatalf("frame A: got %q, %v", got, ok)
	}
	// The read ends between B's FS and its CR. B's body has a CR segment
	// separator, so FS and CR are both in the buffer, but not as a trailer.
	if got, ok := receive(t, c, vt+bodyB+fs); ok {
		t.Fatalf("a read ending between FS and CR must complete nothing, got %q", got)
	}
	if got, ok := receive(t, c, cr); !ok || got != bodyB {
		t.Fatalf("frame B after its CR: got %q, %v", got, ok)
	}
}

func TestMLLPCodecTrailerSplitOnFirstFrame(t *testing.T) {
	c := modules.NewMLLPCodec("")
	if _, ok := receive(t, c, vt+bodyA+fs); ok {
		t.Fatalf("a read ending between FS and CR must complete nothing")
	}
	if c.GetLastMessage() != nil {
		t.Fatalf("no message should be decoded before the trailer arrives")
	}
	if got, ok := receive(t, c, cr); !ok || got != bodyA {
		t.Fatalf("frame A after its CR: got %q, %v", got, ok)
	}
}

func TestMLLPCodecSeveralFramesInOneRead(t *testing.T) {
	c := modules.NewMLLPCodec("\n")
	got, ok := receive(t, c, vt+bodyA+fs+cr+vt+bodyB+fs+cr+vt+bodyC+fs+cr)
	if !ok || got != bodyA+"\n"+bodyB+"\n"+bodyC {
		t.Fatalf("several frames in one read: got %q, %v", got, ok)
	}
	// Everything was consumed, so a partial next frame completes nothing.
	if got, ok := receive(t, c, vt+"MSH|^~\\&|D\rEVN"); ok {
		t.Fatalf("partial frame after a fully consumed read completed %q", got)
	}
}

func TestMLLPCodecFrameThenStartOfNext(t *testing.T) {
	c := modules.NewMLLPCodec("")
	split := len(bodyB) / 2
	if got, ok := receive(t, c, vt+bodyA+fs+cr+vt+bodyB[:split]); !ok || got != bodyA {
		t.Fatalf("frame A with the start of B: got %q, %v", got, ok)
	}
	if got, ok := receive(t, c, bodyB[split:]+fs+cr); !ok || got != bodyB {
		t.Fatalf("rest of frame B: got %q, %v", got, ok)
	}
}

func TestMLLPCodecFrameThenNextUpToFS(t *testing.T) {
	c := modules.NewMLLPCodec("")
	if got, ok := receive(t, c, vt+bodyA+fs+cr+vt+bodyB+fs); !ok || got != bodyA {
		t.Fatalf("frame A with B up to its FS: got %q, %v", got, ok)
	}
	if got, ok := receive(t, c, cr); !ok || got != bodyB {
		t.Fatalf("frame B after its CR: got %q, %v", got, ok)
	}
}

// fuzzBody turns arbitrary text into a body the codec carries unchanged: the
// codec strips VT and FS and trims surrounding whitespace, and LF is reserved
// here as the join character so decoded frames can be told apart.
func fuzzBody(s string) string {
	s = strings.NewReplacer(vt, "", fs, "", "\n", "").Replace(s)
	return strings.TrimSpace(s)
}

// FuzzMLLPCodecFrames sends two frames as one stream cut into three reads at
// arbitrary offsets, the way TCP can deliver them. Every read that reports a
// complete frame must yield new frames, and together the reads must yield
// exactly the two bodies, in order, once each.
func FuzzMLLPCodecFrames(f *testing.F) {
	frameA := len(vt + bodyA + fs + cr)
	f.Add(bodyA, bodyB, uint16(0), uint16(0))
	f.Add(bodyA, bodyB, uint16(frameA), uint16(frameA))
	f.Add(bodyA, bodyB, uint16(frameA), uint16(frameA+len(vt+bodyB+fs)))
	f.Add(bodyA, bodyB, uint16(frameA-1), uint16(frameA+len(bodyB)/2))
	f.Add("MSH|1\r\rPID", "MSH|2\rPID\r", uint16(3), uint16(9))

	f.Fuzz(func(t *testing.T, rawA, rawB string, cut1, cut2 uint16) {
		a, b := fuzzBody(rawA), fuzzBody(rawB)
		if a == "" || b == "" {
			return
		}
		var stream bytes.Buffer
		sender := modules.NewMLLPCodec("")
		if err := sender.SendMessage(&stream, a); err != nil {
			t.Fatalf("SendMessage(a): %v", err)
		}
		if err := sender.SendMessage(&stream, b); err != nil {
			t.Fatalf("SendMessage(b): %v", err)
		}
		data := stream.Bytes()
		k1 := int(cut1) % (len(data) + 1)
		k2 := int(cut2) % (len(data) + 1)
		if k1 > k2 {
			k1, k2 = k2, k1
		}

		c := modules.NewMLLPCodec("\n")
		var got []string
		for _, read := range [][]byte{data[:k1], data[k1:k2], data[k2:]} {
			if !c.ReceiveData(read) {
				continue
			}
			msg := c.GetLastMessage()
			if msg == nil {
				t.Fatalf("ReceiveData returned true but no message was decoded")
			}
			got = append(got, *msg)
		}
		want := a + "\n" + b
		if joined := strings.Join(got, "\n"); joined != want {
			t.Fatalf("reads cut at %d and %d of %d\nwant %q\ngot  %q", k1, k2, len(data), want, got)
		}
	})
}
