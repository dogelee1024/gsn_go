package common

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"

	"github.com/openwallet1/gsn-go/common/constant"
	"github.com/openwallet1/gsn-go/common/contract"
	"github.com/openwallet1/gsn-go/common/log"
	"github.com/openwallet1/gsn-go/common/model/eip712"
	trans "github.com/openwallet1/gsn-go/common/model/struct"
	"github.com/openwallet1/gsn-go/common/utils"
	ctr "github.com/umbracle/ethgo/contract"
	"github.com/umbracle/ethgo/wallet"
)

type ContractInteractor struct {
	nodeUrl                       string
	key                           *wallet.Key
	domainSeparatorName           string
	calldataEstimationSlackFactor int
	ctx                           context.Context

	paymasterInstance *contract.IPaymaster
	relayHubInstance  *contract.IRelayHub
	forwarderInstance *contract.IForwarder

	deployment trans.GSNContractsDeployment

	relayHubConfiguration *contract.RelayHubConfig

	ethClient *jsonrpc.Client
}

type RelayCallABI struct {
	DomainSeparatorName string
	MaxAcceptanceBudget uint64
	RelayRequest        interface{}
	Signature           []byte
	ApprovalData        []byte
}

func NewContractInteractor(ctx context.Context, nodeUrl, privateKey string, walletType int32, deployment *trans.GSNContractsDeployment, domainSeparatorName string, calldataEstimationSlackFactor int) *ContractInteractor {
	contractInteractor := &ContractInteractor{
		nodeUrl:                       nodeUrl,
		ctx:                           ctx,
		deployment:                    *deployment,
		domainSeparatorName:           domainSeparatorName, // deprecated
		calldataEstimationSlackFactor: calldataEstimationSlackFactor,
	}
	key, err := wallet.NewWalletFromPrivKey([]byte(privateKey))
	if err != nil {
		return nil
	}
	contractInteractor.key = key
	contractInteractor.ethClient, _ = jsonrpc.NewClient(nodeUrl)
	contractInteractor.initialize()
	return contractInteractor
}

func (c *ContractInteractor) initialize() {
	c.resloveDeployment()
	if c.relayHubInstance != nil {
		c.relayHubConfiguration, _ = c.relayHubInstance.GetConfiguration()
	}

}

func (c *ContractInteractor) resloveDeployment() {
	c.resloveDeploymentFromPaymaster(c.deployment.PaymasterAddress)
}

func (c *ContractInteractor) resloveDeploymentFromPaymaster(paymasterAdderss string) error {
	c.paymasterInstance = contract.NewIPaymaster(ethgo.HexToAddress(paymasterAdderss), ctr.WithJsonRPC(c.ethClient.Eth()), ctr.WithSender(c.key))
	forwarder, err := c.paymasterInstance.GetTrustedForwarder()
	if err != nil || utils.IsSameAddress(constant.ZERO_ADDRESS, forwarder) {
		log.ZError(c.ctx, "get forwarder from paymaster", err)
		return err
	}

	relayhub, err := c.paymasterInstance.GetRelayHub()
	if err != nil || utils.IsSameAddress(constant.ZERO_ADDRESS, relayhub) {
		log.ZError(c.ctx, "get relayhub from paymaster", err)
		return err
	}
	c.deployment.ForwarderAddress = forwarder
	c.deployment.RelayHubAddress = relayhub
	c.resloveDeploymentFromRelayhub(relayhub)
	c.resloveDeploymentFromForwarder(forwarder)
	return nil
}

func (c *ContractInteractor) resloveDeploymentFromRelayhub(relayhubAddress string) {
	c.relayHubInstance = contract.NewIRelayhub(ethgo.HexToAddress(relayhubAddress), ctr.WithJsonRPC(c.ethClient.Eth()), ctr.WithSender(c.key))
}

func (c *ContractInteractor) resloveDeploymentFromForwarder(forwarderAddress string) {
	c.forwarderInstance = contract.NewIForwarder(ethgo.HexToAddress(forwarderAddress), ctr.WithJsonRPC(c.ethClient.Eth()), ctr.WithSender(c.key))
}

func (c *ContractInteractor) GetDeployment() trans.GSNContractsDeployment {
	return c.deployment
}

func (c *ContractInteractor) GetSenderNonce(sender string) (uint64, error) {
	nonce, err := c.forwarderInstance.GetNonce(sender)
	if err != nil {
		return 0, err
	}
	return nonce, nil
}

func (c *ContractInteractor) GetGasAndDataLimitsFromPaymaster() (*contract.GasAndDataLimits, error) {
	return c.paymasterInstance.GetGasAndDataLimits()
}

func (c *ContractInteractor) GetBlockNumberRightNow() (uint64, error) {
	return c.ethClient.Eth().BlockNumber()
}

