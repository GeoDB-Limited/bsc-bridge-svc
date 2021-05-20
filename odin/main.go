package odin

import (
	"context"
	odinapp "github.com/GeoDB-Limited/odin-core/app"
	odincoinswap "github.com/GeoDB-Limited/odin-core/x/coinswap/types"
	odinmint "github.com/GeoDB-Limited/odin-core/x/mint/types"
	"github.com/bsc-bridge-svc/internal/config"
	sdktxclient "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	sdkauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	sdkauth "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var (
	encoding = odinapp.MakeEncodingConfig()
)

// Client defines an interface for the wrapped cosmos sdk service client.
type Client interface {
	WithSigner() Client
	GetAccount(string) (sdkauth.AccountI, error)
	ClaimWithdrawal(string, sdk.Coin) error
	GetExchangeRate(string) (sdk.Dec, error)
}

// client defines typed wrapper for the cosmos sdk service client.
type client struct {
	connection *grpc.ClientConn
	cfg        config.Config
	context    context.Context
	signer     *signer
}

// signer defines data to sign the transactions
type signer struct {
	address    sdk.AccAddress
	privateKey *secp256k1.PrivKey
}

// New creates a client that uses the given cosmos sdk service client.
func New(ctx context.Context, cfg config.Config) Client {
	//ctx, _ = context.WithTimeout(ctx, 15*time.Second)
	conn, err := grpc.DialContext(ctx, cfg.OdinEndpoint(), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(errors.Wrapf(err, "failed to dial %s", cfg.OdinEndpoint()))
	}

	return &client{
		connection: conn,
		cfg:        cfg,
		context:    ctx,
	}
}

// WithSigner initializes odin signer to broadcast transactions
func (c *client) WithSigner() Client {
	odinapp.SetBech32AddressPrefixesAndBip44CoinType(sdk.GetConfig())
	address, pk := c.cfg.OdinSigner()

	return &client{
		connection: c.connection,
		cfg:        c.cfg,
		context:    c.context,
		signer: &signer{
			privateKey: pk,
			address:    address,
		},
	}
}

// ClaimWithdrawal claims withdrawing from Odin
func (c *client) ClaimWithdrawal(address string, amount sdk.Coin) error {
	receiverAddress, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return errors.Wrapf(err, "failed to parse receiver address: %s", address)
	}

	msg := odinmint.NewMsgWithdrawCoinsToAccFromTreasury(sdk.NewCoins(amount), receiverAddress, c.signer.address)
	txBytes, err := c.signTx(&msg)
	if err != nil {
		return errors.Wrapf(err, "failed to sign the transaction to to claim withdrawal with message: %s", msg.String())
	}

	serviceClient := tx.NewServiceClient(c.connection)
	resp, err := serviceClient.BroadcastTx(
		c.context,
		&tx.BroadcastTxRequest{
			Mode:    tx.BroadcastMode_BROADCAST_MODE_BLOCK,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to broadcast transaction")
	}

	if resp.TxResponse.Code != 0 {
		return errors.Errorf("failed to withdraw coins from minting module: %s", resp.TxResponse.RawLog)
	}

	return nil
}

// signTx signs the transaction with the given message
func (c *client) signTx(msg sdk.Msg) ([]byte, error) {
	txBuilder := encoding.TxConfig.NewTxBuilder()
	txBuilder.SetMemo(c.cfg.OdinMemo())
	denom, _ := c.cfg.OdinExchange()
	fee := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewIntFromBigInt(c.cfg.OdinGasPrice())))
	txBuilder.SetFeeAmount(fee)
	txBuilder.SetGasLimit(c.cfg.OdinGasLimit().Uint64())

	if err := txBuilder.SetMsgs(msg); err != nil {
		return nil, errors.Wrapf(err, "failed to set transaction builder message: %s", msg.String())
	}

	account, err := c.GetAccount(c.signer.address.String())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get account of signer: %s", c.signer.address.String())
	}
	accSequence := account.GetSequence()
	accNumber := account.GetAccountNumber()

	signV2 := signing.SignatureV2{
		PubKey: c.signer.privateKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  encoding.TxConfig.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: accSequence,
	}
	if err := txBuilder.SetSignatures(signV2); err != nil {
		return nil, errors.Wrap(err, "failed to set transaction builder signatures")
	}

	signerData := sdkauthsigning.SignerData{
		ChainID:       c.cfg.OdinChainID(),
		AccountNumber: accNumber,
		Sequence:      accSequence,
	}

	signV2, err = sdktxclient.SignWithPrivKey(
		encoding.TxConfig.SignModeHandler().DefaultMode(),
		signerData,
		txBuilder,
		c.signer.privateKey,
		encoding.TxConfig,
		accSequence,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign with private key")
	}

	err = txBuilder.SetSignatures(signV2)
	if err != nil {
		return nil, errors.Wrap(err, "failed to set transaction builder signatures")
	}

	txBytes, err := encoding.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get transaction bytes")
	}

	return txBytes, nil
}

// GetAccount returns the odin account by given address
func (c *client) GetAccount(address string) (sdkauth.AccountI, error) {
	authClient := sdkauth.NewQueryClient(c.connection)
	response, err := authClient.Account(c.context, &sdkauth.QueryAccountRequest{Address: address})
	if err != nil {
		return nil, errors.Wrap(err, "failed to query account")
	}

	var account sdkauth.AccountI
	if err := encoding.Marshaler.UnpackAny(response.Account, &account); err != nil {
		return nil, errors.Wrap(err, "failed to parse query response")
	}

	return account, nil
}

// GetExchangeRate returns rate of assets.
func (c *client) GetExchangeRate(from string) (sdk.Dec, error) {
	coinswapClient := odincoinswap.NewQueryClient(c.connection)
	denom, _ := c.cfg.OdinExchange()
	response, err := coinswapClient.Rate(
		c.context,
		&odincoinswap.QueryRateRequest{From: from, To: denom},
	)
	if err != nil {
		return sdk.Dec{}, errors.Wrap(err, "failed to query account")
	}

	return response.Rate, nil
}
