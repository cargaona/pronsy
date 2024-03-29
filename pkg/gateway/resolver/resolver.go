package resolver

import (
	"crypto/tls"
	"crypto/x509"
	"dns-proxy/pkg/domain/proxy"
	"fmt"
	"net"
	"strconv"
	"time"
)

type resolver struct {
	dnsIP       string
	port        int
	readTimeOut uint
}

func New(ip string, port int, readTimeOut uint) proxy.Resolver {
	return &resolver{
		dnsIP:       ip,
		port:        port,
		readTimeOut: readTimeOut,
	}
}

func (r *resolver) GetTLSConnection() (*tls.Conn, error) {
	certs, err := r.getRootsCA()
	if err != nil {
		return nil, err
	}
	roots := x509.NewCertPool()
	for _, cert := range certs {
		roots.AddCert(cert)
	}
	conn, err := tls.Dial(
		proxy.SocketTCP, net.JoinHostPort(r.dnsIP, strconv.Itoa(r.port)),
		&tls.Config{
			RootCAs: roots,
		})
	if err != nil {
		return nil, err
	}
	err = conn.SetReadDeadline(time.Now().Add(time.Duration(r.readTimeOut) * time.Millisecond))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (r *resolver) Resolve(request []byte) ([]byte, error) {
	conn, err := r.GetTLSConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, e := conn.Write(request)
	if e != nil {
		return nil, fmt.Errorf("could not send request to DNS Provider %s", r.dnsIP)
	}
	var reply [2045]byte
	n, err := conn.Read(reply[:])
	if err != nil {
		return nil, fmt.Errorf("could not read response from DNS Provider %s", r.dnsIP)
	}
	return reply[:n], nil
}

func (r *resolver) getRootsCA() ([]*x509.Certificate, error) {
	// TODO: There's no need to go get the certs from the provider on every request.
	// Create a cache for the certs and update them every now and then to give a faster response.
	conn, err := tls.Dial(
		proxy.SocketTCP,
		net.JoinHostPort(r.dnsIP, strconv.Itoa(r.port)),
		&tls.Config{
			InsecureSkipVerify: true,
		})
	if err != nil {
		return nil, err
	}
	return conn.ConnectionState().PeerCertificates, nil
}
