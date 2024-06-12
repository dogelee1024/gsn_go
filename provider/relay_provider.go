package provider

import (
	"context"

	"github.com/openwallet1/gsn-go/common/errs"
	"github.com/openwallet1/gsn-go/common/log"
	trans "github.com/openwallet1/gsn-go/common/model/struct"
	"github.com/openwallet1/gsn-go/common/utils"
)

type JsonRpcPayload struct {
	Params []*Payload `json:"params"`
}

type Payload struct {
	From string `json:"from"`
	To   string `json:"to"`
	Data string `json:"data"`
	Gas  string `json:"gas"`
}

// JsonRpcCallback represents the JSON-RPC callback function
type JsonRpcCallback func(error, interface{})

type RelayProvider struct {
	relayClient *RelayClient
	ctx         context.Context
}

func NewRelayProvider(ctx context.Context, rawConstructorInput *GSNUnresolvedConstructorInput) *RelayProvider {
	relayClient := NewRelayClient(ctx, rawConstructorInput)
	return &RelayProvider{
		ctx:         ctx,
		relayClient: relayClient,
	}
}

func (r *RelayProvider) Init() error {
	return r.relayClient.Init()
}

func (r *RelayProvider) SendTransaction(payload *Payload) (*trans.RelayingResult, error) {
	log.ZInfo(r.ctx, "Sending transaction", utils.StructToJsonString(payload))
	gsnTransactionDetails := &trans.GsnTransactionDetails{
		From: payload.From,
		To:   payload.To,
		Data: payload.Data,
		Gas:  payload.Gas,
	}
	gsnTransactionDetails, err := r.fixGasFees(gsnTransactionDetails)
	if err != nil {
		return nil, err
	}
	return r.relayClient.RelayTransaction(gsnTransactionDetails)
}

// func (r *RelayClient) ethSendTransaction(payload JsonRpcPayload, callback JsonRpcCallback) {
// 	log.Info("ethSendTransaction", "payload", utils.StructToJsonString(payload))
// 	var wg sync.WaitGroup
// 	var gsnTransactionDetails GsnTransactionDetails
// 	var err error

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		gsnTransactionDetails, err = r.fixGasFees(payload.Params[0])
// 		if err != nil {
// 			s.logger.Println(err)
// 			callback(err, nil)
// 			return
// 		}
// 	}()

// 	wg.Wait()

// 	if err != nil {
// 		return
// 	}

// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		result, err := s.relayClient.relayTransaction(gsnTransactionDetails)
// 		if err != nil {
// 			s.onRelayTransactionRejected(err, callback)
// 			return
// 		}
// 		s.onRelayTransactionFulfilled(result, payload, callback)
// 	}()

// 	wg.Wait()
// }

func (r *RelayProvider) fixGasFees(txDetails *trans.GsnTransactionDetails) (*trans.GsnTransactionDetails, error) {
	// if txDetails.MaxFeePerGas != "" && txDetails.MaxPriorityFeePerGas != "" {
	// 	txDetails.Gas = ""
	// 	return txDetails, nil
	// }
	// if txDetails.Gas != "" && txDetails.MaxFeePerGas == "" && txDetails.MaxPriorityFeePerGas == "" {
	// 	txDetails.MaxFeePerGas = txDetails.Gas
	// 	txDetails.MaxPriorityFeePerGas = txDetails.Gas
	// 	txDetails.Gas = ""
	// 	return txDetails, nil
	// }
	// if txDetails.Gas == "" && txDetails.MaxFeePerGas == "" && txDetails.MaxPriorityFeePerGas == "" {
	// 	maxPriorityFeePerGas, maxFeePerGas, err := r.relayClient.CalculateGasFees()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	txDetails.MaxPriorityFeePerGas = maxPriorityFeePerGas
	// 	txDetails.MaxFeePerGas = maxFeePerGas
	// 	return txDetails, nil
	// }
	maxPriorityFeePerGas, maxFeePerGas, err := r.relayClient.CalculateGasFees()
	if err != nil {
		return nil, errs.Wrap(errs.ErrGasCalculator, "Relay Provider: cannot provide only one of maxFeePerGas and maxPriorityFeePerGas")
	}
	txDetails.MaxPriorityFeePerGas = maxPriorityFeePerGas
	txDetails.MaxFeePerGas = maxFeePerGas
	return txDetails, nil
}

// func (r *RelayProvider) calculateGasFees() (string, string) {
// 	return r.RelayClient.calculateGasFees()
// }
