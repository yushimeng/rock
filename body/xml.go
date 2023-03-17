package body

import (
	"encoding/xml"
	"strings"
)

type CatalogQueryBody struct {
	XMLName  xml.Name `xml:"Query"`
	CmdType  string   `xml:"CmdType"`
	Sn       string   `xml:"SN"`
	DeviceId string   `xml:"DeviceID"`
}

func (x *CatalogQueryBody) Builder() ([]byte, error) {

	sb := &strings.Builder{}

	sb.WriteString(string(xml.Header))
	output, err := xml.Marshal(x)
	if err != nil {
		return nil, err
	}

	sb.WriteString(string(output))
	return []byte(sb.String()), nil
}
