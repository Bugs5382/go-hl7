package builder

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
	"strconv"
	"strings"
	"time"

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// Message is the top-level HL7 message (the Message extends RootBase). It
// parses text or builds from a MessageHeader.
type Message struct {
	RootBase
	opt builderOptions
}

// NewMessage builds an empty message from a validated MessageHeader, or parses
// the provided text. It mirrors the Message constructor, returning an
// error where the spec throws (Go necessity).
func NewMessage(opts MessageOptions) (*Message, error) {
	opt, err := normalizedClientMessageParserOptions(opts)
	if err != nil {
		return nil, err
	}

	m := &Message{opt: opt}
	m.initRootBase(m, opt)

	if opt.Text != "" {
		totalMsh := 0
		for _, line := range utils.Split(opt.Text, nil) {
			if strings.HasPrefix(line, "MSH") {
				totalMsh++
			}
		}
		if totalMsh > 0 && totalMsh != 1 {
			return nil, helpers.NewHL7FatalError("Multiple MSH segments found. Use Batch.")
		}
	}

	if opt.MessageHeader != nil {
		msh := opt.MessageHeader
		if msh.MSH9_1 == "" || msh.MSH9_2 == "" {
			return nil, helpers.NewHL7FatalError("MSH.9.1 & MSH 9.2 must be defined.")
		}
		if len(msh.MSH9_1) != 3 {
			return nil, helpers.NewHL7FatalError("MSH.9.1 must be 3 characters in length.")
		}
		if len(msh.MSH9_2) != 3 {
			return nil, helpers.NewHL7FatalError("MSH.9.2 must be 3 characters in length.")
		}
		if msh.hasMSH9_3 && (len(msh.MSH9_3) < 3 || len(msh.MSH9_3) > 10) {
			return nil, helpers.NewHL7FatalError("MSH.9.3 must be 3 to 10 characters in length if specified.")
		}
		if msh.hasMSH10 && (len(msh.MSH10) == 0 || len(msh.MSH10) > 199) {
			return nil, helpers.NewHL7FatalError("MSH.10 must be greater than 0 and less than 199 characters.")
		}

		m.Set("MSH.7", utils.CreateHL7Date(time.Now(), m.opt.Date))
		m.Set("MSH.9.1", msh.MSH9_1)
		m.Set("MSH.9.2", msh.MSH9_2)
		if msh.hasMSH9_3 {
			m.Set("MSH.9.3", msh.MSH9_3)
		} else {
			m.Set("MSH.9.3", msh.MSH9_1+"_"+msh.MSH9_2)
		}
		if msh.hasMSH10 {
			m.Set("MSH.10", msh.MSH10)
		} else {
			m.Set("MSH.10", utils.RandomString(20))
		}
		if msh.hasMSH11_1 {
			m.Set("MSH.11.1", msh.MSH11_1)
		} else if msh.hasMSH11 {
			m.Set("MSH.11", msh.MSH11)
		}
		if msh.hasMSH11_2 {
			m.Set("MSH.11.2", msh.MSH11_2)
		}
		m.Set("MSH.12", "2.7")
	}

	return m, nil
}

// AddSegment adds a new segment to the message (the addSegment).
func (m *Message) AddSegment(path string) (*Segment, error) {
	if path == "" {
		return nil, helpers.NewHL7ParserError("Missing segment path.")
	}
	prepared := m.preparePath(path)
	if len(prepared) != 1 {
		return nil, helpers.NewHL7ParserError("Invalid segment " + path + ".")
	}
	child := m.addChild(prepared[0])
	if seg, ok := child.(*Segment); ok {
		return seg, nil
	}
	return nil, helpers.NewHL7ParserError("Invalid segment " + path + ".")
}

// Get resolves a segment or field path (the Message.get).
func (m *Message) Get(path string) HL7Node {
	if path == "" {
		return emptySingleton
	}
	p := m.preparePath(path)
	rv := m.Read(p)
	if rv == nil {
		return emptySingleton
	}
	return rv
}

