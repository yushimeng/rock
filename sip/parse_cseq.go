package sip

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse a string representation of a CSeq header, returning a slice of at most one CSeq.
func parseCSeq(headerName string, headerText string) (
	headers Header, err error) {
	var cseq CSeqHeader
	ind := strings.IndexAny(headerText, abnfWs)
	if ind < 1 || len(headerText)-ind < 2 {
		err = fmt.Errorf(
			"CSeq field should have precisely one whitespace section: '%s'",
			headerText,
		)
		return
	}

	var seqno uint64
	seqno, err = strconv.ParseUint(headerText[:ind], 10, 32)
	if err != nil {
		return
	}

	if seqno > maxCseq {
		err = fmt.Errorf("invalid CSeq %d: exceeds maximum permitted value "+
			"2**31 - 1", seqno)
		return
	}

	cseq.SeqNo = uint32(seqno)
	cseq.MethodName = RequestMethod(headerText[ind+1:])
	return &cseq, nil
}
