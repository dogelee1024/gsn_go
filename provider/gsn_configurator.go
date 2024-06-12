package provider

import (
	"fmt"
	"math/big"

	"github.com/openwallet1/gsn-go/common"
)

type GSNUnresolvedConstructorInput struct {
	Config      *GSNConfig
	RelayParams *RelayParams
}

type RelayParams struct {
	NodeUrl    string
	PrivateKey string
	WalletType int
}

type GSNConfig struct {
	CalldataEstimationSlackFactor    int
	PreferredRelays                  []string
	BlacklistedRelays                []string
	PastEventsQueryMaxPageSize       int
	PastEventsQueryMaxPageCount      int
	GasPriceFactorPercent            float64
	GasPriceSlackPercent             float64
	GetGasFeesBlocks                 int
	GetGasFeesPercentile             int
	GasPriceOracleURL                string
	GasPriceOraclePath               string
	MinMaxPriorityFeePerGas          *big.Int
	MaxRelayNonceGap                 int
	RelayTimeoutGrace                int
	MethodSuffix                     string
	RequiredVersionRange             string
	JSONStringifyRequest             bool
	AuditorsCount                    int
	SkipERC165Check                  bool
	ClientID                         string
	RequestValidSeconds              int
	MaxViewableGasLimit              *big.Int
	MinViewableGasLimit              *big.Int
	Environment                      string
	MaxApprovalDataLength            int
	MaxPaymasterDataLength           int
	ClientDefaultConfigURL           string
	UseClientDefaultConfigURL        bool
	PerformDryRunViewRelayCall       bool
	PerformEstimateGasFromRealSender bool
	PaymasterAddress                 string
	TokenPaymasterDomainSeparators   map[string]string
	WaitForSuccessSliceSize          int
	WaitForSuccessPingGrace          int64
	DomainSeparatorName              string
}

const (
	GAS_PRICE_PERCENT               = 20
	GAS_PRICE_SLACK_PERCENT         = 80
	DEFAULT_RELAY_TIMEOUT_GRACE_SEC = 1800
	GSN_RUNTIME_VERSION             = "3.3.0-beta.10"
	GSN_REQUIRED_VERSION            = "^2.2.0"
	MAX_RELAY_NONCE_GAP             = 3
)

var (
	defaultGsnConfig = GSNConfig{
		CalldataEstimationSlackFactor:    1,
		PreferredRelays:                  []string{},
		BlacklistedRelays:                []string{},
		PastEventsQueryMaxPageSize:       int(^uint(0) >> 1),
		PastEventsQueryMaxPageCount:      20,
		GasPriceFactorPercent:            GAS_PRICE_PERCENT,
		GasPriceSlackPercent:             GAS_PRICE_SLACK_PERCENT,
		GetGasFeesBlocks:                 5,
		GetGasFeesPercentile:             50,
		GasPriceOracleURL:                "",
		GasPriceOraclePath:               "",
		MinMaxPriorityFeePerGas:          big.NewInt(1e9),
		MaxRelayNonceGap:                 MAX_RELAY_NONCE_GAP,
		RelayTimeoutGrace:                DEFAULT_RELAY_TIMEOUT_GRACE_SEC,
		MethodSuffix:                     "_v4",
		RequiredVersionRange:             GSN_REQUIRED_VERSION,
		JSONStringifyRequest:             true,
		AuditorsCount:                    0,
		SkipERC165Check:                  false,
		ClientID:                         "1",
		RequestValidSeconds:              172800,
		MaxViewableGasLimit:              big.NewInt(20000000),
		MinViewableGasLimit:              big.NewInt(300000),
		Environment:                      "defaultEnvironment",
		MaxApprovalDataLength:            0,
		MaxPaymasterDataLength:           0,
		ClientDefaultConfigURL:           fmt.Sprintf("https://client-config.opengsn.org/%s/client-config.json", GSN_RUNTIME_VERSION),
		UseClientDefaultConfigURL:        true,
		PerformDryRunViewRelayCall:       true,
		PerformEstimateGasFromRealSender: false,
		PaymasterAddress:                 "",
		TokenPaymasterDomainSeparators:   map[string]string{},
		WaitForSuccessSliceSize:          3,
		WaitForSuccessPingGrace:          3000,
		DomainSeparatorName:              "GSN Relayed Transaction",
	}
)

type GSNDependencies struct {
	httpClient *common.HttpHelper
	// logger?: LoggerInterface
	contractInteractor *common.ContractInteractor
	gasLimitCalculator *common.RelayCallGasLimitCalculationHelper
	// knownRelaysManager: KnownRelaysManager
	accountManager *AccountManager
	// transactionValidator: RelayedTransactionValidator
	// pingFilter: PingFilter
	// relayFilter: RelayFilter
	// asyncApprovalData: ApprovalDataCallback
	// asyncPaymasterData: PaymasterDataCallback
	// asyncSignTypedData?: SignTypedDataCallback
}
