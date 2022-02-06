package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	Port            int
	TcpMaxConnPool  int
	CacheTTL        int
	ResolverTimeOut uint
	ProviderHost    string
	ProviderPort    int
}

func GetConfig() (*Config, error) {
	var s Config
	err := envconfig.Process("dnsproxy", &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
