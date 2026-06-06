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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// FileBatch is the FHS-rooted file-batch builder. It collects
// Message and Batch children under a File Header Segment (FHS) and closes with a
// File Trailing Segment (FTS), and can serialize the tree to disk.
type FileBatch struct {
	RootBase
	opt           fileBuilderOptions
	fileName      string
	lines         []string
	batchCount    int
	messagesCount int
}

// NewFileBatch builds an empty file batch (writing FHS.7 with the current
// date), or parses the text/file/buffer into the MSH lines it carries. It
// returns an error on invalid input.
func NewFileBatch(opts FileOptions) (*FileBatch, error) {
	opt, err := normalizedClientFileParserOptions(opts)
	if err != nil {
		return nil, err
	}

	f := &FileBatch{opt: opt}
	f.initRootBase(f, opt.builderOptions)

	if opt.Text != "" {
		for _, line := range utils.Split(opt.Text, nil) {
			if strings.HasPrefix(line, "MSH") {
				f.lines = append(f.lines, line)
			}
		}
	} else {
		f.Set("FHS.7", utils.CreateHL7Date(time.Now(), f.opt.Date))
	}

	return f, nil
}

// AddMessage adds a Message to the file batch (the message branch of the
// FileBatch.add). When a Batch already exists, the message is routed into the
// first Batch and the BTS counter bumped, since an MSH may not sit outside an
// open BHS.
func (f *FileBatch) AddMessage(message *Message) {
	f.setDirty()
	if f.batchCount >= 1 {
		batch := f.firstBatch()
		seg := batch.GetFirstSegment("BTS")
		seg.Set("1", batch.messagesCount+1)
		batch.Add(message, batch.messagesCount+1)
	} else {
		f.messagesCount = f.messagesCount + 1
		f.children = append(f.childrenOf(), message)
	}
}

// AddBatch adds a Batch to the file batch (the batch branch of the
// FileBatch.add). It returns an HL7ParserError when messages were already added
// individually, since an MSH may not sit both inside a batch and loose in the
// same file.
func (f *FileBatch) AddBatch(batch *Batch) error {
	f.setDirty()
	if f.messagesCount >= 1 {
		return helpers.NewHL7ParserError("Unable to add a batch segment, since there is already messages added individually.")
	}
	f.batchCount = f.batchCount + 1
	f.children = append(f.childrenOf(), batch)
	return nil
}

