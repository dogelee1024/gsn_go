package provider

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/openwallet1/gsn-go/common"
	"github.com/openwallet1/gsn-go/common/errs"
	"github.com/openwallet1/gsn-go/common/log"
	"github.com/openwallet1/gsn-go/common/model/eip712"
	trans "github.com/openwallet1/gsn-go/common/model/struct"
	"github.com/openwallet1/gsn-go/common/utils"
)

const (
	EmptyDataCallback = "0x"
)

type RelayClient struct {
	rawConstructorInput *GSNUnresolvedConstructorInput
	dependencies        *GSNDependencies
	initialized         bool
	ctx                 context.Context
	config              *GSNConfig
	relayParams         *RelayParams

	wrappedUnderlyingProvider *ethclient.Client
	mu                        sync.Mutex // 保护并发访问

}

func NewRelayClient(ctx context.Context, rawConstructorInput *GSNUnresolvedConstructorInput) *RelayClient {
	return &RelayClient{
		ctx:                 ctx,
		rawConstructorInput: rawConstructorInput,
	}
}

func (r *RelayClient) Init() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.initialized {
		return errs.ErrInitInternal
	}

	//r.dependencies = dependencies
	wrappedUnderlyingProvider, err := r.wrapInputProviderLike(r.rawConstructorInput.RelayParams.NodeUrl)
	if err != nil {
		return err
	}
	r.wrappedUnderlyingProvider = wrappedUnderlyingProvider
	r.InitInternal()
	r.initialized = true
	return nil
}

// InitInternal 内部初始化方法
func (r *RelayClient) InitInternal() error {
	// 解析配置
	config, err := r.resolveConfiguration(r.rawConstructorInput)
	if err != nil {
		return err
	}
	r.config = config
	r.relayParams = r.rawConstructorInput.RelayParams

	// 解析依赖项
	dependencies, err := r.resolveDependencies()
	if err != nil {
		return err
	}
	r.dependencies = dependencies

	// 进行 ERC165 检查
	// if !r.config.skipErc165Check {
	// 	if err := r.dependencies.contractInteractor._validateERC165InterfacesClient(); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (r *RelayClient) resolveConfiguration(input *GSNUnresolvedConstructorInput) (*GSNConfig, error) {
	config := input.Config
	if config == nil {
		config = &GSNConfig{}
	}

	utils.CopyNonZeroValues(config, defaultGsnConfig)
	defaultGsnConfig.PreferredRelays = config.PreferredRelays
	defaultGsnConfig.PaymasterAddress = config.PaymasterAddress
	return &defaultGsnConfig, nil
}

func (r *RelayClient) wrapInputProviderLike(nodeUrl string) (*ethclient.Client, error) {
	client, err := ethclient.Dial(nodeUrl)
	if err != nil {
		log.ZError(r.ctx, "Failed to connect to the node client:", err)
		return nil, err
	}
	return client, nil
}

func (r *RelayClient) resolveDependencies() (*GSNDependencies, error) {
	contractInteractor := common.NewContractInteractor(r.ctx, r.relayParams.NodeUrl, r.relayParams.PrivateKey, int32(r.relayParams.WalletType), &trans.GSNContractsDeployment{PaymasterAddress: r.config.PaymasterAddress}, r.config.DomainSeparatorName, r.config.CalldataEstimationSlackFactor)
	chainID, err := r.wrappedUnderlyingProvider.ChainID(r.ctx)
	if err != nil {
		return nil, err
	}
	accountManager := NewAccountManager(chainID.Uint64())
	accountManager.AddAccount(r.relayParams.PrivateKey)
	gasLimitCalculator := &common.RelayCallGasLimitCalculationHelper{
		ContractInteractor:                      contractInteractor,
		CalldataEstimationSlackFactor:           r.config.CalldataEstimationSlackFactor,
		RelayHubCalculateChargeViewCallGasLimit: r.config.MaxViewableGasLimit,
	}

	return &GSNDependencies{
		contractInteractor: contractInteractor,
		accountManager:     accountManager,
		gasLimitCalculator: gasLimitCalculator,
	}, nil
}

