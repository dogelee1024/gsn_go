package client

import (
	"math/big"
)

type Client interface {
	GetGasUnitPrice() (*big.Int, error)
	GetEstimatedGas(to string, from string, data []byte) (uint64, error)

	Close()
}