// GetFirstSegment returns the first child segment (the getFirstSegment).
func (m *Message) GetFirstSegment() *Segment {
	ch := m.childrenOf()
	if len(ch) == 0 {
		return nil
	}
	return ch[0].(*Segment)
}

// GetLastSegment returns the last child segment (the getLastSegment).
func (m *Message) GetLastSegment() *Segment {
	ch := m.childrenOf()
	if len(ch) == 0 {
		return nil
	}
	return ch[len(ch)-1].(*Segment)
}

// Read resolves a split path, returning a SegmentList for a bare segment name
// (the Message.read).
func (m *Message) Read(path []string) HL7Node {
	if len(path) == 0 {
		return emptySingleton
	}
	segmentName := path[0]
	rest := path[1:]
	if len(rest) == 0 {
		var segments []*Segment
		for _, c := range m.childrenOf() {
			if seg, ok := c.(*Segment); ok && seg.Name() == segmentName {
				segments = append(segments, seg)
			}
		}
		if len(segments) > 0 {
			return newSegmentList(m, segments)
		}
	} else {
		seg := m.getFirstSegment(segmentName)
		if seg != nil {
			return seg.Read(rest)
		}
	}
	return emptySingleton
}

// Set writes a value at a segment/field path (the Message.set).
func (m *Message) Set(path string, value any) HL7Node {
	if arr, ok := value.([]any); ok {
		for i, item := range arr {
			m.Set(path+"."+strconv.Itoa(i+1), item)
		}
		return m
	}
	p := m.preparePath(path)
	m.Write(p, m.prepareValue(value))
	return m
}

// ToFile wraps the message in a FileBatch and writes it to disk, returning the
// generated file name (the Message.toFile).
func (m *Message) ToFile(name string, newLine bool, location string, extension string) (string, error) {
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

	fileBatch.Set("FHS.3", m.Get("MSH.3").String())
	fileBatch.Set("FHS.4", m.Get("MSH.4").String())
	fileBatch.Set("FHS.5", m.Get("MSH.5").String())
	fileBatch.Set("FHS.6", m.Get("MSH.6").String())
	fileBatch.Set("FHS.7", m.Get("MSH.7").String())
	fileBatch.Set("FHS.9", "hl7."+name+"."+m.Get("MSH.7").String()+"."+fileBatch.opt.Extension)
	fileBatch.Set("FHS.11", m.Get("MSH.10").String())

	fileBatch.AddMessage(m)
	fileBatch.End()
	if err := fileBatch.CreateFile(name); err != nil {
		return "", err
	}
	return fileBatch.FileName(), nil
}

// TotalSegment counts segments matching name (the totalSegment).
func (m *Message) TotalSegment(name string) int {
	count := 0
	for _, c := range m.childrenOf() {
		if seg, ok := c.(*Segment); ok && seg.Name() == name {
			count++
		}
	}
	return count
}

// createChild builds a Segment from trimmed text (the Message.createChild).
func (m *Message) createChild(text string, index int) HL7Node {
	return newSegment(m, strings.TrimSpace(text))
}

// pathCore returns the empty root path (the Message.pathCore).
func (m *Message) pathCore() []string { return []string{} }

// writeCore writes a segment path, appending a new segment when absent (the
// Message.writeCore).
func (m *Message) writeCore(path []string, value string) HL7Node {
	segmentName := path[0]
	rest := path[1:]
	index, found := m.getFirstSegmentIndex(segmentName)
	if !found {
		index = len(m.childrenOf())
	}
	return m.writeAtIndex(rest, value, index, segmentName)
}

// getFirstSegment returns the first segment named name (the _getFirstSegment).
func (m *Message) getFirstSegment(name string) *Segment {
	for _, c := range m.childrenOf() {
		if seg, ok := c.(*Segment); ok && seg.Name() == name {
			return seg
		}
	}
	return nil
}

// getFirstSegmentIndex returns the index of the first segment named name
// (the _getFirstSegmentIndex).
func (m *Message) getFirstSegmentIndex(name string) (int, bool) {
	for i, c := range m.childrenOf() {
		if seg, ok := c.(*Segment); ok && seg.Name() == name {
			return i, true
		}
	}
	return 0, false
}
