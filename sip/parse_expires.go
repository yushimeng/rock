package sip

// // Parse a string representation of a Call-ID header, returning a slice of at most one CallID.
// func parseExpires(headerName string, headerText string) (header Header, err error) {
// 	headerText = strings.TrimSpace(headerText)

// 	// if strings.ContainsAny(string(callId), abnfWs) {
// 	// 	err = fmt.Errorf("unexpected whitespace in CallID header body '%s'", headerText)
// 	// 	return
// 	// }
// 	// if strings.Contains(string(callId), ";") {
// 	// 	err = fmt.Errorf("unexpected semicolon in CallID header body '%s'", headerText)
// 	// 	return
// 	// }
// 	if len(headerText) == 0 {
// 		err = fmt.Errorf("empty Call-ID body")
// 		return
// 	}

// 	var callId = CallIDHeader(headerText)

// 	return &callId, nil
// }

/*
// Header is a single SIP header.
type Header interface {
	// Name returns header name.
	Name() string
	Value() string
	// Clone() Header
	String() string
	// StringWrite is better way to reuse single buffer
	StringWrite(w io.StringWriter)

	// Next() Header
	headerClone() Header
}
*/
