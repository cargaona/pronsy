package proxy

import (
	"crypto/tls"

	"golang.org/x/net/dns/dnsmessage"
)

//type Message []byte
//type MessageFromRequest = Message
//type MessageFromProvider = Message

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
	AutoPurge()
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
	cache    Cache
	logger   Logger
}

func NewDNSProxy(r Resolver, p DNSParser, c Cache, l Logger) Service {
	return &service{
		resolver: r,
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
		s.logger.Info("DNS %s: %s ", protocol, q.Name.String())
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
	dnsresponse, _ := s.parser.TCPMsgToDNS(response)
	response, _ = s.parser.DNSToMsg(dnsresponse, SocketUDP)
	return response, nil
}

//func (s *service) solve(messageFromRequest MessageFromRequest, protocol string) (MessageFromProvider, error) {
//	var err error
//	var message *dnsmessage.Message
//	if protocol == SocketUDP {
//		// At this point 'message' is DNS format but UDP
//		message, err = s.parser.UDPMsgToDNS(messageFromRequest)
//		// messageFromRequest came as raw UDP and now is raw TCP.
//		messageFromRequest, err = s.parser.DNSToMsg(message, SocketTCP)
//	}
//	if protocol == SocketTCP {
//		message, err = s.parser.TCPMsgToDNS(messageFromRequest)
//	}
//	if err != nil {
//		s.logger.Err("error parsing UnsolvedMsg: %v \n", err)
//		return nil, err
//	}
//	for _, q := range message.Questions {
//		s.logger.Info("DNS %s: %s ", SocketUDP, q.Name.String())
//	}
//	// Check if record exists in the cache and return it.
//	// If exists, it was saved as 'dnsmessage' UDP or TCP. Check using 'dnsmessage'.
//	cachedMessage, err := s.cache.Get(*message)
//	if err != nil {
//		s.logger.Err("Cache error: %v", err)
//	}
//	if cachedMessage != nil {
//		s.logger.Debug("Message found in cache")
//		cachedMessage.Header.ID = message.Header.ID
//		// Message from cache returns as 'dnsmessage'. Parse to 'raw' to return.
//		cachedMessageRaw, err := s.parser.DNSToMsg(cachedMessage, protocol)
//		if err != nil {
//			return nil, errors.New("wrong value stored in cache")
//		}
//		return cachedMessageRaw, nil
//	}
//	// If it was not cached resolve it.
//	// Resolve the DNS against the DNS provider.
//	// The resolver returns a TCP Raw response that can be returned by this method.
//	solvedMessageRaw, err := s.resolver.Resolve(messageFromRequest)
//	if err != nil {
//		s.logger.Err("resolution Error: %v \n", err)
//		return nil, err
//	}
//	// To update the cache a 'dnsmessage' type is needed.
//	// Cast the Raw TCP response from the Resolver.
//	DNSTCPResponse, err := s.parser.TCPMsgToDNS(solvedMessageRaw)
//	if err != nil {
//		return nil, err
//	}
//	err = s.cache.Store(*DNSTCPResponse)
//	if err != nil {
//		s.logger.Err("error updating cache: %v \n", err)
//	}
//	// If the protocol is TCP the message is ready to be sent.
//	if protocol == SocketTCP {
//		return solvedMessageRaw, nil
//	}
//	// If it was not TCP it was UDP. TCP to UDP cast is needed.
//	solvedMessageRaw, _ = s.parser.DNSToMsg(DNSTCPResponse, SocketUDP)
//	return solvedMessageRaw, nil
//}
//
