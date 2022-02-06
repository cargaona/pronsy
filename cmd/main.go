package main

import (
	"dns-proxy/internal/config"
	"dns-proxy/pkg/controller/tcp"
	"dns-proxy/pkg/controller/udp"
	"dns-proxy/pkg/domain/proxy"
	"dns-proxy/pkg/gateway/cache"
	"dns-proxy/pkg/gateway/logger"
	"dns-proxy/pkg/gateway/parser"
	"dns-proxy/pkg/gateway/resolver"
	"fmt"
	"log"
	"runtime"
	"time"
)

func main() {
	// Get the config of the application.
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", cfg)

	// Create and start cacheUDP autopurge.
	cacheUDP := cache.New(
		time.Duration(cfg.CacheTTL)*time.Second,
		logger.New("UDP CACHE", true),
		cfg.CacheEnabled,
	)

	cacheTCP := cache.New(
		time.Duration(cfg.CacheTTL)*time.Second,
		logger.New("TCP CACHE", true),
		cfg.CacheEnabled,
	)

	go cacheTCP.AutoPurge()
	go cacheUDP.AutoPurge()

	// Create DNS Proxy injecting dependencies.
	proxySvc := proxy.NewDNSProxy(
		resolver.New(cfg.ProviderHost, cfg.ProviderPort, cfg.ResolverTimeOut),
		parser.NewDNSParser(),
		cacheUDP,
		logger.New("PROXY", true),
	)

	// Start the TCP and UDP servers.
	UDPDNSProxy := udp.New(
		proxySvc,
		udp.NewUDPHandler(
			2400,
			1000,
			logger.New("UDP HANDLER", true),
			cacheUDP,
			parser.NewDNSParser(),
		),
		logger.New("UDP SERVER", true),
		cfg.Port,
		runtime.NumCPU(),
	)

	TCPDNSProxy := tcp.New(
		proxySvc,
		tcp.NewTCPHandler(
			2400,
			logger.New("TCP HANDLER", true),
			cacheTCP,
			parser.NewDNSParser(),
		),
		logger.New("TCP SERVER", true),
		cfg.Port,
		"0.0.0.0",
		cfg.TcpMaxConnPool,
	)

	go TCPDNSProxy.Serve()
	UDPDNSProxy.Serve()
}
