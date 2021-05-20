package requests

import (
	"encoding/json"
	ethcommon "github.com/ethereum/go-ethereum/common"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
	"net/http"
)

type GetUserRequest struct {
	BinanceAddress string `json:"binance_address"`
	OdinAddress    string `json:"odin_address"`
	Amount         string `json:"amount"`
	Denom          string `json:"denom"`
}

func (r GetUserRequest) Validate() error {
	return validation.Errors{
		"binance_address": validation.Validate(r.BinanceAddress, validation.Required),
		"odin_address":    validation.Validate(r.OdinAddress, validation.Required),
		"amount":          validation.Validate(r.Amount, validation.Required),
		"denom":           validation.Validate(r.Denom, validation.Required),
	}.Filter()
}

func NewGetUserRequest(r *http.Request) (*GetUserRequest, error) {
	req := GetUserRequest{}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode request body")
	}

	if !ethcommon.IsHexAddress(req.BinanceAddress) {
		return nil, validation.Errors{
			"address": errors.New("address is not hex allowed"),
		}
	}

	return &req, req.Validate()
}
