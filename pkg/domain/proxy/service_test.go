package proxy

import (
	"crypto/tls"
	"errors"
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

type ResolverMock struct {
	SolveMethod            func(um UnsolvedMsg) (SolvedMsg, error)
	GetTLSConnectionMethod func() (*tls.Conn, error)
}

func (rm *ResolverMock) GetTLSConnection() (*tls.Conn, error) {
	if rm.GetTLSConnectionMethod != nil {
		return rm.GetTLSConnectionMethod()
	}
	return nil, errors.New("Couldn't connect to the resolver")
}

func (rm *ResolverMock) Solve(um UnsolvedMsg) (SolvedMsg, error) {
	if rm.SolveMethod != nil {
		return rm.SolveMethod(um)
	}
	return nil, errors.New("Couldn't solve the message")

}

type CacheMock struct {
	GetMethod   func(*dnsmessage.Message) (SolvedMsg, error)
	StoreMethod func(*dnsmessage.Message, SolvedMsg) error
}

func (cm *CacheMock) Get(dnsm *dnsmessage.Message) (SolvedMsg, error) {
	if cm.GetMethod != nil {
		return cm.GetMethod(dnsm)
	}
	return nil, errors.New("Message not cached")
}

func (cm *CacheMock) Store(dnsm *dnsmessage.Message, m SolvedMsg) error {
	if cm.StoreMethod != nil {
		return cm.StoreMethod(dnsm, m)
	}
	return errors.New("Could not store the message")
}

type ParserMock struct {
	ParseTCPMsgMethod func(Msg) (*dnsmessage.Message, error)
	ParseUDPMsgMethod func(Msg) (*dnsmessage.Message, UnsolvedMsg, error)
}

func (mp *ParserMock) ParseUPDMsg(m Msg) (*dnsmessage.Message, UnsolvedMsg, error) {
	if mp.ParseUDPMsgMethod != nil {
		return mp.ParseUDPMsgMethod(m)
	}
	return nil, nil, errors.New("Could not parse UDP Message")
}

func (mp *ParserMock) ParseTCPMsg(m Msg) (*dnsmessage.Message, error) {
	if mp.ParseTCPMsgMethod != nil {
		return mp.ParseTCPMsgMethod(m)
	}
	return nil, errors.New("Could not store the message")
}

func TestDirectFailuer(t *testing.T) {
	resolverMock := &ResolverMock{}
	parserMock := &ParserMock{}
	cacheMock := &CacheMock{}

	proxyTest := NewDNSProxy(resolverMock, parserMock, cacheMock)
	sm, err := proxyTest.Solve([]byte("Invalid msg format"), "tcp")
	if err == nil || sm != nil {
		t.Error("proxy.Solved() Succeded when it should have failed.")
	}
}
