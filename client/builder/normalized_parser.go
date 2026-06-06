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
	"os"
	"strings"

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/modules"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// MessageHeader carries the required header values for building a message. It
// mirrors the messageHeader option fields.
type MessageHeader struct {
	MSH9_1  string
	MSH9_2  string
	MSH9_3  string
	MSH10   string
	MSH11   string
	MSH11_1 string
	MSH11_2 string

	// hasMSH9_3, hasMSH10, hasMSH11, hasMSH11_1, hasMSH11_2 distinguish an unset
	// field from an empty one, mirroring the `undefined` checks.
	hasMSH9_3  bool
	hasMSH10   bool
	hasMSH11   bool
	hasMSH11_1 bool
	hasMSH11_2 bool
}

// WithMSH9_3 sets the optional MSH.9.3 (message structure), marking it present
// so the constructor writes it instead of deriving MSH.9.1_MSH.9.2.
func (h *MessageHeader) WithMSH9_3(v string) *MessageHeader {
	h.MSH9_3 = v
	h.hasMSH9_3 = true
	return h
}

// WithMSH10 sets the optional MSH.10 (message control id), marking it present so
// the constructor writes it instead of a random string.
func (h *MessageHeader) WithMSH10(v string) *MessageHeader {
	h.MSH10 = v
	h.hasMSH10 = true
	return h
}

// WithMSH11 sets the optional MSH.11 (processing id), marking it present.
func (h *MessageHeader) WithMSH11(v string) *MessageHeader {
	h.MSH11 = v
	h.hasMSH11 = true
	return h
}

// WithMSH11_1 sets the optional MSH.11.1 (processing id), marking it present. It
// takes precedence over MSH.11.
func (h *MessageHeader) WithMSH11_1(v string) *MessageHeader {
	h.MSH11_1 = v
	h.hasMSH11_1 = true
	return h
}

// WithMSH11_2 sets the optional MSH.11.2 (processing mode), marking it present.
func (h *MessageHeader) WithMSH11_2(v string) *MessageHeader {
	h.MSH11_2 = v
	h.hasMSH11_2 = true
	return h
}

// MessageOptions configures a Message. It mirrors the reference's
// ClientBuilderMessageOptions: parse Text, or build from MessageHeader, with
// optional separator and date overrides.
type MessageOptions struct {
	Text          string
	MessageHeader *MessageHeader
	Date          string
	NewLine       string

	SeparatorField        string
	SeparatorComponent    string
	SeparatorRepetition   string
	SeparatorEscape       string
	SeparatorSubComponent string
}

// builderOptions is the fully-normalized option set the tree builds against.
type builderOptions struct {
	Text                  string
	NewLine               string
	SeparatorField        string
	SeparatorComponent    string
	SeparatorRepetition   string
	SeparatorEscape       string
	SeparatorSubComponent string
	Date                  string
	MessageHeader         *MessageHeader
}

// defaultBuilderOptions returns the DEFAULT_CLIENT_BUILDER_OPTS.
func defaultBuilderOptions() builderOptions {
	return builderOptions{
		NewLine:               "\r",
		SeparatorComponent:    "^",
		SeparatorEscape:       "\\",
		SeparatorField:        "|",
		SeparatorRepetition:   "~",
		SeparatorSubComponent: "&",
		Text:                  "",
	}
}

// normalizedClientMessageParserOptions applies the message option
// normalization and validation (normalizedClientMessageParserOptions).
func normalizedClientMessageParserOptions(raw MessageOptions) (builderOptions, error) {
	p := defaultBuilderOptions()
	p.MessageHeader = raw.MessageHeader
	if raw.Text != "" {
		p.Text = raw.Text
	}
	if raw.NewLine != "" {
		p.NewLine = raw.NewLine
	}
	if raw.Date != "" {
		p.Date = raw.Date
	}
	if raw.SeparatorField != "" {
		p.SeparatorField = raw.SeparatorField
	}
	if raw.SeparatorComponent != "" {
		p.SeparatorComponent = raw.SeparatorComponent
	}
	if raw.SeparatorRepetition != "" {
		p.SeparatorRepetition = raw.SeparatorRepetition
	}
	if raw.SeparatorEscape != "" {
		p.SeparatorEscape = raw.SeparatorEscape
	}
	if raw.SeparatorSubComponent != "" {
		p.SeparatorSubComponent = raw.SeparatorSubComponent
	}

	if p.Text != "" && len(p.Text) >= 3 && p.Text[:3] != "MSH" {
		return p, helpers.NewHL7FatalError("text must begin with the MSH segment.")
	}
	if p.NewLine == `\r` || p.NewLine == `\n` {
		return p, helpers.NewHL7FatalError("newLine must be \r or \n")
	}
	if p.Date != "8" && p.Date != "12" && p.Date != "14" {
		p.Date = "14"
	}

	if p.Text == "" && p.MessageHeader != nil {
		p.Text = "MSH" + p.SeparatorField + p.SeparatorComponent + p.SeparatorRepetition + p.SeparatorEscape + p.SeparatorSubComponent
	} else if p.Text != "" {
		plan := modules.NewParserPlan(sliceStr(p.Text, 3, 8))
		if strings.Contains(p.Text, "\r") {
			p.NewLine = "\r"
		} else {
			p.NewLine = "\n"
		}
		p.SeparatorField = plan.SeparatorField
		p.SeparatorComponent = plan.SeparatorComponent
		p.SeparatorRepetition = plan.SeparatorRepetition
		p.SeparatorEscape = plan.SeparatorEscape
		p.SeparatorSubComponent = plan.SeparatorSubComponent
		p.Text = strings.TrimSpace(p.Text)
	} else {
		p.Text = ""
	}

	return p, nil
}

