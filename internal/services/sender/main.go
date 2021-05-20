package sender

import (
	"context"
	"github.com/bsc-bridge-svc/internal/config"
	"github.com/bsc-bridge-svc/internal/data"
	"github.com/bsc-bridge-svc/internal/data/postgres"
	"github.com/bsc-bridge-svc/odin"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"math/big"
)

type Service struct {
	cfg       config.Config
	ctx       context.Context
	log       *logrus.Logger
	transfers postgres.Transfers
	odin      odin.Client
}

func New(cfg config.Config, ctx context.Context) *Service {
	return &Service{
		cfg:       cfg,
		ctx:       ctx,
		log:       cfg.Logger(),
		transfers: postgres.NewTransfers(cfg),
		odin:      odin.New(ctx, cfg).WithSigner(),
	}
}

func (s *Service) StatusTransfer(transfer data.Transfer, status data.Status) error {
	transfer.Status = status
	err := s.transfers.UpdateTransfer(transfer)
	if err != nil {
		return errors.Wrap(err, "failed to update transfer")
	}
	return nil
}

// Send todo: refactor to several function
func (s *Service) Send() error {
	transfers, err := s.transfers.SelectStatus(data.StatusNotSent)
	if err != nil {
		return errors.Wrap(err, "failed to select transfers by status")
	}

	if len(transfers) == 0 {
		return nil
	}

	s.log.Info("Starting sending")
	for _, transfer := range transfers {
		rateCoef, err := s.odin.GetExchangeRate(transfer.Denom)
		if err != nil {
			return errors.Wrap(err, "failed to get exchangeDenom rate")
		}

		transferAmountRaw, ok := new(big.Int).SetString(transfer.Amount, 10)
		if !ok {
			err = errors.Wrap(err, "failed to convert amount to big int")
			newErr := s.StatusTransfer(transfer, data.StatusFailed)
			if newErr != nil {
				panic(errors.Wrap(newErr, "failed to mark status failed"))
			}
			return err
		}

		binanceToken, ok := s.cfg.BinanceToken(transfer.Denom)
		if !ok {
			err = errors.Wrap(err, "failed to find binance token in config")
			newErr := s.StatusTransfer(transfer, data.StatusFailed)
			if newErr != nil {
				panic(errors.Wrap(newErr, "failed to mark status failed"))
			}
			return err
		}
		exchangeDenom, odinPrecision := s.cfg.OdinExchange()
		transferAmount := sdk.NewDecFromBigIntWithPrec(transferAmountRaw, int64(odinPrecision-binanceToken.Precision)).Mul(rateCoef)
		s.log.WithFields(logrus.Fields{
			"withdrawal": transferAmount,
		}).Info("amount to transfer sending")

		coinAmount, _ := sdk.NewDecCoinFromDec(exchangeDenom, transferAmount).TruncateDecimal()
		err = s.odin.ClaimWithdrawal(transfer.Address, coinAmount)
		if err != nil {
			err = errors.Wrap(err, "failed to claim withdrawal")
			newErr := s.StatusTransfer(transfer, data.StatusFailed)
			if newErr != nil {
				panic(errors.Wrap(newErr, "failed to mark status failed"))
			}
			return err
		}

		err = s.StatusTransfer(transfer, data.StatusSent)
		if err != nil {
			panic(errors.Wrap(err, "failed to mark status failed"))
		}
	}
	s.log.Info("Finishing sending")
	return nil
}
