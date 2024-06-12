package eip712

type RelayData struct {
	MaxFeePerGas               string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas       string `json:"maxPriorityFeePerGas"`
	TransactionCalldataGasUsed string `json:"transactionCalldataGasUsed"`
	RelayWorker                string `json:"relayWorker"`
	Paymaster                  string `json:"paymaster"`
	PaymasterData              string `json:"paymasterData"`
	ClientId                   string `json:"clientId"`
	Forwarder                  string `json:"forwarder"`
}
