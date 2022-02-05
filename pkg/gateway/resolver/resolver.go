package resolver

import (
	"crypto/tls"
	"crypto/x509"
	"dns-proxy/pkg/domain/proxy"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
)

const cflRootCert = `-----BEGIN CERTIFICATE-----
MIIEQzCCAyugAwIBAgIQCidf5wTW7ssj1c1bSxpOBDANBgkqhkiG9w0BAQwFADBh
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMSAwHgYDVQQDExdEaWdpQ2VydCBHbG9iYWwgUm9vdCBD
QTAeFw0yMDA5MjMwMDAwMDBaFw0zMDA5MjIyMzU5NTlaMFYxCzAJBgNVBAYTAlVT
MRUwEwYDVQQKEwxEaWdpQ2VydCBJbmMxMDAuBgNVBAMTJ0RpZ2lDZXJ0IFRMUyBI
eWJyaWQgRUNDIFNIQTM4NCAyMDIwIENBMTB2MBAGByqGSM49AgEGBSuBBAAiA2IA
BMEbxppbmNmkKaDp1AS12+umsmxVwP/tmMZJLwYnUcu/cMEFesOxnYeJuq20ExfJ
qLSDyLiQ0cx0NTY8g3KwtdD3ImnI8YDEe0CPz2iHJlw5ifFNkU3aiYvkA8ND5b8v
c6OCAa4wggGqMB0GA1UdDgQWBBQKvAgpF4ylOW16Ds4zxy6z7fvDejAfBgNVHSME
GDAWgBQD3lA1VtFMu2bwo+IbG8OXsj3RVTAOBgNVHQ8BAf8EBAMCAYYwHQYDVR0l
BBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMBIGA1UdEwEB/wQIMAYBAf8CAQAwdgYI
KwYBBQUHAQEEajBoMCQGCCsGAQUFBzABhhhodHRwOi8vb2NzcC5kaWdpY2VydC5j
b20wQAYIKwYBBQUHMAKGNGh0dHA6Ly9jYWNlcnRzLmRpZ2ljZXJ0LmNvbS9EaWdp
Q2VydEdsb2JhbFJvb3RDQS5jcnQwewYDVR0fBHQwcjA3oDWgM4YxaHR0cDovL2Ny
bDMuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0R2xvYmFsUm9vdENBLmNybDA3oDWgM4Yx
aHR0cDovL2NybDQuZGlnaWNlcnQuY29tL0RpZ2lDZXJ0R2xvYmFsUm9vdENBLmNy
bDAwBgNVHSAEKTAnMAcGBWeBDAEBMAgGBmeBDAECATAIBgZngQwBAgIwCAYGZ4EM
AQIDMA0GCSqGSIb3DQEBDAUAA4IBAQDeOpcbhb17jApY4+PwCwYAeq9EYyp/3YFt
ERim+vc4YLGwOWK9uHsu8AjJkltz32WQt960V6zALxyZZ02LXvIBoa33llPN1d9R
JzcGRvJvPDGJLEoWKRGC5+23QhST4Nlg+j8cZMsywzEXJNmvPlVv/w+AbxsBCMqk
BGPI2lNM8hkmxPad31z6n58SXqJdH/bYF462YvgdgbYKOytobPAyTgr3mYI5sUje
CzqJx1+NLyc8nAK8Ib2HxnC+IrrWzfRLvVNve8KaN9EtBH7TuMwNW4SpDCmGr6fY
1h3tDjHhkTb9PA36zoaJzu0cIw265vZt6hCmYWJC+/j+fgZwcPwL
-----END CERTIFICATE-----
`

type resolver struct {
	ip          string
	port        int
	rootCert    string
	readTimeOut uint
}

func NewResolver(ip string, port int, rto uint) proxy.Resolver {
	return &resolver{ip, port, cflRootCert, rto}
}

func (r *resolver) GetTLSConnection() (*tls.Conn, error) {
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(r.rootCert)) {
		log.Println("Fail to parse rootCert")
		return nil, errors.New("Fail to parse rootCert")
	}
	dnsConn, err := tls.Dial(proxy.TCP, r.ip+":"+strconv.Itoa(r.port), &tls.Config{
		RootCAs: roots,
	})
	if err != nil {
		return nil, err
	}
	err = dnsConn.SetReadDeadline(time.Now().Add(time.Duration(r.readTimeOut) * time.Millisecond))
	if err != nil {
		return nil, err
	}
	return dnsConn, nil
}

func (r *resolver) Resolve(um proxy.MessageFromRequest) (proxy.MessageFromProvider, error) {
	conn, err := r.GetTLSConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, e := conn.Write(um)
	if e != nil {
		fmt.Printf("%v", e)
	}
	var reply [2045]byte
	n, err := conn.Read(reply[:])
	if err != nil {
		fmt.Printf("Could read response from DNSProvider: %v \n", err)
		return nil, errors.New("could not read response from DNSProvider")
	}
	return reply[:n], nil
}
