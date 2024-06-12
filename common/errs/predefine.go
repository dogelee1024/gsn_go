package errs

var (
	ErrArgs           = NewCodeError(ArgsError, "ArgsError")
	ErrCtxDeadline    = NewCodeError(CtxDeadlineExceededError, "CtxDeadlineExceededError")
	ErrNetwork        = NewCodeError(NetworkError, "NetworkError")
	ErrNetworkTimeOut = NewCodeError(NetworkTimeoutError, "NetworkTimeoutError")
	ErrRecordNotFound = NewCodeError(RecordNotFoundError, "RecordNotFoundError")
	ErrInitInternal   = NewCodeError(InitInternalError, "InitInternalError")
	ErrHttpArgs       = NewCodeError(HttpArgsError, "HttpArgsError")
	ErrUnknown        = NewCodeError(UnknownError, "UnknownError")

	ErrPrivateKeyNotFound  = NewCodeError(PrivateKeyNotFoundError, "PrivateKeyNotFoundError")
	ErrContractExecute     = NewCodeError(ContractExecuteError, "ContractExecuteError")
	ErrContractResultParse = NewCodeError(ContractResultParseError, "ContractResultParseError")
	ErrGasCalculator       = NewCodeError(GasCalculatorError, "GasCalculatorError")

	ErrRelayServerPingFailed = NewCodeError(RelayServerPingFailed, "RelayServerPingFailed")
)
