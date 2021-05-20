package config

import (
	"database/sql"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"math/big"
)

type Config interface {
	Listener
	Logger
	Binancer
	Databaser
	Odiner
}

type config struct {
	Listener *listener  `yaml:"listener"`
	Log      *logger    `yaml:"log"`
	Binance  *binancer  `yaml:"binance"`
	Database *databaser `yaml:"db"`
	Odin     *odiner    `yaml:"odin"`
}

func (c config) BinanceApiKey() string {
	return c.Binance.ApiKey
}

func (c config) BinanceToken(token string) (BinanceToken, bool) {
	return c.Binance.BinanceToken(token)
}

func (c config) Address() string {
	return c.Listener.Address()
}

func (c config) Logger() *logrus.Logger {
	return c.Log.Logger()
}

func (c config) BinanceEndpoint() string {
	return c.Binance.BinanceEndpoint()
}

func (c config) DB() *sql.DB {
	return c.Database.DB()
}

func (c config) OdinChainID() string {
	return c.Odin.OdinChainID()
}

func (c config) OdinEndpoint() string {
	return c.Odin.OdinEndpoint()
}

func (c config) OdinSigner() (sdk.AccAddress, *secp256k1.PrivKey) {
	return c.Odin.OdinSigner()
}

func (c config) OdinMemo() string {
	return c.Odin.OdinMemo()
}

func (c config) OdinExchange() (string, int) {
	return c.Odin.OdinExchange()
}

func (c config) OdinGasPrice() *big.Int {
	return c.Odin.OdinGasPrice()
}

func (c config) OdinGasLimit() *big.Int {
	return c.Odin.OdinGasLimit()
}

func New(path string) Config {
	cfg := config{}

	yamlConfig, err := ioutil.ReadFile(path)
	if err != nil {
		panic(errors.Wrap(err, fmt.Sprintf("failed to read config: %s", path)))
	}

	err = yaml.Unmarshal(yamlConfig, &cfg)
	if err != nil {
		panic(errors.Wrap(err, fmt.Sprintf("failed to unmarshal config: %s", path)))
	}

	return &cfg
}
