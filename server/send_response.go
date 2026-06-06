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
	"slices"
	"strings"
	"time"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/modules"
	"github.com/Bugs5382/go-hl7/client/utils"
	srvutils "github.com/Bugs5382/go-hl7/server/utils"
)

// ResponseSender is the ACK-sending contract a handler's response satisfies. It
// mirrors the ISendRequest / the BaseSendResponse | SendResponse union the
// InboundHandler receives, so a caller may supply a custom response type.
type ResponseSender interface {
	// GetAckMessage returns the ACK that was sent, or nil before one is sent.
	GetAckMessage() *builder.Message
	// GetCodec returns the MLLP codec backing this response.
	GetCodec() *modules.MLLPCodec
	// GetSocket returns the client socket.
	GetSocket() net.Conn
	// SendResponse builds and sends the standard AA/AE/AR(+CA/CR/CE) ACK.
	SendResponse(ackType string) error
	// SendCustomResponse sends a fully-formed ACK (Message or raw string)
	// verbatim.
	SendCustomResponse(message any) error
}

// SendResponse builds and sends an ACK back to the client. It folds the
// BaseSendResponse and SendResponse into one type: the base codec/socket/ack
// accessors plus sendResponse/sendCustomResponse and the ACK builder. It is an
// EventEmitter (response.sent) via the embedded EventEmitter.
type SendResponse struct {
	EventEmitter
	ack          *builder.Message
	codec        *modules.MLLPCodec
	message      *builder.Message
	mshOverrides map[string]srvutils.MSHOverride
	socket       net.Conn
}

// compile-time guard that SendResponse satisfies ResponseSender.
var _ ResponseSender = (*SendResponse)(nil)

// NewSendResponse constructs a SendResponse over the client socket and inbound
// message, mirroring the SendResponse/BaseSendResponse constructor.
func NewSendResponse(socket net.Conn, message *builder.Message, mshOverrides map[string]srvutils.MSHOverride) *SendResponse {
	return &SendResponse{
		codec:        modules.NewMLLPCodec(""),
		message:      message,
		mshOverrides: mshOverrides,
		socket:       socket,
	}
}

// GetAckMessage returns the sent ACK (nil before send), mirroring the
// getAckMessage.
func (r *SendResponse) GetAckMessage() *builder.Message { return r.ack }

// GetCodec returns the MLLP codec, mirroring the getCodec.
func (r *SendResponse) GetCodec() *modules.MLLPCodec { return r.codec }

// GetSocket returns the client socket, mirroring the getSocket.
func (r *SendResponse) GetSocket() net.Conn { return r.socket }

// SendResponse builds the ACK for the given code and writes it to the client.
// It mirrors the sendResponse: an HL7ServerError from the MSA.1 validation
// is propagated; any other failure falls back to an AE/Z99 ACK. On success it
// emits response.sent.
func (r *SendResponse) SendResponse(ackType string) error {
	ack, err := r.createAckMessage(ackType, r.message)
	if err != nil {
		var serverErr *srvutils.HL7ServerError
		if errors.As(err, &serverErr) {
			return serverErr
		}
		ack, err = r.createAEAckMessage()
		if err != nil {
			return err
		}
	}
	r.ack = ack

	if err := r.codec.SendMessage(r.socket, ack.String()); err != nil {
		return err
	}
	r.emit("response.sent")
	return nil
}

// SendCustomResponse sends a fully-formed ACK verbatim (no validation, swapping,
// or overrides). message is a *builder.Message or a raw HL7 string. It mirrors
// the sendCustomResponse.
func (r *SendResponse) SendCustomResponse(message any) error {
	var ack *builder.Message
	switch v := message.(type) {
	case *builder.Message:
		ack = v
	case string:
		parsed, err := builder.NewMessage(builder.MessageOptions{Text: v})
		if err != nil {
			return err
		}
		ack = parsed
	default:
		return srvutils.NewHL7ServerError("sendCustomResponse accepts a Message or a string.")
	}

	r.ack = ack
	if err := r.codec.SendMessage(r.socket, ack.String()); err != nil {
		return err
	}
	r.emit("response.sent")
	return nil
}

// createAckMessage builds the standard ACK: validate the code against the
// inbound MSH.12 version, swap sender/receiver, echo MSH.11 and the control id,
// set MSH.9 to ACK^<event>, then apply any MSH overrides. Mirrors the
// _createAckMessage.
func (r *SendResponse) createAckMessage(ackType string, message *builder.Message) (*builder.Message, error) {
	spec := message.Get("MSH.12").String()
	if err := validateMSA1(spec, ackType); err != nil {
		return nil, err
	}

	sendApp := message.Get("MSH.5").String()
	sendFac := message.Get("MSH.6").String()
	recvApp := message.Get("MSH.3").String()
	recvFac := message.Get("MSH.4").String()
	processingID := message.Get("MSH.11").String()
	origControlID := message.Get("MSH.10").String()

	eventCode := message.Get("MSH.9.2").String()
	msh9 := "ACK"
	if eventCode != "" {
		msh9 = "ACK^" + eventCode
	}

	text := strings.Join([]string{
		"MSH|^~\\&|" + sendApp + "|" + sendFac + "|" + recvApp + "|" + recvFac + "|" + utils.CreateHL7Date(time.Now(), "") + "||" + msh9 + "|" + utils.RandomString(20) + "|" + processingID + "|" + spec,
		"MSA|" + ackType + "|" + origControlID,
	}, "\r")

	ackMessage, err := builder.NewMessage(builder.MessageOptions{Text: text})
	if err != nil {
		return nil, err
	}

	for path, override := range r.mshOverrides {
		var value string
		if override.Func != nil {
			value = override.Func(message)
		} else {
			value = override.String
		}
		ackMessage.Set("MSH."+path, value)
	}

	return ackMessage, nil
}

// createAEAckMessage builds the AE/Z99 fallback ACK, mirroring the
// _createAEAckMessage.
func (r *SendResponse) createAEAckMessage() (*builder.Message, error) {
	text := strings.Join([]string{
		"MSH|^~\\&|||||" + utils.CreateHL7Date(time.Now(), "") + "||ACK^Z99^ACK|" + utils.RandomString(20) + "|P|2.7",
		"MSA|AE|",
	}, "\r")
	return builder.NewMessage(builder.MessageOptions{Text: text})
}

// validateMSA1 checks an ACK code against the inbound version, mirroring the
// _validateMSA1: 2.1 allows AA/AR/AE only; later versions add CA/CR/CE.
func validateMSA1(spec, ackType string) error {
	switch spec {
	case "2.1":
		if !slices.Contains(srvutils.MSA1Valuesv21, ackType) {
			return srvutils.NewHL7ServerError("Invalid MSA.1 value: " + ackType + " for HL7 version 2.1")
		}
	default:
		if !slices.Contains(srvutils.MSA1Valuesv21, ackType) && !slices.Contains(srvutils.MSA1Valuesv2x, ackType) {
			return srvutils.NewHL7ServerError("Invalid MSA.1 value: " + ackType + " for HL7 version " + spec)
		}
	}
	return nil
}
