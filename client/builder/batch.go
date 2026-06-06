// MIT License
//
// Copyright (c) 2026 Shane
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package builder

import (
	"strconv"
	"strings"
	"time"

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// Batch is the BHS-rooted batch builder (the Batch). It collects Message
// children under a Batch Header Segment (BHS) and closes with a Batch Trailing
// Segment (BTS) carrying the message count.
type Batch struct {
	RootBase
	opt           builderOptions
	lines         []string
	messagesCount int
}

// NewBatch builds an empty batch (writing BHS.7 with the current date), or
// parses the provided text into the MSH lines it carries. It mirrors the
// Batch constructor, returning an error where the spec throws.
func NewBatch(opts BatchOptions) (*Batch, error) {
	opt, err := normalizedClientBatchParserOptions(opts)
	if err != nil {
		return nil, err
	}

	b := &Batch{opt: opt}
	b.initRootBase(b, opt)

	if opt.Text != "" {
		for _, line := range utils.Split(opt.Text, nil) {
			if strings.HasPrefix(line, "MSH") {
				b.lines = append(b.lines, line)
			}
		}
	} else {
		b.Set("BHS.7", utils.CreateHL7Date(time.Now(), b.opt.Date))
	}

	return b, nil
}

// Add adds a Message to the batch, bumping the message count. With index < 0 the
// message is appended; otherwise it is inserted at index (the Batch.add).
func (b *Batch) Add(message *Message, index int) {
	b.setDirty()
	b.messagesCount = b.messagesCount + 1
	if index < 0 {
		b.children = append(b.childrenOf(), message)
	} else {
		ch := b.childrenOf()
		ch = append(ch, nil)
		copy(ch[index+1:], ch[index:])
		ch[index] = message
		b.children = ch
	}
}

// End closes the batch, appending the BTS trailer with the message count (the
// Batch.end).
func (b *Batch) End() {
	segment := b.addSegment("BTS")
	segment.Set("1", b.messagesCount)
}

// Text returns the normalized source text the batch was parsed from (the
// _opt.text).
func (b *Batch) Text() string { return b.opt.Text }

// Get resolves a BHS segment or field path (the Batch.get).
func (b *Batch) Get(path string) HL7Node {
	if path == "" {
		return emptySingleton
	}
	p := b.preparePath(path)
	rv := b.Read(p)
	if rv == nil {
		return emptySingleton
	}
	return rv
}

// GetFirstSegment returns the first segment found in the batch matching name
// (the getFirstSegment).
func (b *Batch) GetFirstSegment(name string) *Segment {
	return b.firstSegment(name)
}

// Messages parses the carried text into individual Message nodes, one per MSH
// line (the Batch.messages). It panics with HL7FatalError when there is no
// batch text to parse, mirroring the throw.
func (b *Batch) Messages() []*Message {
	if b.lines != nil && b.opt.NewLine != "" {
		messages := make([]*Message, 0, len(b.lines))
		for _, line := range b.lines {
			text := strings.TrimSuffix(line, b.opt.NewLine)
			m, err := NewMessage(MessageOptions{Text: text})
			if err != nil {
				panic(err)
			}
			messages = append(messages, m)
		}
		return messages
	}
	panic(helpers.NewHL7FatalError("No messages inside batch."))
}

// Read resolves a split path, returning a SegmentList for a bare segment name
// (the Batch.read).
func (b *Batch) Read(path []string) HL7Node {
	if len(path) == 0 {
		panic(helpers.NewHL7FatalError("Unable to process the read function correctly."))
	}
	segmentName := path[0]
	rest := path[1:]
	if len(rest) == 0 {
		var segments []*Segment
		for _, c := range b.childrenOf() {
			if seg, ok := c.(*Segment); ok && seg.Name() == segmentName {
				segments = append(segments, seg)
			}
		}
		if len(segments) > 0 {
			return newSegmentList(b, segments)
		}
	} else {
		seg := b.firstSegment(segmentName)
		if seg != nil {
			return seg.Read(rest)
		}
	}
	panic(helpers.NewHL7FatalError("Unable to process the read function correctly."))
}

// Set writes a value at a batch segment/field path (the Batch.set).
func (b *Batch) Set(path string, value any) HL7Node {
	if arr, ok := value.([]any); ok {
		for i, item := range arr {
			b.Set(path+"."+strconv.Itoa(i+1), item)
		}
		return b
	}
	p := b.preparePath(path)
	b.Write(p, b.prepareValue(value))
	return b
}

// Start (re)writes BHS.7 with the current date in the given style (the
// Batch.start). An empty style defaults to the 14-character form.
func (b *Batch) Start(style string) {
	b.Set("BHS.7", utils.CreateHL7Date(time.Now(), style))
}

// ToFile wraps the batch in a FileBatch and writes it to disk, returning the
// generated file name (the Batch.toFile).
func (b *Batch) ToFile(name string, newLine bool, location string, extension string) (string, error) {
	if extension == "" {
		extension = "hl7"
	}
	nl := ""
	if newLine {
		nl = "\n"
	}
	fileBatch, err := NewFileBatch(FileOptions{
		Extension: extension,
		Location:  location,
		NewLine:   nl,
	})
	if err != nil {
		return "", err
	}
	fileBatch.Start()

	fileBatch.Set("FHS.3", b.Get("BHS.3").String())
	fileBatch.Set("FHS.4", b.Get("BHS.4").String())
	fileBatch.Set("FHS.5", b.Get("BHS.5").String())
	fileBatch.Set("FHS.6", b.Get("BHS.6").String())
	fileBatch.Set("FHS.7", b.Get("BHS.7").String())
	fileBatch.Set("FHS.9", "hl7."+name+"."+b.Get("BHS.7").String()+"."+fileBatch.opt.Extension)

	fileBatch.AddBatch(b)
	fileBatch.End()
	if err := fileBatch.CreateFile(name); err != nil {
		return "", err
	}
	return fileBatch.FileName(), nil
}

// createChild builds a Segment from trimmed text (the Batch.createChild).
func (b *Batch) createChild(text string, index int) HL7Node {
	return newSegment(b, strings.TrimSpace(text))
}

// pathCore returns the empty root path (the Batch.pathCore).
func (b *Batch) pathCore() []string { return []string{} }

// writeCore writes a segment path, appending a new segment at index 0 (the
// Batch.writeCore).
func (b *Batch) writeCore(path []string, value string) HL7Node {
	if len(path) == 0 {
		panic(helpers.NewHL7ParserError("Segment name is not defined."))
	}
	segmentName := path[0]
	rest := path[1:]
	return b.writeAtIndex(rest, value, 0, segmentName)
}

// addSegment appends a new segment at the root (the _addSegment).
func (b *Batch) addSegment(path string) *Segment {
	if path == "" {
		panic(helpers.NewHL7ParserError("Missing segment path."))
	}
	prepared := b.preparePath(path)
	if len(prepared) != 1 {
		panic(helpers.NewHL7ParserError("Invalid segment " + path + "."))
	}
	return b.addChild(prepared[0]).(*Segment)
}

// firstSegment returns the first segment named name (the _getFirstSegment).
func (b *Batch) firstSegment(name string) *Segment {
	for _, c := range b.childrenOf() {
		if seg, ok := c.(*Segment); ok && seg.Name() == name {
			return seg
		}
	}
	panic(helpers.NewHL7FatalError("Unable to process _getFirstSegment."))
}
