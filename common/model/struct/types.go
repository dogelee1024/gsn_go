package trans

import (
	"time"
)

type PintRequest struct {
	Paymaster string `json:"paymaster"`
}

type PingResponse struct {
	RelayWorkerAddress      string `json:"relayWorkerAddress"`
	RelayManagerAddress     string `json:"relayManagerAddress"`
	RelayHubAddress         string `json:"relayHubAddress"`
	OwnerAddress            string `json:"ownerAddress"`
	MaxMaxFeePerGas         string `json:"maxMaxFeePerGas"`
	MinMaxFeePerGas         string `json:"minMaxFeePerGas"`
	MinMaxPriorityFeePerGas string `json:"minMaxPriorityFeePerGas"`
	MaxAcceptanceBudget     string `json:"maxAcceptanceBudget"`
	NetworkID               string `json:"networkId,omitempty"`
	ChainID                 string `json:"chainId,omitempty"`
	Ready                   bool   `json:"ready"`
	Version                 string `json:"version"`
}

type RegistrarRelayInfo struct {
	LastSeenBlockNumber  int64  `json:"lastSeenBlockNumber"`
	LastSeenTimestamp    int64  `json:"lastSeenTimestamp"`
	FirstSeenBlockNumber int64  `json:"firstSeenBlockNumber"`
	FirstSeenTimestamp   int64  `json:"firstSeenTimestamp"`
	RelayUrl             string `json:"relayUrl"`
	RelayManager         string `json:"relayManager"`
}

type RelayInfo struct {
	PingResponse PingResponse       `json:"pingResponse"`
	RelayInfo    RegistrarRelayInfo `json:"relayInfo"`
}

type RelayingAttempt struct {
	RelayRequestID string         `json:"relayRequestID,omitempty"`
	ValidUntilTime *time.Time     `json:"validUntilTime,omitempty"`
	Transaction    *Transaction   `json:"transaction,omitempty"`
	IsRelayError   *bool          `json:"isRelayError,omitempty"`
	Error          *error         `json:"error,omitempty"`
	AuditPromise   *AuditResponse `json:"auditPromise,omitempty"`
}

type AuditRequest struct {
	SignedTx string `json:"signedTx,omitempty"`
}

type AuditResponse struct {
	CommitTxHash string `json:"commitTxHash,omitempty"`
	Message      string `json:"message,omitempty"`
}

// AccessListEntry represents an entry in the AccessList.
type AccessListEntry struct {
	Address     string   `json:"address"`
	StorageKeys []string `json:"storageKeys"`
}

// AccessList is a slice of AccessListEntry.
type AccessList []AccessListEntry

type RelaySelectionResult struct {
	RelayInfo       *RelayInfo   `json:"relayInfo,omitempty"`
	MaxDeltaPercent int          `json:"maxDeltaPercent"`
	UpdatedGasFees  *EIP1559Fees `json:"updatedGasFees,omitempty"`
}

type EIP1559Fees struct {
	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
}

type WaitForSuccessResults[T any] struct {
	Results []T              `json:"results"`
	Errors  map[string]error `json:"errors"`
}
