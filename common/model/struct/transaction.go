package trans

import (
	"math/big"

	"github.com/openwallet1/gsn-go/common/constant"
	eip712_model "github.com/openwallet1/gsn-go/common/model/eip712"
)

// Define RelayMetadata structure
type RelayMetadata struct {
	ApprovalData        string `json:"approvalData"`
	RelayHubAddress     string `json:"relayHubAddress"`
	RelayLastKnownNonce int    `json:"relayLastKnownNonce"`
	RelayRequestId      string `json:"relayRequestId"`
	RelayMaxNonce       int    `json:"relayMaxNonce"`
	Signature           string `json:"signature"`
	MaxAcceptanceBudget string `json:"maxAcceptanceBudget"`
	DomainSeparatorName string `json:"domainSeparatorName"`
}

// Define RelayTransactionRequest structure
type RelayTransactionRequest struct {
	RelayRequest eip712_model.RelayRequest `json:"relayRequest"`
	Metadata     RelayMetadata             `json:"metadata"`
}

type RelayTransactionResponse struct {
	SignedTx       constant.PrefixedHexString            `json:"signedTX"`
	NonceGapFilled map[string]constant.PrefixedHexString `json:"nonceGapFilled"`
}

type RelayingResult struct {
	RelayRequestID  string           `json:"relayRequestID"`
	Transaction     *Transaction     `json:"transaction"`
	AuditPromises   []*AuditResponse `json:"auditPromises"`
	SubmissionBlock uint64           `json:"submissionBlock"`
}

type GsnTransactionDetails struct {
	From                 string                      `json:"from"`
	Data                 string                      `json:"data"`
	To                   string                      `json:"to"`
	Value                *constant.IntString         `json:"value,omitempty"`
	Gas                  string                      `json:"gas,omitempty"`
	MaxFeePerGas         string                      `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string                      `json:"maxPriorityFeePerGas"`
	PaymasterData        *constant.PrefixedHexString `json:"paymasterData,omitempty"`
	ClientId             *constant.IntString         `json:"clientId,omitempty"`
	UseGSN               *bool                       `json:"useGSN,omitempty"`
}

type GSNContractsDeployment struct {
	ForwarderAddress         string `json:"forwarderAddress,omitempty"`
	PaymasterAddress         string `json:"paymasterAddress,omitempty"`
	PenalizerAddress         string `json:"penalizerAddress,omitempty"`
	RelayRegistrarAddress    string `json:"relayRegistrarAddress,omitempty"`
	RelayHubAddress          string `json:"relayHubAddress,omitempty"`
	StakeManagerAddress      string `json:"stakeManagerAddress,omitempty"`
	ManagerStakeTokenAddress string `json:"managerStakeTokenAddress,omitempty"`
}

type Transaction struct {
	Hash                 *string     `json:"hash,omitempty"`
	To                   *string     `json:"to,omitempty"`
	From                 *string     `json:"from,omitempty"`
	Nonce                uint64      `json:"nonce"`
	GasLimit             *big.Int    `json:"gasLimit"`
	GasPrice             *big.Int    `json:"gasPrice,omitempty"`
	Data                 string      `json:"data"`
	Value                *big.Int    `json:"value"`
	ChainId              int64       `json:"chainId"`
	R                    *string     `json:"r,omitempty"`
	S                    *string     `json:"s,omitempty"`
	V                    *int64      `json:"v,omitempty"`
	Type                 *int64      `json:"type,omitempty"`
	AccessList           *AccessList `json:"accessList,omitempty"`
	MaxPriorityFeePerGas *big.Int    `json:"maxPriorityFeePerGas,omitempty"`
	MaxFeePerGas         *big.Int    `json:"maxFeePerGas,omitempty"`
}
