package proxy

import (
	"crypto/tls"
	"errors"
	"fmt"

	"golang.org/x/net/dns/dnsmessage"
)

type Message []byte
type MessageFromRequest = Message
type MessageFromProvider = Message

const (
	TCP = "tcp"
	UDP = "udp"
)

type Resolver interface {
	Resolve(um MessageFromRequest) (MessageFromProvider, error)
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
	UDPMsgToDNS(m Message) (*dnsmessage.Message, error)
	TCPMsgToDNS(m Message) (*dnsmessage.Message, error)
	DNSToMsg(dnsm *dnsmessage.Message, protocol string) (Message, error)
}

// Logger interface is used to inject different implementations of loggers.
type Logger interface {
	Info(format string, i ...interface{})
	Err(format string, i ...interface{})
	Debug(format string, i ...interface{})
}

// Service interface is used to define the two kind of requests this proxy can solve UDP or TCP.
type Service interface {
	SolveTCP(MessageFromRequest) (MessageFromProvider, error)
	SolveUDP(MessageFromRequest) (MessageFromProvider, error)
}

type service struct {
	resolver Resolver
	parser   DNSParser
	cache    Cache
	logger   Logger
}

func NewDNSProxy(r Resolver, mp DNSParser, c Cache, l Logger) Service {
	return &service{r, mp, c, l}
}

func (s *service) SolveTCP(messageFromRequest MessageFromRequest) (MessageFromProvider, error) {
	return s.solve(messageFromRequest, TCP)
}

func (s *service) SolveUDP(messageFromRequest MessageFromRequest) (MessageFromProvider, error) {
	return s.solveUDP(messageFromRequest)
}

func (s *service) solveUDP(messageFromRequest MessageFromRequest) (MessageFromProvider, error) {
	// At this point 'message' is DNS format but UDP
	message, err := s.parser.UDPMsgToDNS(messageFromRequest)
	// messageFromRequest came as raw UDP and now is raw TCP.
	messageFromRequest, err = s.parser.DNSToMsg(message, TCP)
	if err != nil {
		s.logger.Err("error parsing UnsolvedMsg: %v \n", err)
		return nil, err
	}
	for _, q := range message.Questions {
		s.logger.Info("DNS %s: %s ", UDP, q.Name.String())
	}
	// Check if record exists in the cache and return it.
	// If exists, it was saved as 'dnsmessage' UDP or TCP. Check using 'dnsmessage'.
	cachedMessage, err := s.cache.Get(*message)
	if err != nil {
		s.logger.Err("Cache error: %v", err)
	}
	if cachedMessage != nil {
		s.logger.Debug("Message found in cache")
		cachedMessage.Header.ID = message.Header.ID
		// Message from cache returns as 'dnsmessage'. Parse to 'raw' to return.
		cachedMessageRaw, err := s.parser.DNSToMsg(cachedMessage, UDP)
		if err != nil {
			return nil, errors.New("wrong value stored in cache")
		}
		return cachedMessageRaw, nil
	}
	// If it was not cached resolve it.
	// Resolve the DNS against the DNS provider.
	// The resolver returns a TCP Raw response that can be returned by this method.
	solvedMessageRaw, err := s.resolver.Resolve(messageFromRequest)
	if err != nil {
		s.logger.Err("resolution Error: %v \n", err)
		return nil, err
	}
	// To update the cache a 'dnsmessage' type is needed.
	// Cast the Raw TCP response from the Resolver.
	DNSTCPResponse, err := s.parser.TCPMsgToDNS(solvedMessageRaw)
	if err != nil {
		return nil, err
	}
	err = s.cache.Store(*DNSTCPResponse)
	if err != nil {
		s.logger.Err("error updating cache: %v \n", err)
	}
	// The raw message returned should be in the format that it was requested at the origin.
	solvedMessageRaw, _ = s.parser.DNSToMsg(DNSTCPResponse, UDP)
	return solvedMessageRaw, nil
}

func (s *service) solve(messageFromRequest MessageFromRequest, protocol string) (MessageFromProvider, error) {
	var err error
	var message *dnsmessage.Message
	if protocol == UDP {
		message, err = s.parser.UDPMsgToDNS(messageFromRequest)
		messageFromRequest, err = s.parser.DNSToMsg(message, TCP) // Parse UDP message in order to send it to DoT Resolver (TCP)
	} else if protocol == TCP {
		message, err = s.parser.TCPMsgToDNS(messageFromRequest)
	} else {
		return nil, fmt.Errorf("invalid msg format: %s \n ", protocol)
	}
	if err != nil {
		s.logger.Err("error parsing UnsolvedMsg: %v \n", err)
		return nil, err
	}
	for _, q := range message.Questions {
		s.logger.Info("DNS %s: %s ", protocol, q.Name.String())
	}

	// Check if record exists in the cache and return it.
	//	solvedMessage, err := s.getFromCache(message, protocol)
	cachedMessage, err := s.cache.Get(*message)
	if err != nil {
		s.logger.Err("Cache error: %v", err)
	}
	if cachedMessage != nil {
		s.logger.Debug("Message found in cache")
		cachedMessage.Header.ID = message.Header.ID
		solvedMessage, err := s.parser.DNSToMsg(cachedMessage, protocol)
		if err != nil {
			return nil, errors.New("wrong value stored in cache")
		}
		return solvedMessage, nil
	}
	// Resolve the DNS against the DNS provider.
	solvedMessage, err := s.resolver.Resolve(messageFromRequest)
	if err != nil {
		s.logger.Err("resolution Error: %v \n", err)
		return nil, err
	}
	// Update the cache and don't wait until it's updated to return the value.
	tcpDNSResponse, err := s.parser.TCPMsgToDNS(solvedMessage) // Resolver response is always TCP
	err = s.cache.Store(*tcpDNSResponse)
	if err != nil {
		s.logger.Err("resolution Error: %v \n", err)
	}
	solvedMessage, _ = s.parser.DNSToMsg(tcpDNSResponse, protocol)
	return solvedMessage, nil
}
