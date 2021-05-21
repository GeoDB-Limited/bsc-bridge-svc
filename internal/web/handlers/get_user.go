package handlers

import (
	"fmt"
	"github.com/bsc-bridge-svc/internal/data"
	"github.com/bsc-bridge-svc/internal/web/ctx"
	"github.com/bsc-bridge-svc/internal/web/render"
	"github.com/bsc-bridge-svc/internal/web/requests"
	"github.com/bsc-bridge-svc/internal/web/utils"
	ethcommon "github.com/ethereum/go-ethereum/common"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"math/big"
	"net/http"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	log := ctx.Log(r)

	request, err := requests.NewGetUserRequest(r)
	if err != nil {
		if verr, ok := err.(validation.Errors); ok {
			log.WithError(verr).Debug("failed to parse get user request")
			render.Respond(w, http.StatusBadRequest, render.Message(fmt.Sprintf("request was invalid in some way: %s", verr.Error())))
			return
		}
		log.WithError(err).Error("something bad happened")
		render.Respond(w, http.StatusInternalServerError, render.Message("something bad happened parsing the request"))
		return
	}

	user, err := ctx.Users(r).GetUser(request.BinanceAddress, request.Denom)
	if err != nil {
		log.WithError(err).Error("failed to get user")
		render.Respond(w, http.StatusInternalServerError, render.Message("something bad happened"))
		return
	}

	binanceToken, ok := ctx.Config(r).BinanceToken(request.Denom)
	if !ok {
		log.Debug("unsupported denom")
		render.Respond(w, http.StatusBadRequest, render.Message("unsupported denom"))
		return
	}

	if user == nil {
		balanceAmount, err := ctx.Bridge(r).GetAccount(binanceToken.Address, ethcommon.HexToAddress(request.BinanceAddress))
		if err != nil {
			log.WithError(err).Error("failed to get account from binance")
			render.Respond(w, http.StatusInternalServerError, render.Message("something bad happened"))
			return
		}
		if balanceAmount == nil {
			log.Debug("account not found")
			render.Respond(w, http.StatusNotFound, render.Message(fmt.Sprintf("account with provided address not found: %s", request.BinanceAddress)))
			return
		}
		user, err = saveBalance(r, request, balanceAmount)
		if err != nil {
			log.WithError(err).Error("failed to save balance")
			render.Respond(w, http.StatusInternalServerError, render.Message("something bad happened"))
			return
		}
	}

	amountToWithdraw, err := utils.ParseAmount(request.Amount, binanceToken.Precision)
	if err != nil {
		log.WithError(err).Debug("failed to parse amountToWithdraw")
		render.Respond(w, http.StatusBadRequest, render.Message(fmt.Sprintf("request was invalid in some way: %s", err.Error())))
		return
	}

	log.WithField("request_amount", amountToWithdraw).Info("Parsed request amountToWithdraw")
	log.WithField("user_amount", user.Amount).Info("User amount")

	bigAmount, ok := new(big.Int).SetString(user.Amount, 10)
	if !ok {
		log.WithError(err).Debug(fmt.Sprintf("failed to parse amount: %s", user.Amount))
		render.Respond(w, http.StatusBadRequest, render.Message(fmt.Sprintf("failed to parse amount: %s", user.Amount)))
		return
	}
	remainder, neg := utils.SufficientAmount(bigAmount, amountToWithdraw)
	if neg {
		log.WithError(err).Debug("insufficient funds")
		render.Respond(w, http.StatusBadRequest, render.Message("insufficient funds"))
		return
	}

	err = ctx.Transfers(r).CreateTransfer(data.Transfer{
		Address: request.OdinAddress,
		Amount:  amountToWithdraw.String(),
		Denom:   request.Denom,
		Status:  data.StatusNotSent,
		UserID:  user.ID,
	})
	if err != nil {
		log.WithError(err).Debug("failed to create transfer")
		render.Respond(w, http.StatusInternalServerError, render.Message("failed to create transfer"))
		return
	}

	user.Amount = remainder.String()
	if err := ctx.Users(r).UpdateUser(*user); err != nil {
		log.WithError(err).Error("failed to update amount")
		render.Respond(w, http.StatusInternalServerError, render.Message("failed to update amount"))
		return
	}

	render.Respond(w, http.StatusOK, render.Message(user.ToReturn()))
}

func saveBalance(r *http.Request, request *requests.GetUserRequest, balanceAmount *big.Int) (*data.User, error) {
	user := data.User{
		Address: request.BinanceAddress,
		Amount:  balanceAmount.String(),
		Denom:   request.Denom,
	}
	err := ctx.Users(r).CreateUser(user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
