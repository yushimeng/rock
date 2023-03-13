package transport

import (
	"net"
	"strconv"

	"github.com/yushimeng/rock/sip"
)

var (
	SIPDebug bool
)

const (
	// Transport for different sip messages. GO uses lowercase, but for message parsing, we should
	// use this constants for setting message Transport
	TransportUDP = "UDP"
	TransportTCP = "TCP"
	TransportTLS = "TLS"
	TransportWS  = "WS"
)

// Protocol implements network specific features.
type Transport interface {
	Addr() string
	Network() string
	ListenAndServe(handler sip.MessageHandler) error
	GetConnection(addr string) (Connection, error)
	CreateConnection(addr string) (Connection, error)
	String() string
	Close() error
}

func ParseAddr(addr string) (host string, port int, err error) {
	host, pstr, err := net.SplitHostPort(addr)
	if err != nil {
		return host, port, err
	}

	port, err = strconv.Atoi(pstr)
	return host, port, err
}
