package client

import (
	"context"
	"math/big"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
)

var _ Client = (*EvmClient)(nil)

type EvmClient struct {
	url       string
	ethClient *jsonrpc.Client
	ethKey    *wallet.Key

	status int
	ctx    context.Context
}

func NewEvmClient(url string, key *wallet.Key) Client {
	client := &EvmClient{
		url:    url,
		ethKey: key,
	}
	ethClient, err := jsonrpc.NewClient(url)
	if err != nil {
		panic(err)
	}
	client.ethClient = ethClient
	return client
}

func (e *EvmClient) Close() {
	e.ethClient.Close()
}

func (e *EvmClient) GetGasUnitPrice() (*big.Int, error) {
	gasPrice, err := e.ethClient.Eth().GasPrice()
	if err != nil {
		return nil, err
	}
	return big.NewInt(int64(gasPrice)), nil
}

func (e *EvmClient) GetEstimatedGas(to string, from string, data []byte) (uint64, error) {
	var (
		estimateGas uint64
		err         error
	)

	toAddr := ethgo.HexToAddress(to)
	contractAddress := ethgo.HexToAddress(from)

	msg := &ethgo.CallMsg{
		From: contractAddress,
		To:   &toAddr,
		Data: data,
	}
	estimateGas, err = e.ethClient.Eth().EstimateGas(msg)

	if err != nil {
		return 0, err
	}

	return estimateGas, nil
}