func (c *ContractInteractor) GetBalance(address string) (*big.Int, error) {
	return c.ethClient.Eth().GetBalance(ethgo.HexToAddress(address), ethgo.Latest)
}

func (c *ContractInteractor) GetGasPrice() (uint64, error) {
	return c.ethClient.Eth().GasPrice()
}

func (c *ContractInteractor) GetTransactionCount(address string) (uint64, error) {
	return c.ethClient.Eth().GetNonce(ethgo.HexToAddress(address), ethgo.Latest)
}

func (c *ContractInteractor) EstimateCalldataCostForRequest(relayRequestOriginal *eip712.RelayRequest, maxApprovalDataLength, maxPaymasterDataLength int) (string, error) {

	// 保护原始对象，避免临时修改
	relayRequest := relayRequestOriginal

	relayRequest.RelayData.TransactionCalldataGasUsed = "0xffffffffff"
	relayRequest.RelayData.PaymasterData = "0x" + strings.Repeat("ff", maxPaymasterDataLength)
	// maxAcceptanceBudget := "0xffffffffff"
	// signature := "0x" + strings.Repeat("ff", 65)
	// approvalData := "0x" + strings.Repeat("ff", maxApprovalDataLength)

	maxAcceptanceBudget := uint64(0xffffffffff)
	signature := common.Hex2Bytes("0xff" + strings.Repeat("ff", 64))
	approvalData := common.Hex2Bytes("0x" + strings.Repeat("ff", maxApprovalDataLength))

	encodedData, err := c.relayHubInstance.EncodeABI("relayCall", c.domainSeparatorName, maxAcceptanceBudget, relayRequest, signature, approvalData)
	if err != nil {
		return "", err
	}

	calculatedCalldataGasUsed, err := c.calculateCalldataGasUsed(encodedData, ethgo.HexToAddress(string(relayRequest.Request.From)), ethgo.HexToAddress(string(relayRequest.Request.To)))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("0x%x", calculatedCalldataGasUsed), nil
}

func (c *ContractInteractor) EstimateInnerCallGasLimit(encodedData string, from ethgo.Address, to ethgo.Address) (uint64, error) {
	data, err := hex.DecodeString(encodedData)
	if err != nil {
		return 0, fmt.Errorf("failed to decode encoded data: %v", err)
	}

	gas, err := c.ethClient.Eth().EstimateGas(&ethgo.CallMsg{
		From: from,
		To:   &to,
		Data: data,
	})
	if err != nil {
		return 0, err
	}
	return gas * uint64(c.calldataEstimationSlackFactor), nil
}

func (c *ContractInteractor) calculateCalldataGasUsed(encodedData []byte, from ethgo.Address, to ethgo.Address) (uint64, error) {
	// data, err := hex.DecodeString("0xa9059cbb000000000000000000000000693438ff63fedd1559870fbe2b1ad4128b74468e000000000000000000000000000000000000000000000000a6cf9e50b8320000")
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to decode encoded data: %v", err)
	// }

	gas, err := c.ethClient.Eth().EstimateGas(&ethgo.CallMsg{
		From: from,
		To:   &to,
		Data: common.Hex2Bytes("0xa9059cbb000000000000000000000000693438ff63fedd1559870fbe2b1ad4128b74468e000000000000000000000000000000000000000000000000a6cf9e50b8320000"),
	})
	if err != nil {
		return 0, err
	}
	return gas * uint64(c.calldataEstimationSlackFactor), nil
}

func encodeABI(relayCall RelayCallABI) (string, error) {
	// 函数签名
	funcSignature := "relayCall(string,uint256,RelayRequest,bytes,bytes)"
	funcHash := crypto.Keccak256Hash([]byte(funcSignature)).Hex()

	// ABI 定义
	parsedABI, err := abi.JSON(bytes.NewReader([]byte(`
		[{"name":"relayCall","type":"function","inputs":[
			{"name":"domainSeparatorName","type":"string"},
			{"name":"maxAcceptanceBudget","type":"uint256"},
			{"name":"relayRequest","type":"RelayRequest"},
			{"name":"signature","type":"bytes"},
			{"name":"approvalData","type":"bytes"}]}]
	`)))
	if err != nil {
		return "", fmt.Errorf("failed to parse ABI: %v", err)
	}

	// 编码参数
	data, err := parsedABI.Pack("relayCall", relayCall.DomainSeparatorName, relayCall.MaxAcceptanceBudget, relayCall.RelayRequest, relayCall.Signature, relayCall.ApprovalData)
	if err != nil {
		return "", fmt.Errorf("failed to pack values: %v", err)
	}

	// 拼接函数选择器和编码参数
	finalData := common.Hex2Bytes(funcHash[2:10])
	finalData = append(finalData, data...)

	return hexutil.Encode(finalData), nil
}