// BatchOptions configures a Batch. It mirrors the ClientBuilderMessageOptions
// as consumed by the Batch constructor: parse Text (a BHS- or multi-MSH body),
// or build an empty BHS, with optional separator and date overrides.
type BatchOptions struct {
	Text          string
	MessageHeader *MessageHeader
	Date          string
	NewLine       string

	SeparatorField        string
	SeparatorComponent    string
	SeparatorRepetition   string
	SeparatorEscape       string
	SeparatorSubComponent string
}

// FileOptions configures a FileBatch. It mirrors the ClientBuilderFileOptions:
// parse Text (an FHS body), read from FullFilePath or FileBuffer, or build an
// empty FHS, with an Extension and Location for createFile.
type FileOptions struct {
	Text         string
	Date         string
	NewLine      string
	Extension    string
	Location     string
	FullFilePath string
	FileBuffer   []byte

	SeparatorField        string
	SeparatorComponent    string
	SeparatorRepetition   string
	SeparatorEscape       string
	SeparatorSubComponent string
}

// normalizedClientBatchParserOptions applies the batch option normalization and
// validation (normalizedClientBatchParserOptions). The batch body must begin
// with BHS or a multi-MSH stream; a single MSH is rejected.
func normalizedClientBatchParserOptions(raw BatchOptions) (builderOptions, error) {
	p := defaultBuilderOptions()
	p.MessageHeader = raw.MessageHeader
	if raw.Text != "" {
		p.Text = raw.Text
	}
	if raw.NewLine != "" {
		p.NewLine = raw.NewLine
	}
	if raw.Date != "" {
		p.Date = raw.Date
	}
	if raw.SeparatorField != "" {
		p.SeparatorField = raw.SeparatorField
	}
	if raw.SeparatorComponent != "" {
		p.SeparatorComponent = raw.SeparatorComponent
	}
	if raw.SeparatorRepetition != "" {
		p.SeparatorRepetition = raw.SeparatorRepetition
	}
	if raw.SeparatorEscape != "" {
		p.SeparatorEscape = raw.SeparatorEscape
	}
	if raw.SeparatorSubComponent != "" {
		p.SeparatorSubComponent = raw.SeparatorSubComponent
	}

	if p.Text != "" && sliceStr(p.Text, 0, 3) != "BHS" && sliceStr(p.Text, 0, 3) != "MSH" {
		return p, helpers.NewHL7FatalError("text must begin with the BHS or MSH segment.")
	}
	if p.Text != "" && sliceStr(p.Text, 0, 3) == "MSH" && !utils.IsBatch(p.Text) {
		return p, helpers.NewHL7FatalError("Unable to process a single MSH as a batch. Use Message.")
	}
	if p.NewLine == `\r` || p.NewLine == `\n` {
		return p, helpers.NewHL7FatalError("newLine must be \r or \n")
	}
	if p.Date != "8" && p.Date != "12" && p.Date != "14" {
		p.Date = "14"
	}

	if p.Text == "" {
		p.Text = "BHS" + p.SeparatorField + p.SeparatorComponent + p.SeparatorRepetition + p.SeparatorEscape + p.SeparatorSubComponent
	} else {
		plan := modules.NewParserPlan(sliceStr(p.Text, 3, 8))
		if strings.Contains(p.Text, "\r") {
			p.NewLine = "\r"
		} else {
			p.NewLine = "\n"
		}
		p.SeparatorField = plan.SeparatorField
		p.SeparatorComponent = plan.SeparatorComponent
		p.SeparatorRepetition = plan.SeparatorRepetition
		p.SeparatorEscape = plan.SeparatorEscape
		p.SeparatorSubComponent = plan.SeparatorSubComponent
	}

	return p, nil
}

