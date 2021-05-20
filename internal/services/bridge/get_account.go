package bridge

import (
	"encoding/json"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"math/big"
	"net/http"
	"net/url"
)

func (s *Service) buildRequest(contractAddress, address ethcommon.Address) (*url.URL, error) {
	base, err := url.Parse(s.cfg.BinanceEndpoint())
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse binance endpoint url")
	}

	params := url.Values{}
	params.Add(Module, Account)
	params.Add(Action, TokenBalance)
	params.Add(ContractAddress, contractAddress.String())
	params.Add(Address, address.String())
	params.Add(Tag, Latest)
	params.Add(Apikey, s.cfg.BinanceApiKey())
	base.RawQuery = params.Encode()
	return base, nil
}

func (s *Service) GetAccount(contractAddress, address ethcommon.Address) (*big.Int, error) {
	urlPath, err := s.buildRequest(contractAddress, address)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build request")
	}

	req, _ := http.NewRequest(http.MethodGet, urlPath.String(), nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send request ot bscscan")
	}

	fullResp := Response{}
	if err := json.NewDecoder(resp.Body).Decode(&fullResp); err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}
	bigRes, ok := new(big.Int).SetString(fullResp.Result, 10)
	if !ok {
		return nil, errors.New("failed to parse response amount")
	}
	return bigRes, nil
}
