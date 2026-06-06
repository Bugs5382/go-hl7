package hl7_test

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
	"testing"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/hl7"
)

// createAckMessage builds a v2.1 ACK off an inbound message: it swaps the
// inbound MSH routing fields and echoes MSH.10.
func createAckMessage(t *testing.T, ackType string, message *builder.Message) *builder.Message {
	t.Helper()
	messageBuild := hl7.New(hl7.V2_1)
	messageBuild.BuildMSH(hl7.Props{
		"msh_10": "12345",
		"msh_11": "T",
		"msh_3":  message.Get("MSH.5").String(),
		"msh_4":  message.Get("MSH.6").String(),
		"msh_5":  message.Get("MSH.3").String(),
		"msh_6":  message.Get("MSH.4").String(),
		"msh_9":  "ACK",
	})
	messageBuild.BuildMSA(hl7.Props{
		"msa_1": ackType,
		"msa_2": message.Get("MSH.10").String(),
	})
	msg, err := messageBuild.ToMessage()
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	return msg
}

func TestServerParserCreatesAAAck(t *testing.T) {
	messageString := "MSH|^~\\&|||||20220304102435|ESBCKGRND|SIU^S12|521\r" +
		"SCH||60014711||||Sch|||5|MIN|^^5^20220218153000^20220218153500|ESEOD^CADENCE^EOD^PROCESSING||||ESEOD^CADENCE^EOD^PROCESSING||||ESEOD^CADENCE^EOD^PROCESSING|||||Sch\r" +
		"PID|1||3002505^^^MRN^MRN||CHILD^AMB^^^^^D||20150122|F|||123 STREET^^BROOKLYN^^11233^^L||(718)250-0000^P^H^^^718^2500000~^NET^Internet^cool@gmail.com|||SINGLE||60014711|111-52-5454||One^Mother^^|||||||||N\r" +
		"ZPD|Cent Amer In|MYCH|||||||||||||||||||||N|F\r" +
		"PD1||||9454^KOTHARI^VIPUL^^^^^^PROVID^^^^PROVID\r" +
		"PV1||OUTPATIENT|GGEVAC^^^^^^^^^^EDEP||||||||||||||||60014711|||||||||||||||||||||||||20220218||||||60014711\r" +
		"RGS|1||10008938^MAIN CAMPUS COVID\r" +
		"AIS|1|||||||||Sch\r" +
		"AIG|1||^COVID-19 VACCINE|2^RESOURCE||||20220218153000|0|MIN|5|MIN"

	message, err := builder.NewMessage(builder.MessageOptions{Text: messageString})
	if err != nil {
		t.Fatalf("parse inbound: %v", err)
	}

	ackMessage := createAckMessage(t, "AA", message)
	if got := ackMessage.Get("MSH.9.1").String(); got != "ACK" {
		t.Fatalf("MSH.9.1 = %q, want ACK", got)
	}
	if got := ackMessage.Get("MSA.1").String(); got != "AA" {
		t.Fatalf("MSA.1 = %q, want AA", got)
	}
}
