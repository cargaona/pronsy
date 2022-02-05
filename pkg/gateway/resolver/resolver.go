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

const newCloudFlare = `-----BEGIN CERTIFICATE-----
MIIF+TCCBYCgAwIBAgIQD3WjbTLBawPHyl9fcUoDcDAKBggqhkjOPQQDAzBWMQsw
CQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMTAwLgYDVQQDEydEaWdp
Q2VydCBUTFMgSHlicmlkIEVDQyBTSEEzODQgMjAyMCBDQTEwHhcNMjExMDI1MDAw
MDAwWhcNMjIxMDI1MjM1OTU5WjByMQswCQYDVQQGEwJVUzETMBEGA1UECBMKQ2Fs
aWZvcm5pYTEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEZMBcGA1UEChMQQ2xvdWRm
bGFyZSwgSW5jLjEbMBkGA1UEAxMSY2xvdWRmbGFyZS1kbnMuY29tMFkwEwYHKoZI
zj0CAQYIKoZIzj0DAQcDQgAE+ylE8pg/2L2CVtMsvY4JnzErmCaeIpaNe0v82sV7
eymqjjVsnApIBWyJc+0gDs1GIfDsTbOl6a8bOJnl9NrxhKOCBBIwggQOMB8GA1Ud
IwQYMBaAFAq8CCkXjKU5bXoOzjPHLrPt+8N6MB0GA1UdDgQWBBQZRRsjGPh02iIU
y0Zr4hOzYBWCQDCBpgYDVR0RBIGeMIGbghJjbG91ZGZsYXJlLWRucy5jb22CFCou
Y2xvdWRmbGFyZS1kbnMuY29tgg9vbmUub25lLm9uZS5vbmWHBAEBAQGHBAEAAAGH
BKKfJAGHBKKfLgGHECYGRwBHAAAAAAAAAAAAERGHECYGRwBHAAAAAAAAAAAAEAGH
ECYGRwBHAAAAAAAAAAAAAGSHECYGRwBHAAAAAAAAAAAAZAAwDgYDVR0PAQH/BAQD
AgeAMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjCBmwYDVR0fBIGTMIGQ
MEagRKBChkBodHRwOi8vY3JsMy5kaWdpY2VydC5jb20vRGlnaUNlcnRUTFNIeWJy
aWRFQ0NTSEEzODQyMDIwQ0ExLTEuY3JsMEagRKBChkBodHRwOi8vY3JsNC5kaWdp
Y2VydC5jb20vRGlnaUNlcnRUTFNIeWJyaWRFQ0NTSEEzODQyMDIwQ0ExLTEuY3Js
MD4GA1UdIAQ3MDUwMwYGZ4EMAQICMCkwJwYIKwYBBQUHAgEWG2h0dHA6Ly93d3cu
ZGlnaWNlcnQuY29tL0NQUzCBhQYIKwYBBQUHAQEEeTB3MCQGCCsGAQUFBzABhhho
dHRwOi8vb2NzcC5kaWdpY2VydC5jb20wTwYIKwYBBQUHMAKGQ2h0dHA6Ly9jYWNl
cnRzLmRpZ2ljZXJ0LmNvbS9EaWdpQ2VydFRMU0h5YnJpZEVDQ1NIQTM4NDIwMjBD
QTEtMS5jcnQwDAYDVR0TAQH/BAIwADCCAX4GCisGAQQB1nkCBAIEggFuBIIBagFo
AHcAKXm+8J45OSHwVnOfY6V35b5XfZxgCvj5TV0mXCVdx4QAAAF8uWHB7wAABAMA
SDBGAiEAywu8Te3xzKNYwRup8r2yqBarJOqzTPv+43ay4/u8lZACIQD2CmGDyBOo
WMWrdSORUhPvAAGGDQ+AiVEr29SGK9RRCQB1AFGjsPX9AXmcVm24N3iPDKR6zBsn
y/eeiEKaDf7UiwXlAAABfLlhwfoAAAQDAEYwRAIgf8AhijjUmD/AY3HXlZcqNyFB
g6ZUlYjAAy4bcyqmYfgCIAfiUkbqx0lV3F9nm2KFCqrQ7T1qTQmd0xssEP0IXe70
AHYAQcjKsd8iRkoQxqE6CUKHXk4xixsD6+tLx2jwkGKWBvYAAAF8uWHBuAAABAMA
RzBFAiBk2pXuLzS2KFlvrgsWpBKinb/3AG+pgWSc1k4A8a0QeAIhAI8C1YuZelsI
zYV3zpHiyIlkxKTdlYg0lnP+/UwdHTaAMAoGCCqGSM49BAMDA2cAMGQCMBFSpqTJ
DKHJXG+cnPDLLcY/H0IwpBHKQ8M/qSHx0tRRW1uK5oRoGiDDDHYuDxtKGwIwSaH5
q4tyVvs3fmARob3fVmfy1gq53ktKYhZQwGFwr5q065Zfi4QRB9OVqiRbtx4a
-----END CERTIFICATE----- 
`
const dnssb = `-----BEGIN CERTIFICATE-----
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
-----END CERTIFICATE-----
`

type resolver struct {
	ip          string
	port        int
	rootCert    string
	readTimeOut uint
}

func NewResolver(ip string, port int, rto uint) proxy.Resolver {
	return &resolver{ip, port, dnssb, rto}
}

func (r *resolver) GetTLSConnection() (*tls.Conn, error) {
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM([]byte(r.rootCert)) {
		log.Println("Fail to parse rootCert")
		return nil, errors.New("Fail to parse rootCert")
	}
	dnsConn, err := tls.Dial(proxy.SocketTCP, r.ip+":"+strconv.Itoa(r.port), &tls.Config{
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

func (r *resolver) Resolve(um []byte) ([]byte, error) {
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
