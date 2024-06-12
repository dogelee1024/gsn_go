package contract

import (
	"fmt"
	"math/big"

	"github.com/openwallet1/gsn-go/common/contract/artifact"
	"github.com/openwallet1/gsn-go/common/errs"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/contract"
)

type IRelayHub struct {
	c *contract.Contract
}

type RelayHubConfig struct {
	MaxWorkerCount      *big.Int
	GasReserve          *big.Int
	PostOverhead        *big.Int
	GasOverhead         *big.Int
	MinimumUnstakeDelay *big.Int
	DevAddress          string
	DevFee              uint8
	BaseRelayFee        *big.Int
	PctRelayFee         uint16
}

// type RelayCallABI struct {
// 	DomainSeparatorName string
// 	MaxAcceptanceBudget uint64
// 	RelayRequest        RelayRequest
// 	Signature           []byte
// 	ApprovalData        []byte
// }

func NewIRelayhub(addr ethgo.Address, opts ...contract.ContractOption) *IRelayHub {
	return &IRelayHub{c: contract.NewContract(addr, artifact.IRelayhubAbi(), opts...)}
}

func (i *IRelayHub) EncodeABI(method string, args ...interface{}) ([]byte, error) {
	m := i.c.GetABI().GetMethod(method)
	if m == nil {
		return nil, errs.Wrap(errs.ErrUnknown, fmt.Sprintln("method %s not found", method))
	}

	data, err := m.Encode(args)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (i *IRelayHub) GetConfiguration(block ...ethgo.BlockNumber) (*RelayHubConfig, error) {
	var out map[string]interface{}
	var ok bool

	out, err := i.c.Call("getConfiguration", ethgo.EncodeBlock(block...))
	if err != nil {
		return &RelayHubConfig{}, err
	}

	// 解析返回值
	configMap, ok := out["config"].(map[string]interface{})
	if !ok {
		return &RelayHubConfig{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode limits")
	}

	maxWorkerCount, ok := configMap["maxWorkerCount"].(*big.Int)
	if !ok {
		return &RelayHubConfig{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode maxWorkerCount")
	}

	gasReserve, ok := configMap["gasReserve"].(*big.Int)
	if !ok {
		return &RelayHubConfig{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode gasReserve")
	}

	gasOverhead, ok := configMap["gasOverhead"].(*big.Int)
	if !ok {
		return &RelayHubConfig{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode gasOverhead")
	}

	minimumUnstakeDelay, ok := configMap["minimumUnstakeDelay"].(*big.Int)
	if !ok {
		return &RelayHubConfig{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode minimumUnstakeDelay")
	}

	devAddress, ok := configMap["devAddress"].(string)
	if !ok {
		return &RelayHubConfig{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode devAddress")
	}

	devFee, ok := configMap["devFee"].(uint8)
	if !ok {
		return &RelayHubConfig{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode devFee")
	}

	baseRelayFee, ok := configMap["baseRelayFee"].(*big.Int)
	if !ok {
		return &RelayHubConfig{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode baseRelayFee")
	}

	pctRelayFee, ok := configMap["pctRelayFee"].(uint16)
	if !ok {
		return &RelayHubConfig{}, errs.Wrap(errs.ErrContractResultParse, "failed to decode pctRelayFee")
	}

	config := &RelayHubConfig{
		MaxWorkerCount:      maxWorkerCount,
		GasReserve:          gasReserve,
		GasOverhead:         gasOverhead,
		MinimumUnstakeDelay: minimumUnstakeDelay,
		DevAddress:          devAddress,
		DevFee:              devFee,
		BaseRelayFee:        baseRelayFee, // default
		PctRelayFee:         pctRelayFee,  // default
	}

	return config, nil
}
