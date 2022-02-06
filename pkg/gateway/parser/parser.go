package parser

import (
	"dns-proxy/pkg/domain/proxy"
	"errors"
	"fmt"

	"golang.org/x/net/dns/dnsmessage"
)

type dnsParser struct{}

func NewDNSParser() proxy.DNSParser {
	return &dnsParser{}
}

func (mp *dnsParser) DNSToMsg(dnsm *dnsmessage.Message, protocol string) ([]byte, error) {
	message, err := dnsm.Pack()
	if err != nil {
		return nil, err
	}
	if protocol == proxy.SocketTCP {
		//	https://www.ietf.org/rfc/rfc1035.txt
		//	4.2.2. TCP usage
		//	Messages sent over TCP connections use server port 53 (decimal).  The
		//	message is prefixed with a two byte length field which gives the message
		//	length, excluding the two byte length field.  This length field allows
		//	the low-level processing to assemble a complete message before beginning
		//	to parse it.
		var prefixBytesTCP []byte
		prefixBytesTCP = make([]byte, 2)
		prefixBytesTCP[0] = 0
		prefixBytesTCP[1] = byte(len(message))
		message = append(prefixBytesTCP, message...)
		return message, nil
	} else if protocol == proxy.SocketUDP {
		return message, nil
	} else {
		return nil, errors.New("invalid msg format")
	}
}

func (mp *dnsParser) UDPMsgToDNS(m []byte) (*dnsmessage.Message, error) {
	var dnsm dnsmessage.Message
	err := dnsm.Unpack(m[:])
	if err != nil {
		return nil, fmt.Errorf("unable to unpack udp message: %s", err)
	}
	return &dnsm, nil
}

func (mp *dnsParser) TCPMsgToDNS(m []byte) (*dnsmessage.Message, error) {
	var dnsm dnsmessage.Message
	// Unpack the mesage starting from the second position.
	err := dnsm.Unpack(m[2:])
	if err != nil {
		return nil, fmt.Errorf("unable to unpack tcp message: %s", err)
	}
	return &dnsm, nil
}