// CreateFile writes the framed bytes to disk under the configured Location. It
// is a no-op when no Location was set.
func (f *FileBatch) CreateFile(name string) error {
	fhsDate := f.Get("FHS.7").String()

	if name == "" {
		return helpers.NewHL7FatalError("Missing file name.")
	}
	if helpers.NameFormat.MatchString(name) {
		return helpers.NewHL7FatalError("name must not contain certain characters: `!@#$%^&*()+\\-=\\[\\]{};':\"\\\\|,.<>\\/?~.")
	}

	if f.opt.Location != "" {
		if _, err := os.Stat(f.opt.Location); os.IsNotExist(err) {
			if err := os.MkdirAll(f.opt.Location, 0o755); err != nil {
				return helpers.NewHL7FatalError(err.Error())
			}
		}

		f.fileName = "hl7." + name + "." + fhsDate + "." + f.opt.Extension

		file, err := os.OpenFile(filepath.Join(f.opt.Location, f.fileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return helpers.NewHL7FatalError(err.Error())
		}
		defer func() { _ = file.Close() }()
		if _, err := file.WriteString(f.String()); err != nil {
			return helpers.NewHL7FatalError(err.Error())
		}
	}
	return nil
}

// End closes the file batch, appending the FTS trailer with the batch+message
// count.
func (f *FileBatch) End() {
	segment := f.addSegment("FTS")
	segment.Set("1", f.batchCount+f.messagesCount)
}

// FileName returns the generated file name.
func (f *FileBatch) FileName() string { return f.fileName }

// Text returns the normalized source text the file batch was parsed from.
func (f *FileBatch) Text() string { return f.opt.Text }

// Get resolves an FHS segment or field path.
func (f *FileBatch) Get(path string) HL7Node {
	if path == "" {
		return emptySingleton
	}
	p := f.preparePath(path)
	rv := f.Read(p)
	if rv == nil {
		return emptySingleton
	}
	return rv
}

// Messages parses the carried text into individual Message nodes, one per MSH
// line. It panics with HL7FatalError when there is no
// file text to parse.
func (f *FileBatch) Messages() []*Message {
	if f.lines != nil && f.opt.NewLine != "" {
		messages := make([]*Message, 0, len(f.lines))
		for _, line := range f.lines {
			text := strings.TrimSuffix(line, f.opt.NewLine)
			m, err := NewMessage(MessageOptions{Text: text})
			if err != nil {
				panic(err)
			}
			messages = append(messages, m)
		}
		return messages
	}
	panic(helpers.NewHL7FatalError("No messages inside file segment."))
}

// Read resolves a split path, returning a SegmentList for a bare segment name.
func (f *FileBatch) Read(path []string) HL7Node {
	if len(path) == 0 {
		panic(helpers.NewHL7FatalError("Unable to process the read function correctly."))
	}
	segmentName := path[0]
	rest := path[1:]
	if len(rest) == 0 {
		var segments []*Segment
		for _, c := range f.childrenOf() {
			if seg, ok := c.(*Segment); ok && seg.Name() == segmentName {
				segments = append(segments, seg)
			}
		}
		if len(segments) > 0 {
			return newSegmentList(f, segments)
		}
	} else {
		seg := f.firstSegment(segmentName)
		if seg != nil {
			return seg.Read(rest)
		}
	}
	panic(helpers.NewHL7FatalError("Unable to process the read function correctly."))
}

// Set writes a value at a file-batch segment/field path.
func (f *FileBatch) Set(path string, value any) HL7Node {
	if arr, ok := value.([]any); ok {
		for i, item := range arr {
			f.Set(path+"."+strconv.Itoa(i+1), item)
		}
		return f
	}
	p := f.preparePath(path)
	f.Write(p, f.prepareValue(value))
	return f
}

// Start (re)writes FHS.7 with the current date.
func (f *FileBatch) Start() {
	f.Set("FHS.7", utils.CreateHL7Date(time.Now(), ""))
}

// createChild builds a Segment from trimmed text.
func (f *FileBatch) createChild(text string, index int) HL7Node {
	return newSegment(f, strings.TrimSpace(text))
}

// pathCore returns the empty root path.
func (f *FileBatch) pathCore() []string { return []string{} }

// writeCore writes a segment path, appending a new segment at index 0.
func (f *FileBatch) writeCore(path []string, value string) HL7Node {
	if len(path) == 0 {
		panic(helpers.NewHL7ParserError("Segment name is not defined."))
	}
	segmentName := path[0]
	rest := path[1:]
	return f.writeAtIndex(rest, value, 0, segmentName)
}

// addSegment appends a new segment at the root.
func (f *FileBatch) addSegment(path string) *Segment {
	if path == "" {
		panic(helpers.NewHL7ParserError("Missing segment path."))
	}
	prepared := f.preparePath(path)
	if len(prepared) != 1 {
		panic(helpers.NewHL7ParserError("\"Invalid segment " + path + ".\""))
	}
	return f.addChild(prepared[0]).(*Segment)
}

// firstBatch returns the first Batch child.
func (f *FileBatch) firstBatch() *Batch {
	for _, c := range f.childrenOf() {
		if batch, ok := c.(*Batch); ok {
			return batch
		}
	}
	panic(helpers.NewHL7FatalError("Unable to process _getFirstBatch."))
}

// firstSegment returns the first segment named name.
func (f *FileBatch) firstSegment(name string) *Segment {
	for _, c := range f.childrenOf() {
		if seg, ok := c.(*Segment); ok && seg.Name() == name {
			return seg
		}
	}
	panic(helpers.NewHL7ParserError("Unable to process _getFirstSegment."))
}
