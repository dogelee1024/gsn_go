package client

import "github.com/umbracle/ethgo/wallet"

// Current only support alg chain can deploy the contract
type AlgClient struct {
	EvmClient
}

func NewAlgClient(url string, key *wallet.Key) Client {
	client := NewEvmClient(url, key)
	evmClient := *client.(*EvmClient)

	return &AlgClient{
		EvmClient: evmClient,
	}
}
