package main

import (
	"dns-proxy/internal/config"
	"dns-proxy/pkg/controller/server"
	"dns-proxy/pkg/domain/proxy"
	"dns-proxy/pkg/gateway/cache"
	"dns-proxy/pkg/gateway/logger"
	"dns-proxy/pkg/gateway/parser"
	"dns-proxy/pkg/gateway/resolver"
	"time"
)

func main() {
	// Get the config of the application.
	cfg := config.GetConfig()

	// Create and start cache autopurge.
	cache := cache.New(
		time.Duration(cfg.CacheTTL)*time.Second,
		logger.New("CACHE"),
	)

	go cache.AutoPurge()

	// Create DNS Proxy injecting dependencies.
	proxySvc := proxy.NewDNSProxy(
		resolver.NewResolver(cfg.DnsHost, 853, cfg.ResolverReadTo),
		parser.NewDNSParser(),
		cache,
		logger.New("PROXY"),
	)

	// Start the TCP and UDP servers.
	DNSProxy := server.New(
		proxySvc,
		cfg.Port,
		"0.0.0.0",
		cfg.TcpMaxConnPool,
	)

	go DNSProxy.StartTCPServer()
	DNSProxy.StartUDPServer()
}
