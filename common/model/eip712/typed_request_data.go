package eip712

import (
	"math/big"

	"github.com/umbracle/ethgo/signing"
)

type TypedRequestData struct {
	Types       map[string][]TypedDataField
	Domain      EIP712Domain
	PrimaryType string
	Message     *Message
}

// 定义 EIP712Domain 结构体
type EIP712Domain struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	ChainId           uint   `json:"chainId"`
	VerifyingContract string `json:"verifyingContract"`
}

// 定义 EIP712Domain 结构体
type Message struct {
	From           string    `json:"from"`
	To             string    `json:"to"`
	Data           string    `json:"data"`
	Value          string    `json:"value"`
	Nonce          string    `json:"nonce"`
	Gas            string    `json:"gas"`
	ValidUntilTime string    `json:"validUntilTime"`
	RelayData      RelayData `json:"relayData"`
}

// TypedData is a struct for the EIP-712 typed data
type TypedData struct {
	Types       map[string][]TypedDataField `json:"types"`
	PrimaryType string                      `json:"primaryType"`
	Domain      EIP712Domain                `json:"domain"`
	Message     Message                     `json:"message"`
}

// TypedDataField represents a field in the typed data
type TypedDataField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

var (
	EIP712DomainType = []*signing.EIP712Type{
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"}}
	RelayDataType = []*signing.EIP712Type{
		{Name: "maxFeePerGas", Type: "uint256"},
		{Name: "maxPriorityFeePerGas", Type: "uint256"},
		{Name: "transactionCalldataGasUsed", Type: "uint256"},
		{Name: "relayWorker", Type: "address"},
		{Name: "paymaster", Type: "address"},
		{Name: "forwarder", Type: "address"}, {Name: "paymasterData", Type: "bytes"}, {Name: "clientId", Type: "uint256"}}
	ForwardRequestType = []*signing.EIP712Type{
		{Name: "from", Type: "address"},
		{Name: "to", Type: "address"},
		{Name: "value", Type: "uint256"},
		{Name: "gas", Type: "uint256"},
		{Name: "nonce", Type: "uint256"},
		{Name: "data", Type: "bytes"},
		{Name: "validUntilTime", Type: "uint256"},
	}
	RelayRequestType = []*signing.EIP712Type{
		{Name: "from", Type: "address"},
		{Name: "to", Type: "address"},
		{Name: "value", Type: "uint256"},
		{Name: "gas", Type: "uint256"},
		{Name: "nonce", Type: "uint256"},
		{Name: "data", Type: "bytes"},
		{Name: "validUntilTime", Type: "uint256"},
		{Name: "relayData", Type: "RelayData"},
	}
)

// 构造函数
func NewTypedRequestData(name string, chainId *big.Int, verifier string, relayRequest RelayRequest) *signing.EIP712TypedData {
	types := map[string][]*signing.EIP712Type{
		"EIP712Domain": EIP712DomainType,
		"RelayRequest": RelayRequestType,
		"RelayData":    RelayDataType,
	}
	domain := &signing.EIP712Domain{
		Name:              name,
		ChainId:           chainId,
		VerifyingContract: verifier,
		Version:           "3",
	}

	relayData := map[string]interface{}{
		"maxFeePerGas":               relayRequest.RelayData.MaxFeePerGas,
		"maxPriorityFeePerGas":       relayRequest.RelayData.MaxPriorityFeePerGas,
		"transactionCalldataGasUsed": relayRequest.RelayData.TransactionCalldataGasUsed,
		"relayWorker":                relayRequest.RelayData.RelayWorker,
		"paymaster":                  relayRequest.RelayData.Paymaster,
		"paymasterData":              relayRequest.RelayData.PaymasterData,
		"clientId":                   relayRequest.RelayData.ClientId,
		"forwarder":                  relayRequest.RelayData.Forwarder,
	}

	// message := map[string]interface{
	// 	From:           relayRequest.Request.From,
	// 	To:             relayRequest.Request.To,
	// 	Data:           relayRequest.Request.Data,
	// 	Value:          relayRequest.Request.Value,
	// 	Nonce:          relayRequest.Request.Nonce,
	// 	Gas:            relayRequest.Request.Gas,
	// 	ValidUntilTime: relayRequest.Request.ValidUntilTime,
	// 	RelayData:      relayRequest.RelayData,
	// }
	request := relayRequest.Request
	message := map[string]interface{}{
		"from":           request.From,
		"to":             request.To,
		"value":          request.Value,
		"gas":            request.Gas,
		"nonce":          request.Nonce,
		"data":           request.Data,
		"validUntilTime": request.ValidUntilTime,
		"relayData":      relayData,
	}

	return &signing.EIP712TypedData{
		Types:       types,
		Domain:      domain,
		PrimaryType: "RelayRequest",
		Message:     message,
	}
}