func (r *RelayClient) CalculateGasFees() (string, string, error) {
	gasFees, err := r.dependencies.contractInteractor.GetGasPrice()
	if err != nil {
		return "", "", nil
	}

	// Calculate priority fee
	pct := big.NewInt(int64(r.config.GasPriceFactorPercent + 100))
	priorityFee := new(big.Int).Mul(big.NewInt(int64(gasFees)), pct)
	priorityFee.Div(priorityFee, big.NewInt(100))

	if r.config.MinMaxPriorityFeePerGas != nil && priorityFee.Cmp(r.config.MinMaxPriorityFeePerGas) < 0 {
		priorityFee = r.config.MinMaxPriorityFeePerGas
	}

	maxPriorityFeePerGas := fmt.Sprintf("0x%x", priorityFee)

	// Calculate max fee
	maxFeePerGas := new(big.Int).Add(big.NewInt(int64(gasFees)), priorityFee)
	maxFeePerGas.Mul(maxFeePerGas, pct)
	maxFeePerGas.Div(maxFeePerGas, big.NewInt(100))

	if maxFeePerGas.Cmp(big.NewInt(0)) == 0 {
		maxFeePerGas = priorityFee
	}

	maxFeePerGasHex := fmt.Sprintf("0x%x", maxFeePerGas)

	return maxPriorityFeePerGas, maxFeePerGasHex, nil
}

func (r *RelayClient) RelayTransaction(gsnTransactionDetails *trans.GsnTransactionDetails) (*trans.RelayingResult, error) {
	if !r.initialized {
		r.Init()
	}

	gsnTransactionDetails.MaxFeePerGas = utils.ToHex(gsnTransactionDetails.MaxFeePerGas)
	gsnTransactionDetails.MaxPriorityFeePerGas = utils.ToHex(gsnTransactionDetails.MaxPriorityFeePerGas)

	if gsnTransactionDetails.Gas == "" {
		//todo estimateInnerCallGasLimit
		// from := r.dependencies.contractInteractor.GetDeployment().ForwarderAddress
		// data := gsnTransactionDetails.Data + strings.Replace(gsnTransactionDetails.From, "0x", "", 0)
		// value := "0"

		// txDetailsFromForwarder := gsnTransactionDetails
		// txDetailsFromForwarder.From = from
		// txDetailsFromForwarder.Data = data
		// txDetailsFromForwarder.Value = (*constant.IntString)(&value)
	}

	relayRequest, err := r.prepareRelayRequest(gsnTransactionDetails)
	if err != nil {
		return nil, errs.Wrap(err, "prepareRelayRequest error")
	}

	//gasAndDataLimits, err := r.dependencies.contractInteractor.GetGasAndDataLimitsFromPaymaster()

	//todo verifyDryRunSuccessful

	relaySelectionManager := NewRelaySelectionManager(r.ctx, r.config, gsnTransactionDetails)
	relaySelectionManager.Initialized()

	paymaster := r.dependencies.contractInteractor.GetDeployment().PaymasterAddress
	submissionBlock, err := r.dependencies.contractInteractor.GetBlockNumberRightNow()
	if err != nil {
		return nil, err
	}

	//for {
	relayHub := r.dependencies.contractInteractor.GetDeployment().RelayHubAddress
	relaySelectionResult, err := relaySelectionManager.SelectNextRelay(relayHub, paymaster)
	if err != nil {
		return nil, err
	}
	if relaySelectionResult != nil && relaySelectionResult.RelayInfo != nil {
		relayRequest.RelayData.MaxFeePerGas = relaySelectionResult.UpdatedGasFees.MaxFeePerGas
		relayRequest.RelayData.MaxPriorityFeePerGas = relaySelectionResult.UpdatedGasFees.MaxPriorityFeePerGas
	}

	relayingAttemp, err := r.attemptRelay(relaySelectionResult, relayRequest, nil)
	if err != nil {
		return nil, err
	}
	//}

	return &trans.RelayingResult{
		RelayRequestID:  relayingAttemp.RelayRequestID,
		Transaction:     relayingAttemp.Transaction,
		AuditPromises:   []*trans.AuditResponse{relayingAttemp.AuditPromise},
		SubmissionBlock: submissionBlock,
	}, nil
}

