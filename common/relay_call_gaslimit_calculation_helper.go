package common

import (
	"math/big"
)

type RelayCallGasLimitCalculationHelper struct {
	ContractInteractor                      *ContractInteractor
	CalldataEstimationSlackFactor           int
	RelayHubCalculateChargeViewCallGasLimit *big.Int
}

func (r *RelayCallGasLimitCalculationHelper) AdjustRelayCallViewGasLimitForRelay(viewCallGasLimit *big.Int, workerAddress string, maxFeePerGas string) (*big.Int, error) {
	workerBalance, err := r.ContractInteractor.GetBalance(workerAddress)
	if err != nil {
		return nil, err
	}

	workerBalanceGasLimit := balanceToGas(workerBalance, big.NewInt(int64(r.ContractInteractor.relayHubConfiguration.PctRelayFee)), maxFeePerGas)

	if workerBalanceGasLimit.Cmp(viewCallGasLimit) < 0 {
		return workerBalanceGasLimit, nil
	}
	return viewCallGasLimit, nil
}

// balanceToGas 计算可用的 gas 数量
func balanceToGas(balance, pctRelayFeeDev *big.Int, maxFeePerGasStr string) *big.Int {
	maxFeePerGas, ok := new(big.Int).SetString(maxFeePerGasStr, 10)
	if !ok {
		return nil
	}

	// pctRelayFeeDev = pctRelayFeeDev + 100
	pctRelayFeeDev = new(big.Int).Add(pctRelayFeeDev, big.NewInt(100))

	// result = (balance / maxFeePerGas) * 100 / pctRelayFeeDev * 3 / 4
	result := new(big.Int).Div(balance, maxFeePerGas)
	result.Mul(result, big.NewInt(100))
	result.Div(result, pctRelayFeeDev)
	result.Mul(result, big.NewInt(3))
	result.Div(result, big.NewInt(4))

	return result
}
