package httpadapter

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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/client"
)

const (
	sampleMessage = "MSH|^~\\&|MY_APP|MY_FAC|EPIC|HOSP|20240101||ADT^A01|MSG00001|P|2.7"
	sampleACK     = "MSH|^~\\&|EPIC|HOSP|MY_APP|MY_FAC|20240101||ACK^A01|RESP1|P|2.7\rMSA|AA|MSG00001"
)

// fakeSender implements Sender and, on SendMessage, feeds a canned ACK back
// through the adapter's ack handler to simulate a server reply.
type fakeSender struct {
	handler   client.OutboundHandler
	ack       *client.InboundResponse
	sendErr   error
	connected bool
	port      int
}

func (f *fakeSender) SendMessage(_ client.MessageItem) error {
	if f.sendErr != nil {
		return f.sendErr
	}
	if f.handler != nil && f.ack != nil {
		_ = f.handler(f.ack)
	}
	return nil
}

func (f *fakeSender) IsConnected() bool { return f.connected }

func (f *fakeSender) GetPort() int { return f.port }

func newACK(t *testing.T) *client.InboundResponse {
	t.Helper()
	ack, err := client.NewInboundResponse(sampleACK)
	if err != nil {
		t.Fatalf("build ACK: %v", err)
	}
	return ack
}

func TestSendHandlerReturnsACK(t *testing.T) {
	a := New()
	ack := newACK(t)
	fake := &fakeSender{handler: a.AckHandler(), ack: ack, connected: true, port: 3000}
	a.Bind(fake)

	req := httptest.NewRequest(http.MethodPost, "/hl7", strings.NewReader(sampleMessage))
	rec := httptest.NewRecorder()
	a.SendHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body: %q)", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != ContentTypeHL7 {
		t.Errorf("content-type = %q, want %q", ct, ContentTypeHL7)
	}
	want := ack.GetMessage().String()
	if got := rec.Body.String(); got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestSendHandlerRejectsNonPost(t *testing.T) {
	a := New()
	a.Bind(&fakeSender{connected: true})

	req := httptest.NewRequest(http.MethodGet, "/hl7", nil)
	rec := httptest.NewRecorder()
	a.SendHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}

func TestSendHandlerRejectsInvalidBody(t *testing.T) {
	a := New()
	a.Bind(&fakeSender{handler: a.AckHandler(), ack: newACK(t), connected: true})

	req := httptest.NewRequest(http.MethodPost, "/hl7", strings.NewReader("not an hl7 message"))
	rec := httptest.NewRecorder()
	a.SendHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestSendHandlerNotBound(t *testing.T) {
	a := New()

	req := httptest.NewRequest(http.MethodPost, "/hl7", strings.NewReader(sampleMessage))
	rec := httptest.NewRecorder()
	a.SendHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestSendHandlerTimeout(t *testing.T) {
	a := New(WithTimeout(50 * time.Millisecond))
	// No ack wired, so SendMessage succeeds but no ACK is ever delivered.
	a.Bind(&fakeSender{connected: true})

	req := httptest.NewRequest(http.MethodPost, "/hl7", strings.NewReader(sampleMessage))
	rec := httptest.NewRecorder()
	a.SendHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want 504", rec.Code)
	}
}

func TestSendHandlerSendError(t *testing.T) {
	a := New()
	a.Bind(&fakeSender{sendErr: errors.New("boom"), connected: true})

	req := httptest.NewRequest(http.MethodPost, "/hl7", strings.NewReader(sampleMessage))
	rec := httptest.NewRecorder()
	a.SendHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}

func TestHealthHandlerConnected(t *testing.T) {
	a := New()
	a.Bind(&fakeSender{connected: true, port: 3000})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	a.HealthHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body health
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !body.Connected || body.Port != 3000 {
		t.Errorf("body = %+v, want connected=true port=3000", body)
	}
}

func TestHealthHandlerUnavailable(t *testing.T) {
	a := New()
	a.Bind(&fakeSender{connected: false, port: 3000})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	a.HealthHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}

func TestHealthHandlerNotBound(t *testing.T) {
	a := New()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	a.HealthHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
}
