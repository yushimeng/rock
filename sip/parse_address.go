package sip

import (
	"fmt"
	"strings"
)

// ParseAddressValue parses an address - such as from a From, To, or
// Contact header. It returns:
// See RFC 3261 section 20.10 for details on parsing an address.
// Note that this method will not accept a comma-separated list of addresses;
// addresses in that form should be handled by ParseAddressValues.
func ParseAddressValue(addressText string, uri *Uri, headerParams HeaderParams) (displayName string, err error) {
	// headerParams = NewParams()
	var semicolon, equal, startQuote, endQuote int = -1, -1, -1, -1
	var name string
	var uriStart, uriEnd int = 0, -1
	var inBrackets bool
	for i, c := range addressText {
		switch c {
		case '"':
			if startQuote < 0 {
				startQuote = i
			}
			endQuote = i
		case '<':
			if uriStart > 0 {
				// This must be additional options parsing
				continue
			}

			if endQuote > 0 {
				displayName = addressText[startQuote+1 : endQuote]
				startQuote, endQuote = -1, -1
			} else {
				displayName = strings.TrimSpace(addressText[:i])
			}
			uriStart = i + 1
			inBrackets = true
		case '>':
			// uri can be without <> in that case there all after ; are header params
			uriEnd = i
			equal = -1
			inBrackets = false
		case ';':
			semicolon = i
			// uri can be without <> in that case there all after ; are header params
			if inBrackets {
				continue
			}

			if uriEnd < 0 {
				uriEnd = i
				continue
			}

			if equal > 0 {
				val := addressText[equal+1 : i]
				headerParams.Add(name, val)
				name, val = "", ""
				equal = 0
			}

		case '=':
			name = addressText[semicolon+1 : i]
			equal = i
		case '*':
			if startQuote > 0 || uriStart > 0 {
				continue
			}
			uri = &Uri{
				Wildcard: true,
			}
			return
		}
	}

	if uriEnd < 0 {
		uriEnd = len(addressText)
	}
	err = ParseUri(addressText[uriStart:uriEnd], uri)
	if err != nil {
		return
	}

	if equal > 0 {
		val := addressText[equal+1:]
		headerParams.Add(name, val)
		name, val = "", ""
	}
	// params := strings.Split(addressText, ";")
	// if len(params) > 1 {
	// 	for _, section := range params[1:] {
	// 		arr := strings.Split(section, "=")
	// 		headerParams.Add(arr[0], arr[1])
	// 	}
	// }

	return
}

func parseToAddressHeader(headerName string, headerText string) (header Header, err error) {

	h := &ToHeader{
		Address: Uri{},
		Params:  NewParams(),
	}
	h.DisplayName, err = ParseAddressValue(headerText, &h.Address, h.Params)
	// h.DisplayName, h.Address, h.Params, err = ParseAddressValue(headerText)

	if h.Address.Wildcard {
		// The Wildcard '*' URI is only permitted in Contact headers.
		err = fmt.Errorf(
			"wildcard uri not permitted in to: header: %s",
			headerText,
		)
		return
	}
	return h, err
}

func parseFromAddressHeader(headerName string, headerText string) (header Header, err error) {

	h := FromHeader{
		Address: Uri{},
		Params:  NewParams(),
	}
	h.DisplayName, err = ParseAddressValue(headerText, &h.Address, h.Params)
	// h.DisplayName, h.Address, h.Params, err = ParseAddressValue(headerText)
	if err != nil {
		return
	}

	if err != nil {
		return
	}

	if h.Address.Wildcard {
		// The Wildcard '*' URI is only permitted in Contact headers.
		err = fmt.Errorf(
			"wildcard uri not permitted in to: header: %s",
			headerText,
		)
		return
	}
	return &h, nil
}

func parseContactAddressHeader(headerName string, headerText string) (header Header, err error) {
	prevIdx := 0
	inBrackets := false
	inQuotes := false

	// Append a comma to simplify the parsing code; we split address sections
	// on commas, so use a comma to signify the end of the final address section.
	addresses := headerText + ","

	head := ContactHeader{
		Params: NewParams(),
	}
	h := &head
	for idx, char := range addresses {
		if char == '<' && !inQuotes {
			inBrackets = true
		} else if char == '>' && !inQuotes {
			inBrackets = false
		} else if char == '"' {
			inQuotes = !inQuotes
		} else if !inQuotes && !inBrackets && char == ',' {
			// if char == ',' {
			if h == nil {
				h = &ContactHeader{
					Params: NewParams(),
				}
			}
			h.DisplayName, err = ParseAddressValue(addresses[prevIdx:idx], &h.Address, h.Params)
			if err != nil {
				return
			}
			prevIdx = idx + 1
			h = h.Next
		}
	}

	return &head, nil
}

func parseRouteHeader(headerName string, headerText string) (header Header, err error) {
	prevIdx := 0
	inBrackets := false
	inQuotes := false

	// Append a comma to simplify the parsing code; we split address sections
	// on commas, so use a comma to signify the end of the final address section.
	addresses := headerText + ","

	head := RouteHeader{}
	h := &head
	for idx, char := range addresses {
		if char == '<' && !inQuotes {
			inBrackets = true
		} else if char == '>' && !inQuotes {
			inBrackets = false
		} else if char == '"' {
			inQuotes = !inQuotes
		} else if !inQuotes && !inBrackets && char == ',' {
			// if char == ',' {
			if h == nil {
				h = &RouteHeader{}
			}
			_, err = ParseAddressValue(addresses[prevIdx:idx], &h.Address, nil)
			if err != nil {
				return
			}
			prevIdx = idx + 1
			h = h.Next
		}
	}

	return &head, nil
}

func parseRecordRouteHeader(headerName string, headerText string) (header Header, err error) {
	prevIdx := 0
	inBrackets := false
	inQuotes := false

	// Append a comma to simplify the parsing code; we split address sections
	// on commas, so use a comma to signify the end of the final address section.
	addresses := headerText + ","

	head := RecordRouteHeader{}
	h := &head
	for idx, char := range addresses {
		if char == '<' && !inQuotes {
			inBrackets = true
		} else if char == '>' && !inQuotes {
			inBrackets = false
		} else if char == '"' {
			inQuotes = !inQuotes
		} else if !inQuotes && !inBrackets && char == ',' {
			// if char == ',' {
			if h == nil {
				h = &RecordRouteHeader{}
			}
			_, err = ParseAddressValue(addresses[prevIdx:idx], &h.Address, nil)
			if err != nil {
				return
			}
			prevIdx = idx + 1
			h = h.Next
		}
	}

	return &head, nil
}
