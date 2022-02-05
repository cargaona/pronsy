package config

import (
	"os"
	"strconv"
)

const TCP_PORT = "PROXY_CONFIG_TCP_PORT"
const UDP_PORT = "PROXY_CONFIG_UDP_PORT"
const PROXY_METHOD_PORT = "PROXY_CONFIG_METHOD"
const TCP_MAX_CONN_POOL = "PROXY_CONFIG_TCP_MAX_CONN_POOL"
const CACHE_TTL = "PROXY_CONFIG_CACHE_TTL"
const RESOLVER_READ_TO = "PROXY_RESOLVER_READ_TO"
const DNS_PROVIDER_HOST = "DNS_PROVIDER_HOST"

type config struct {
	Port           int
	TcpMaxConnPool int
	CacheTTL       int
	ResolverReadTo uint
	DnsHost        string
}

func GetConfig() config {

	dnsHost := os.Getenv(DNS_PROVIDER_HOST)

	cacheTTL, err := strconv.Atoi(os.Getenv(CACHE_TTL))
	if err != nil {
		panic("Config Err: Could not parse Cache TTL")
	}
	//tcpPort, err := strconv.Atoi(os.Getenv(TCP_PORT))
	//if err != nil {
	//	panic("Config Err: Could not parse TCP port")
	//}
	udpPort, err := strconv.Atoi(os.Getenv(UDP_PORT))
	if err != nil {
		panic("Config Err: Could not parse UDP port")
	}
	maxConnPool, err := strconv.Atoi(os.Getenv(TCP_MAX_CONN_POOL))
	if err != nil {
		panic("Config Err: Could not parse TCP_MAX_CONN_POOL.")
	}
	resolverReadTO, err := strconv.Atoi(os.Getenv(RESOLVER_READ_TO))
	if err != nil {
		panic("Config Err: Could not parse RESOLVER_READ_TO.")
	}
	//tcpProxyMehod := os.Getenv(PROXY_METHOD_PORT)
	//if tcpProxyMehod == "direct" {
	//	directTCPPRoxy = true
	//}

	return config{
		Port:           udpPort,
		TcpMaxConnPool: maxConnPool,
		CacheTTL:       cacheTTL,
		ResolverReadTo: uint(resolverReadTO),
		DnsHost:        dnsHost,
	}
}
