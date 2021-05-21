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
	users     postgres.Users
	odin      odin.Client
}

func New(cfg config.Config, ctx context.Context) *Service {
	return &Service{
		cfg:       cfg,
		ctx:       ctx,
		log:       cfg.Logger(),
		transfers: postgres.NewTransfers(cfg),
		users:     postgres.NewUsers(cfg),
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
	transfers, err := s.transfers.New().SelectStatus(data.StatusNotSent)
	if err != nil {
		return errors.Wrap(err, "failed to select transfers by 'not sent'")
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

func (s *Service) Refund() error {
	transfers, err := s.transfers.New().SelectStatus(data.StatusFailed)
	if err != nil {
		return errors.Wrap(err, "failed to select transfers by failed")
	}

	if len(transfers) == 0 {
		return nil
	}

	s.log.Info("Staring refunding...")

	for _, transfer := range transfers {
		user, err := s.users.GetUserById(transfer.UserID)
		if err != nil {
			return errors.Wrap(err, "failed to get user by id")
		}

		userAmount, err := sdk.NewDecFromStr(user.Amount)
		if err != nil {
			return errors.Wrap(err, "failed to parse user coins")
		}
		s.log.WithField("amount", userAmount.String()).Info("user amount")

		refundAmount, err := sdk.NewDecFromStr(transfer.Amount)
		if err != nil {
			return errors.Wrap(err, "failed to parse refund coins")
		}
		s.log.WithField("amount", refundAmount.String()).Info("amount to refund")

		user.Amount = userAmount.Add(refundAmount).TruncateInt().String()

		s.log.WithField("amount", user.Amount).Info("amount after refund")

		if err := s.users.UpdateUser(*user); err != nil {
			return errors.Wrap(err, "failed to update user")
		}

		transfer.Status = data.StatusRefunded
		if err := s.transfers.UpdateTransfer(transfer); err != nil {
			return errors.Wrap(err, "failed to update transfer")
		}
	}
	s.log.Info("Finished refunding...")
	return nil
}
