package proxy

import (
	"crypto/tls"
	"dns-proxy/pkg/domain/denylist"

	"golang.org/x/net/dns/dnsmessage"
)

// SocketTCP and SocketUDP are constants that represent the 'udp' and 'tcp' strings used multiple times.
const (
	SocketTCP = "tcp"
	SocketUDP = "udp"
)

type Resolver interface {
	Resolve(um []byte) ([]byte, error)
	GetTLSConnection() (*tls.Conn, error)
}

// Cache is the interace used to avoid requesting the DNS provider all the time.
type Cache interface {
	Get(dnsm dnsmessage.Message) (*dnsmessage.Message, error)
	Store(dnsm dnsmessage.Message) error
	Flush()
}

// DNSParser is the interface used to parse the messages between the domain entity type and the *dnsmessage.Message type.
type DNSParser interface {
	UDPMsgToDNS(m []byte) (*dnsmessage.Message, error)
	TCPMsgToDNS(m []byte) (*dnsmessage.Message, error)
	DNSToMsg(dnsm *dnsmessage.Message, protocol string) ([]byte, error)
}

// Logger interface is used to inject different implementations of loggers.
type Logger interface {
	Info(format string, i ...interface{})
	Err(format string, i ...interface{})
	Debug(format string, i ...interface{})
}

// Service interface is used to define the two kind of requests this proxy can solve UDP or TCP.
type Service interface {
	SolveTCP([]byte) ([]byte, error)
	SolveUDP([]byte) ([]byte, error)
}

type service struct {
	resolver Resolver
	parser   DNSParser
	denier   denylist.Service
	cache    Cache
	logger   Logger
}

func NewDNSProxy(r Resolver, d denylist.Service, p DNSParser, c Cache, l Logger) Service {
	return &service{
		resolver: r,
		denier:   d,
		parser:   p,
		cache:    c,
		logger:   l,
	}
}

func (s *service) SolveTCP(request []byte) ([]byte, error) {
	return s.solve(request, SocketTCP)
}

func (s *service) SolveUDP(request []byte) ([]byte, error) {
	return s.solve(request, SocketUDP)
}

func (s *service) solve(request []byte, protocol string) ([]byte, error) {
	var err error
	var message *dnsmessage.Message
	if protocol == SocketUDP {
		// At this point 'message' is DNS format but UDP
		message, err = s.parser.UDPMsgToDNS(request)
		// Convert to TCP to request against the DNS provider
		request, err = s.parser.DNSToMsg(message, SocketTCP)
	}
	if protocol == SocketTCP {
		message, err = s.parser.TCPMsgToDNS(request)
	}
	if err != nil {
		s.logger.Err("error parsing UnsolvedMsg: %v \n", err)
		return nil, err
	}
	for _, q := range message.Questions {
		// WIP look for the domain in the denylist before resolve it.
		// denied, err := s.denier.GetDeniedDomain(q.Name.String())
		// if err != nil {
		// 	s.logger.Err("error looking for the domain in the denylist: %v", err)
		// }
		// if denied.Domain != "" {
		// 	return nil, fmt.Errorf("domain %s found in denylist, can't resolve", q.Name.String())
		// TODO: write 'empty' response as the DNS couldn't find the domain instead of sending an error'
		// }
		s.logger.Info("Resolving DNS %s: %s ", protocol, q.Name.String())
	}
	// Resolve the DNS against the DNS provider.
	// The resolver returns a TCP Raw response that can be returned by this method.
	response, err := s.resolver.Resolve(request)
	if err != nil {
		s.logger.Err("resolution Error: %v \n", err)
		return nil, err
	}
	// If the protocol is TCP the message is ready to be sent.
	if protocol == SocketTCP {
		return response, nil
	}
	// If it was not TCP a TCP to UDP cast is needed.
	dnsResponse, _ := s.parser.TCPMsgToDNS(response)
	response, _ = s.parser.DNSToMsg(dnsResponse, SocketUDP)
	return response, nil
}