func (r *RelayClient) prepareRelayRequest(gsnTransactionDetails *trans.GsnTransactionDetails) (*eip712.RelayRequest, error) {
	relayHubAddress := r.dependencies.contractInteractor.GetDeployment().RelayHubAddress
	paymasterAddress := r.dependencies.contractInteractor.GetDeployment().PaymasterAddress
	forwarderAddress := r.dependencies.contractInteractor.GetDeployment().ForwarderAddress

	if relayHubAddress == "" || paymasterAddress == "" || forwarderAddress == "" {
		return nil, errs.Wrap(errs.ErrInitInternal, "Contract addresses are not initialized")
	}
	senderNonce, err := r.dependencies.contractInteractor.GetSenderNonce(string(gsnTransactionDetails.From))
	if err != nil {
		return nil, err
	}

	maxFeePerGasHex := gsnTransactionDetails.MaxFeePerGas
	maxPriorityFeePerGasHex := gsnTransactionDetails.MaxPriorityFeePerGas
	gasLimitHex := gsnTransactionDetails.Gas
	if maxFeePerGasHex == "" || maxPriorityFeePerGasHex == "" || gasLimitHex == "" {
		return nil, errs.Wrap(errs.ErrArgs, "'RelayClient internal exception.  gas fees or gas limit still not calculated. Cannot happen.")
	}
	if !strings.HasPrefix(string(maxFeePerGasHex), "0x") {
		return nil, errs.Wrap(errs.ErrArgs, `Invalid maxFeePerGas hex string: ${maxFeePerGasHex}`)
	}
	if !strings.HasPrefix(string(maxPriorityFeePerGasHex), "0x") {
		return nil, errs.Wrap(errs.ErrArgs, `Invalid maxPriorityFeePerGasHex hex string: ${maxFeePerGasHex}`)
	}
	if !strings.HasPrefix(gasLimitHex, "0x") {
		return nil, errs.Wrap(errs.ErrArgs, `Invalid gasLimitHex hex string: ${maxFeePerGasHex}`)
	}

	// 将十六进制字符串转换为十进制数
	gasLimit, err := strconv.ParseInt(gasLimitHex, 0, 64)
	if err != nil {
		return nil, errs.Wrap(err, "Error parsing gasLimitHex")
	}

	maxFeePerGas, err := strconv.ParseInt(string(maxFeePerGasHex), 0, 64)
	if err != nil {
		return nil, errs.Wrap(err, "Error parsing maxFeePerGasHex")
	}

	maxPriorityFeePerGas, err := strconv.ParseInt(string(maxPriorityFeePerGasHex), 0, 64)
	if err != nil {
		return nil, errs.Wrap(err, "Error parsing maxPriorityFeePerGasHex")
	}

	// 将十进制数转换为十进制字符串
	gasLimitStr := strconv.FormatInt(gasLimit, 10)
	maxFeePerGasStr := strconv.FormatInt(maxFeePerGas, 10)
	maxPriorityFeePerGasStr := strconv.FormatInt(maxPriorityFeePerGas, 10)

	value := "0"
	if gsnTransactionDetails.Value != nil {
		value = string(*gsnTransactionDetails.Value)
	}
	// 获取当前时间的秒数
	secondsNow := time.Now().Unix()
	// 计算有效时间
	validUntilTime := secondsNow + int64(r.config.RequestValidSeconds)

	return &eip712.RelayRequest{
		Request: &eip712.ForwardRequest{
			To:             gsnTransactionDetails.To,
			Data:           gsnTransactionDetails.Data,
			From:           gsnTransactionDetails.From,
			Value:          value,
			Nonce:          strconv.FormatUint(senderNonce, 10),
			Gas:            gasLimitStr,
			ValidUntilTime: strconv.FormatInt(validUntilTime, 10),
		},
		RelayData: eip712.RelayData{
			RelayWorker:                "",
			TransactionCalldataGasUsed: "",
			PaymasterData:              "0x",
			MaxFeePerGas:               maxFeePerGasStr,
			MaxPriorityFeePerGas:       maxPriorityFeePerGasStr,
			Paymaster:                  paymasterAddress,
			ClientId:                   r.config.ClientID,
			Forwarder:                  forwarderAddress,
		},
	}, nil
}

func (r *RelayClient) attemptRelay(relayInfo *trans.RelaySelectionResult, relayRequest *eip712.RelayRequest, viewCallGasLimit *big.Int) (*trans.RelayingAttempt, error) {
	log.ZInfo(r.ctx, "Attempting to relay", "transaction", utils.StructToJsonString(relayRequest))
	r.fillRelayInfo(relayRequest, relayInfo.RelayInfo)
	httpRequest, err := r.prepareRelayHttpRequest(relayRequest, relayInfo.RelayInfo)
	if err != nil {
		log.ZError(r.ctx, "PrepareRealyHttpRequest error", err)
		return nil, err
	}

	//adjustedRelayCallViewGasLimit, err := r.dependencies.gasLimitCalculator.AdjustRelayCallViewGasLimitForRelay(viewCallGasLimit, relayRequest.RelayData.RelayWorker, relayRequest.RelayData.MaxFeePerGas)

	//todo verifyViewCallSuccessful
	relayUrl := relayInfo.RelayInfo.RelayInfo.RelayUrl
	httpHelper := common.NewHttpHelper(relayUrl)

	fmt.Println(utils.StructToJsonString(httpRequest))

	relayTransactionResponse, err := httpHelper.RelayTransaction(httpRequest)
	if err != nil {
		log.ZError(r.ctx, "RelayTransaction error", err)
		return nil, err
	}
	signedTx := string(relayTransactionResponse.SignedTx)

	transactionHash, err := utils.Deraw(signedTx[2:])
	if err != nil {
		log.ZError(r.ctx, "ParseSignedTransaction error", err)
		return nil, err
	}

	auditResponse, err := r.auditTransaction(signedTx, relayUrl)
	if err != nil {
		log.ZError(r.ctx, "AuditTransaction error", err)
		return nil, err
	}
	return &trans.RelayingAttempt{
		RelayRequestID: httpRequest.Metadata.RelayRequestId,
		Transaction: &trans.Transaction{
			Hash: &transactionHash,
		},
		AuditPromise: auditResponse,
	}, nil
}

