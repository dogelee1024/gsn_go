package errs

// 通用错误码
const (
	NetworkError             = 10000 //网络异常
	NetworkTimeoutError      = 10001 //网络超时
	ArgsError                = 10002 //输入参数错误
	CtxDeadlineExceededError = 10003 //上下文超时
	RecordNotFoundError      = 10008 //数据不存在
	InitInternalError        = 10009 //init 异常
	HttpArgsError            = 10010 //http 请求参数异常
	UnknownError             = 10011 //

	PrivateKeyNotFoundError  = 10100 //私钥不存在
	ContractExecuteError     = 10101 //合约调用失败
	ContractResultParseError = 10103 //合约结果解析失败
	GasCalculatorError       = 10104 //gas 计算异常

	RelayServerPingFailed = 10200
)
