package config

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
)

type Binancer interface {
	BinanceEndpoint() string
	BinanceApiKey() string
	BinanceToken(string) (BinanceToken, bool)
}

type BinanceChainConfig struct {
	Endpoint string `yaml:"endpoint"`
}

type binanceToken struct {
	Address     string `yaml:"address"`
	Precision   int    `yaml:"precision"`
}

type BinanceToken struct {
	Address     ethcommon.Address
	Precision   int
}

type binancer struct {
	Chain  BinanceChainConfig      `yaml:"chain"`
	Tokens map[string]binanceToken `yaml:"tokens"`
	ApiKey string                  `yaml:"api_key"`
}

func (b *binancer) BinanceEndpoint() string {
	return b.Chain.Endpoint
}

func (b *binancer) BinanceToken(denom string) (BinanceToken, bool) {
	token, ok := b.Tokens[denom]
	if !ok {
		return BinanceToken{}, false
	}
	return BinanceToken{
		Address:     ethcommon.HexToAddress(token.Address),
		Precision:   token.Precision,
	}, true
}

func (b *binancer) BinanceApiKey() string {
	return b.ApiKey
}