func (r *RelayClient) auditTransaction(hexTransaction string, sourceRelayUrl string) (*trans.AuditResponse, error) {
	// const auditors = this.dependencies.knownRelaysManager.getAuditors([sourceRelayUrl])
	// let failedAuditorsCount = 0
	// for (const auditor of auditors) {
	//   try {
	//     const penalizeResponse = await this.dependencies.httpClient.auditTransaction(auditor, hexTransaction)
	//     if (penalizeResponse.commitTxHash != null) {
	//       return penalizeResponse
	//     }
	//   } catch (e) {
	//     failedAuditorsCount++
	//     this.logger.info(`Audit call failed for relay at URL: ${auditor}. Failed audit calls: ${failedAuditorsCount}/${auditors.length}`)
	//   }
	// }
	// if (auditors.length === failedAuditorsCount && failedAuditorsCount !== 0) {
	//   this.logger.error('All auditors failed!')
	// }
	// return {
	//   message: `Transaction was not audited. Failed audit calls: ${failedAuditorsCount}/${auditors.length}`
	// }
	httpHelper := common.NewHttpHelper(sourceRelayUrl)
	return httpHelper.AuditTransaction(&trans.AuditRequest{
		SignedTx: hexTransaction,
	})
}

func (r *RelayClient) fillRelayInfo(relayRequest *eip712.RelayRequest, relayInfo *trans.RelayInfo) {
	//relayRequest.RelayData.RelayWorker = string(relayInfo.PingResponse.RelayWorkerAddress)
	relayRequest.RelayData.RelayWorker = "0x23bbd3ed8d813d087b122bb977cacf8e7f60d247"
	// transactionCalldataGas, err := r.dependencies.contractInteractor.EstimateCalldataCostForRequest(relayRequest, r.config.MaxApprovalDataLength, r.config.MaxPaymasterDataLength)
	// if err != nil {
	// 	return
	// }
	relayRequest.RelayData.TransactionCalldataGasUsed = "0x6f54"
}

func (r *RelayClient) prepareRelayHttpRequest(relayRequest *eip712.RelayRequest, relayInfo *trans.RelayInfo) (*trans.RelayTransactionRequest, error) {
	//todo switchSigner(relayRequest.Request.From)
	signature, err := r.dependencies.accountManager.Sign(r.config.DomainSeparatorName, relayRequest)
	if err != nil {
		return nil, err
	}

	relayRequestID, err := eip712.GetRelayRequestID(*relayRequest, ethCommon.Hex2Bytes(signature))
	if err != nil {
		return nil, err
	}
	approvalData := EmptyDataCallback

	//relayLastKnownNonce, err := r.dependencies.contractInteractor.GetTransactionCount(string(relayInfo.PingResponse.RelayWorkerAddress))
	relayLastKnownNonce, err := r.dependencies.contractInteractor.GetTransactionCount("0x23bbd3ed8d813d087b122bb977cacf8e7f60d247")

	if err != nil {
		return nil, err
	}
	relayMaxNonce := relayLastKnownNonce + uint64(r.config.MaxRelayNonceGap)
	relayHubAddress := r.dependencies.contractInteractor.GetDeployment().RelayHubAddress

	metadata := trans.RelayMetadata{
		DomainSeparatorName: r.config.DomainSeparatorName,
		//MaxAcceptanceBudget: relayInfo.PingResponse.MaxAcceptanceBudget,
		MaxAcceptanceBudget: "285252",
		RelayHubAddress:     relayHubAddress,
		RelayRequestId:      relayRequestID,
		Signature:           signature,
		ApprovalData:        approvalData,
		RelayMaxNonce:       int(relayMaxNonce),
		RelayLastKnownNonce: int(relayLastKnownNonce),
	}
	httpRequest := &trans.RelayTransactionRequest{
		RelayRequest: *relayRequest,
		Metadata:     metadata,
	}
	return httpRequest, nil
}
