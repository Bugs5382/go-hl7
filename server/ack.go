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

// This file implements the HL7 v2 original/enhanced acknowledgement negotiation
// requested by issue #24. MSH-15 (Accept Acknowledgment Type) governs the
// immediate accept (commit) ACK; MSH-16 (Application Acknowledgment Type)
// governs the application ACK. Both draw their send condition from HL7 table
// 0155 (AL always, NE never, ER on error/reject, SU on success). When both
// fields are empty the message is in original mode: a single application ACK is
// always returned, preserving the pre-#24 behavior.

// AckOutcome is the application's processing outcome for an inbound message. It
// selects the accept/error/reject variant within an ACK code family (CA/CE/CR
// for the accept ACK, AA/AE/AR for the application ACK).
type AckOutcome int

const (
	// OutcomeSuccess maps to CA (accept) or AA (application).
	OutcomeSuccess AckOutcome = iota
	// OutcomeError maps to CE (accept) or AE (application).
	OutcomeError
	// OutcomeReject maps to CR (accept) or AR (application).
	OutcomeReject
)

// AckClass selects the ACK code family: the accept (commit) ACK negotiated by
// MSH-15, or the application ACK negotiated by MSH-16 (and used in original
// mode).
type AckClass int

const (
	// AckAccept is the accept (commit) ACK family: CA/CE/CR.
	AckAccept AckClass = iota
	// AckApplication is the application ACK family: AA/AE/AR.
	AckApplication
)

// NegotiateAck computes the ACK code and whether it should be sent for a given
// acknowledgement-type field value (MSH-15 or MSH-16, HL7 table 0155:
// AL/NE/ER/SU, or empty), ACK class, and processing outcome. An empty field is
// treated as AL (always); an unrecognized code is treated conservatively as AL.
func NegotiateAck(ackTypeField string, class AckClass, outcome AckOutcome) (code string, send bool) {
	return ackCode(class, outcome), sendCondition(ackTypeField, outcome)
}

// sendCondition applies the HL7 table 0155 send rule for the given field value
// and outcome.
func sendCondition(field string, outcome AckOutcome) bool {
	switch field {
	case "NE": // never
		return false
	case "ER": // only on error/reject
		return outcome == OutcomeError || outcome == OutcomeReject
	case "SU": // only on success
		return outcome == OutcomeSuccess
	default: // "", "AL", or unknown -> always
		return true
	}
}

// ackCode maps an ACK class and outcome to its MSA-1 code.
func ackCode(class AckClass, outcome AckOutcome) string {
	if class == AckAccept {
		switch outcome {
		case OutcomeError:
			return "CE"
		case OutcomeReject:
			return "CR"
		default:
			return "CA"
		}
	}
	switch outcome {
	case OutcomeError:
		return "AE"
	case OutcomeReject:
		return "AR"
	default:
		return "AA"
	}
}

// enhanced reports whether the inbound message requested enhanced-mode
// acknowledgement, i.e. MSH-15 and/or MSH-16 is populated.
func (r *SendResponse) enhanced() bool {
	return r.message.Get("MSH.15").String() != "" || r.message.Get("MSH.16").String() != ""
}

// SendAck sends the acknowledgement dictated by the inbound message's HL7
// acknowledgement mode:
//
//   - Original mode (MSH-15 and MSH-16 both empty): one application ACK
//     (AA/AE/AR for success/error/reject), always sent. This preserves the
//     pre-#24 behavior of SendResponse.
//   - Enhanced mode (MSH-15 and/or MSH-16 populated): the immediate accept
//     (commit) ACK (CA/CE/CR) governed by MSH-15's send condition (table 0155:
//     AL always, NE never, ER on error/reject, SU on success). When MSH-15
//     suppresses it, no ACK is written.
//
// It reports whether an ACK was written. MSA-1 is validated against the inbound
// MSH-12 version (CA/CE/CR require HL7 >= 2.2); a validation failure is returned
// and no ACK is sent. The later, optional application ACK negotiated by MSH-16
// is available via SendApplicationAck.
func (r *SendResponse) SendAck(outcome AckOutcome) (bool, error) {
	field := ""
	class := AckApplication
	if r.enhanced() {
		field = r.message.Get("MSH.15").String()
		class = AckAccept
	}
	code, send := NegotiateAck(field, class, outcome)
	if !send {
		return false, nil
	}
	if err := r.SendResponse(code); err != nil {
		return false, err
	}
	return true, nil
}

// SendApplicationAck sends the application-level ACK (AA/AE/AR) governed by
// MSH-16's send condition (table 0155). In enhanced mode this is the deferred
// application acknowledgement that follows the accept ACK; a handler that
// processes synchronously may send it immediately after SendAck. When MSH-16
// suppresses it (NE, or ER/SU not matching outcome), no ACK is written. An
// empty MSH-16 is treated as AL (always). It reports whether an ACK was
// written; MSA-1 is validated against the inbound MSH-12 version.
func (r *SendResponse) SendApplicationAck(outcome AckOutcome) (bool, error) {
	field := r.message.Get("MSH.16").String()
	code, send := NegotiateAck(field, AckApplication, outcome)
	if !send {
		return false, nil
	}
	if err := r.SendResponse(code); err != nil {
		return false, err
	}
	return true, nil
}
