package parser

import (
	"dns-proxy/pkg/domain/proxy"
	"errors"
	"log"

	"golang.org/x/net/dns/dnsmessage"
)

type dnsParser struct{}

func NewDNSParser() proxy.DNSParser {
	return &dnsParser{}
}

func (mp *dnsParser) DNSToMsg(dnsm *dnsmessage.Message, protocol string) (proxy.Message, error) {
	m, err := dnsm.Pack()
	if err != nil {
		return nil, err
	}
	if protocol == "tcp" {
		var tcpBytes []byte
		tcpBytes = make([]byte, 2)
		tcpBytes[0] = 0
		tcpBytes[1] = byte(len(m))
		m = append(tcpBytes, m...)
		return m, nil
	} else if protocol == "udp" {
		return m, nil
	} else {
		return nil, errors.New("invalid msg format")
	}
}

func (mp *dnsParser) UDPMsgToDNS(m proxy.Message) (*dnsmessage.Message, error) {
	var dnsm dnsmessage.Message
	err := dnsm.Unpack(m[:])
	if err != nil {
		log.Printf("Unable to parse UDP Message: %v \n", err)
		return nil, errors.New("unable to unpack request, invalid message.")
	}
	return &dnsm, nil
}

func (mp *dnsParser) TCPMsgToDNS(m proxy.Message) (*dnsmessage.Message, error) {
	var dnsm dnsmessage.Message
	err := dnsm.Unpack(m[2:])
	if err != nil {
		log.Printf("Unable to parse TCP Message: %v \n", err)
		return nil, errors.New("unable to unpack request, invalid message.")
	}
	return &dnsm, nil
}