// fileBuilderOptions is the fully-normalized file-batch option set. It extends
// builderOptions with the file Extension/Location used by createFile.
type fileBuilderOptions struct {
	builderOptions
	Extension string
	Location  string
}

// defaultFileBuilderOptions returns the DEFAULT_CLIENT_FILE_OPTS merged over the
// DEFAULT_CLIENT_BUILDER_OPTS.
func defaultFileBuilderOptions() fileBuilderOptions {
	return fileBuilderOptions{
		builderOptions: defaultBuilderOptions(),
		Extension:      "hl7",
		Location:       "",
	}
}

// normalizedClientFileParserOptions applies the file-batch option normalization
// and validation (normalizedClientFileParserOptions). It reads FullFilePath or
// FileBuffer into Text (normalizing \n to \r) and validates the FHS prefix and
// the 3-character extension.
func normalizedClientFileParserOptions(raw FileOptions) (fileBuilderOptions, error) {
	p := defaultFileBuilderOptions()
	if raw.Text != "" {
		p.Text = raw.Text
	}
	if raw.NewLine != "" {
		p.NewLine = raw.NewLine
	}
	if raw.Date != "" {
		p.Date = raw.Date
	}
	if raw.Extension != "" {
		p.Extension = raw.Extension
	}
	if raw.Location != "" {
		p.Location = raw.Location
	}
	if raw.SeparatorField != "" {
		p.SeparatorField = raw.SeparatorField
	}
	if raw.SeparatorComponent != "" {
		p.SeparatorComponent = raw.SeparatorComponent
	}
	if raw.SeparatorRepetition != "" {
		p.SeparatorRepetition = raw.SeparatorRepetition
	}
	if raw.SeparatorEscape != "" {
		p.SeparatorEscape = raw.SeparatorEscape
	}
	if raw.SeparatorSubComponent != "" {
		p.SeparatorSubComponent = raw.SeparatorSubComponent
	}

	if p.Text != "" && sliceStr(p.Text, 0, 3) != "FHS" {
		return p, helpers.NewHL7FatalError("text must begin with the FHS segment.")
	}
	if p.NewLine == `\r` || p.NewLine == `\n` {
		return p, helpers.NewHL7FatalError("newLine must be \r or \n")
	}
	if p.Extension != "" && len(p.Extension) != 3 {
		return p, helpers.NewHL7FatalError("The extension for file save must be 3 characters long.")
	}
	if raw.FullFilePath != "" && raw.FileBuffer != nil {
		return p, helpers.NewHL7FatalError("You can not have specified a file path and a buffer. Please choose one or the other.")
	}
	if p.Date != "8" && p.Date != "12" && p.Date != "14" {
		p.Date = "14"
	}

	if raw.FullFilePath != "" && raw.FileBuffer == nil {
		data, err := os.ReadFile(raw.FullFilePath)
		if err != nil {
			return p, helpers.NewHL7FatalError(err.Error())
		}
		p.Text = strings.ReplaceAll(string(data), "\n", "\r")
	} else if raw.FullFilePath == "" && raw.FileBuffer != nil {
		p.Text = strings.ReplaceAll(string(raw.FileBuffer), "\n", "\r")
	}

	if p.Text == "" {
		p.Text = "FHS" + p.SeparatorField + p.SeparatorComponent + p.SeparatorRepetition + p.SeparatorEscape + p.SeparatorSubComponent
	} else {
		plan := modules.NewParserPlan(sliceStr(p.Text, 3, 8))
		if strings.Contains(p.Text, "\r") {
			p.NewLine = "\r"
		} else {
			p.NewLine = "\n"
		}
		p.SeparatorField = plan.SeparatorField
		p.SeparatorComponent = plan.SeparatorComponent
		p.SeparatorRepetition = plan.SeparatorRepetition
		p.SeparatorEscape = plan.SeparatorEscape
		p.SeparatorSubComponent = plan.SeparatorSubComponent
	}

	return p, nil
}

// sliceStr returns s[a:b] clamped to the string bounds.
func sliceStr(s string, a, b int) string {
	if a > len(s) {
		return ""
	}
	if b > len(s) {
		b = len(s)
	}
	return s[a:b]
}
