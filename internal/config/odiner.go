package config

import (
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"math/big"
)

type Odiner interface {
	OdinChainID() string
	OdinEndpoint() string

	OdinSigner() (sdk.AccAddress, *secp256k1.PrivKey)
	OdinMemo() string

	OdinExchange() (string, int)

	OdinGasPrice() *big.Int
	OdinGasLimit() *big.Int
}

type odinChainConfig struct {
	Endpoint string   `yaml:"endpoint"`
	ChainID  string   `yaml:"chain_id"`
	Memo     string   `yaml:"memo"`
	GasPrice *big.Int `yaml:"gas_price"`
	GasLimit *big.Int `yaml:"gas_limit"`
}

type odinSignerConfig struct {
	Mnemonic string `yaml:"mnemonic"`
	Password string `yaml:"password"`
}

type odinExchangeConfig struct {
	Denom     string `yaml:"denom"`
	Precision int    `yaml:"precision"`
}

type odiner struct {
	Chain    odinChainConfig    `yaml:"chain"`
	Signer   odinSignerConfig   `yaml:"signer"`
	Exchange odinExchangeConfig `yaml:"exchange"`
}

func (o *odiner) OdinChainID() string {
	return o.Chain.ChainID
}

func (o *odiner) OdinEndpoint() string {
	return o.Chain.Endpoint
}

func (o *odiner) OdinSigner() (sdk.AccAddress, *secp256k1.PrivKey) {
	seed := bip39.NewSeed(o.Signer.Mnemonic, o.Signer.Password)
	master, ch := hd.ComputeMastersFromSeed(seed)

	key, err := hd.DerivePrivateKeyForPath(master, ch, sdk.FullFundraiserPath)
	if err != nil {
		panic(errors.Wrap(err, "failed to derive odin private key for path"))
	}

	pk := secp256k1.PrivKey{Key: key}
	accAddress := sdk.AccAddress(pk.PubKey().Address())

	return accAddress, &pk
}

func (o *odiner) OdinMemo() string {
	return o.Chain.Memo
}

func (o *odiner) OdinExchange() (string, int) {
	return o.Exchange.Denom, o.Exchange.Precision
}

func (o *odiner) OdinGasPrice() *big.Int {
	return o.Chain.GasPrice
}

func (o *odiner) OdinGasLimit() *big.Int {
	return o.Chain.GasLimit
}
