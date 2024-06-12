package contract

import (
	"fmt"

	"github.com/openwallet1/gsn-go/common/contract/artifact"
	"github.com/openwallet1/gsn-go/common/errs"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/contract"
)

type IPaymaster struct {
	c *contract.Contract
}

type GasAndDataLimits struct {
	AcceptanceBudget        uint64
	PreRelayedCallGasLimit  uint64
	PostRelayedCallGasLimit uint64
	CalldataSizeLimit       uint64
}

func NewIPaymaster(addr ethgo.Address, opts ...contract.ContractOption) *IPaymaster {
	return &IPaymaster{c: contract.NewContract(addr, artifact.IPaymasterAbi(), opts...)}
}

func (i *IPaymaster) GetTrustedForwarder(block ...ethgo.BlockNumber) (string, error) {
	var out map[string]interface{}
	var ok bool

	out, err := i.c.Call("getTrustedForwarder", ethgo.EncodeBlock(block...))
	if err != nil {
		return "", err
	}

	retval0, ok := out["trustedForwarder"].(ethgo.Address)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return "", err
	}
	return retval0.Address().String(), nil
}

func (i *IPaymaster) GetRelayHub(block ...ethgo.BlockNumber) (string, error) {
	var out map[string]interface{}
	var ok bool

	out, err := i.c.Call("getRelayHub", ethgo.EncodeBlock(block...))
	if err != nil {
		return "", err
	}

	retval0, ok := out["relayHub"].(ethgo.Address)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return "", err
	}
	return retval0.Address().String(), nil
}

// 定义 GetGasAndDataLimits 函数来调用 getGasAndDataLimits 方法
func (i *IPaymaster) GetGasAndDataLimits(block ...ethgo.BlockNumber) (*GasAndDataLimits, error) {
	var out map[string]interface{}
	var ok bool

	out, err := i.c.Call("getGasAndDataLimits", ethgo.EncodeBlock(block...))
	if err != nil {
		return &GasAndDataLimits{}, err
	}

	// 解析返回值
	limitsMap, ok := out["limits"].(map[string]interface{})
	if !ok {
		return &GasAndDataLimits{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode limits")
	}

	acceptanceBudget, ok := limitsMap["acceptanceBudget"].(uint64)
	if !ok {
		return &GasAndDataLimits{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode acceptanceBudget")
	}

	preRelayedCallGasLimit, ok := limitsMap["preRelayedCallGasLimit"].(uint64)
	if !ok {
		return &GasAndDataLimits{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode preRelayedCallGasLimit")
	}

	postRelayedCallGasLimit, ok := limitsMap["postRelayedCallGasLimit"].(uint64)
	if !ok {
		return &GasAndDataLimits{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode postRelayedCallGasLimit")
	}

	calldataSizeLimit, ok := limitsMap["calldataSizeLimit"].(uint64)
	if !ok {
		return &GasAndDataLimits{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode calldataSizeLimit")
	}

	limits := &GasAndDataLimits{
		AcceptanceBudget:        acceptanceBudget,
		PreRelayedCallGasLimit:  preRelayedCallGasLimit,
		PostRelayedCallGasLimit: postRelayedCallGasLimit,
		CalldataSizeLimit:       calldataSizeLimit,
	}

	return limits, nil
}
