package contract

import (
	"fmt"
	"math/big"

	"github.com/openwallet1/gsn-go/common/contract/artifact"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/contract"
)

type IForwarder struct {
	c *contract.Contract
}

func NewIForwarder(addr ethgo.Address, opts ...contract.ContractOption) *IForwarder {
	return &IForwarder{c: contract.NewContract(addr, artifact.IForwarderAbi(), opts...)}
}

func (i *IForwarder) GetNonce(from string, block ...ethgo.BlockNumber) (uint64, error) {
	var out map[string]interface{}
	var ok bool

	out, err := i.c.Call("getNonce", ethgo.EncodeBlock(block...), from)
	if err != nil {
		return 0, err
	}

	retval0, ok := out["0"].(*big.Int)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return 0, err
	}
	return retval0.Uint64(), nil
}
