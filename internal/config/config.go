package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	TCPMaxConnPool  int
	UDPMaxQueueSize int
	CacheEnabled    bool
	CacheTTL        int
	ResolverTimeOut uint
	ProviderHost    string
	ProviderPort    int
	Port            int
}

func GetConfig() (*Config, error) {
	var s Config
	err := envconfig.Process("pronsy", &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
