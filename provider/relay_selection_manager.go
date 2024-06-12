package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/openwallet1/gsn-go/common"
	"github.com/openwallet1/gsn-go/common/errs"
	"github.com/openwallet1/gsn-go/common/log"
	trans "github.com/openwallet1/gsn-go/common/model/struct"
	"github.com/openwallet1/gsn-go/common/utils"
)

type RelaySelectionManager struct {
	ctx context.Context

	config                *GSNConfig
	gsnTransactionDetails *trans.GsnTransactionDetails
	isInitialized         bool
}

func NewRelaySelectionManager(ctx context.Context, config *GSNConfig, gsnTransactionDetails *trans.GsnTransactionDetails) *RelaySelectionManager {
	return &RelaySelectionManager{
		ctx:                   ctx,
		config:                config,
		gsnTransactionDetails: gsnTransactionDetails,
		isInitialized:         false,
	}
}

func (r *RelaySelectionManager) Initialized() {
	//todo
	r.isInitialized = true
}

func (r *RelaySelectionManager) SelectNextRelay(relayHub string, paymaster string) (*trans.RelaySelectionResult, error) {
	for {
		slice := r.getNextSlice()
		var relayInfo *trans.RelaySelectionResult
		if len(slice) > 0 {
			var err error
			relayInfo, err = r.nextRelayInternal(slice, relayHub, paymaster)
			if err != nil {
				return nil, err
			}
			if relayInfo == nil {
				continue
			}
		}
		return relayInfo, nil
	}
}

func (r *RelaySelectionManager) nextRelayInternal(relays []string, relayHub string, paymaster string) (*trans.RelaySelectionResult, error) {
	// log.ZInfo(r.ctx, "nextRelay: find fastest relay from: ", utils.StructToJsonString(relays))
	// allPingResults, err := r.waitForSuccess(relays, relayHub, paymaster)
	// if err != nil {
	// 	return nil, err
	// }
	// log.ZInfo(r.ctx, "race finished with a result: ", utils.StructToJsonString(allPingResults))
	// winner, skippedRelays := r.selectWinnerFromResult(allPingResults)
	// r.handleWaitForSuccessResults(allPingResults, skippedRelays, winner)
	// if winner == nil {
	// 	return nil, nil
	// }
	// if winner.PingResponse.RelayManagerAddress != "" {
	// 	return &RelaySelectionResult{
	// 		RelayInfo:       *winner,
	// 		UpdatedGasFees:  r.gsnTransactionDetails,
	// 		MaxDeltaPercent: 0,
	// 	}, nil
	// }
	// managerAddress := winner.PingResponse.RelayManagerAddress
	// r.logger.Debug("finding relay register info for manager address: " + managerAddress + "; known info: " + toJSON(winner))
	// event := r.knownRelaysManager.GetRelayInfoForManager(managerAddress)
	// if event == nil {
	// 	r.logger.Error("Could not find registration info in the RelayRegistrar for the selected preferred relay")
	// 	return nil, nil
	// }
	// relayInfo := *event
	// relayInfo.RelayUrl = winner.RelayInfo.RelayUrl
	return &trans.RelaySelectionResult{
		RelayInfo: &trans.RelayInfo{
			//PingResponse: winner.PingResponse,
			RelayInfo: trans.RegistrarRelayInfo{RelayUrl: relays[0]},
		},
		UpdatedGasFees: &trans.EIP1559Fees{
			MaxFeePerGas:         string(r.gsnTransactionDetails.MaxFeePerGas),
			MaxPriorityFeePerGas: string(r.gsnTransactionDetails.MaxPriorityFeePerGas),
		},
		MaxDeltaPercent: 0,
	}, nil
}

func (r *RelaySelectionManager) waitForSuccess(relays []string, relayHub string, paymaster string) (*trans.WaitForSuccessResults[*trans.RelayInfo], error) {
	asMap := make(map[string]string)
	for _, it := range relays {
		asMap[it] = it
	}
	asArray := make([]string, 0, len(asMap))
	for _, v := range asMap {
		asArray = append(asArray, v)
	}

	resultsChan := make(chan *trans.RelayInfo)
	errorsChan := make(chan error)
	for _, relay := range asArray {
		go func(relay string) {
			partialInfo, err := r.getRelayAddressPing(relay, relayHub, paymaster)
			if err != nil {
				errorsChan <- err
			} else {
				resultsChan <- partialInfo
			}
		}(relay)
	}
	var results []*trans.RelayInfo
	errors := make(map[string]error)
	timeout := time.After(time.Duration(r.config.WaitForSuccessPingGrace))
	for _, relay := range asArray {
		select {
		case result := <-resultsChan:
			results = append(results, result)
		case err := <-errorsChan:
			errors[relay] = err
		case <-timeout:
			return &trans.WaitForSuccessResults[*trans.RelayInfo]{Results: results, Errors: errors}, nil
		}
	}
	return &trans.WaitForSuccessResults[*trans.RelayInfo]{Results: results, Errors: errors}, nil
}

func (r *RelaySelectionManager) getRelayAddressPing(relayInfo string, relayHub string, paymaster string) (*trans.RelayInfo, error) {
	log.ZInfo(r.ctx, "getRelayAddressPing URL: ", relayInfo)
	httpHelper := common.NewHttpHelper(relayInfo)
	pingResponse, err := httpHelper.GetPingResponse(relayInfo, string(paymaster))
	if err != nil {
		return nil, err
	}
	if !pingResponse.Ready {
		return nil, errs.ErrRelayServerPingFailed
	}
	if !utils.IsSameAddress(string(relayHub), string(pingResponse.RelayHubAddress)) {
		return nil, errs.Wrap(errs.ErrRelayServerPingFailed, fmt.Sprintf("Client is using RelayHub %s while the server responded with RelayHub address %s", relayHub, pingResponse.RelayHubAddress))
	}
	//r.pingFilter(pingResponse)
	return &trans.RelayInfo{
		PingResponse: *pingResponse,
		RelayInfo:    trans.RegistrarRelayInfo{RelayUrl: relayInfo},
	}, nil
}

func (r *RelaySelectionManager) getNextSlice() []string {
	if !r.isInitialized {
		panic("init() not called")
	}
	// for _, relays := range r.remainingRelays {
	// 	bulkSize := min(r.config.WaitForSuccessSliceSize, len(relays))
	// 	slice := relays[:bulkSize]
	// 	if len(slice) > 0 {
	// 		return slice
	// 	}
	// }
	return r.config.PreferredRelays
}
