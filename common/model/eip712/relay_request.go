package eip712

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/umbracle/ethgo/abi"
)

// RelayRequest represents the relay request structure
type RelayRequest struct {
	Request   *ForwardRequest `json:"request"`
	RelayData RelayData       `json:"relayData"`
}

// CloneRelayRequest deep copies the given RelayRequest
func CloneRelayRequest(relayRequest RelayRequest) RelayRequest {
	// Serialize the relayRequest to JSON
	data, err := json.Marshal(relayRequest)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return RelayRequest{}
	}

	// Deserialize the JSON back into a new RelayRequest
	var cloned RelayRequest
	err = json.Unmarshal(data, &cloned)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return RelayRequest{}
	}

	return cloned
}

// 去掉十六进制前缀
func removeHexPrefix(hexStr string) string {
	return hexStr[2:]
}

// 填充字符串到指定长度
func padStart(str string, length int, padChar string) string {
	for len(str) < length {
		str = padChar + str
	}
	return str
}

// 实现 getRelayRequestID 函数
func GetRelayRequestID(relayRequest RelayRequest, signature []byte) (string, error) {
	//types := []string{"address", "uint256", "bytes"}
	// parameters := []interface{}{
	// 	relayRequest.Request.From,
	// 	relayRequest.Request.Nonce,
	// 	signature,
	// }

	addressType, _ := abi.NewType("address")
	uintType, _ := abi.NewType("uint256")
	bytesType, _ := abi.NewType("bytes")
	types := []*abi.TupleElem{&abi.TupleElem{
		Name: "from",
		Elem: addressType,
	}, &abi.TupleElem{
		Name: "nonce",
		Elem: uintType,
	}, &abi.TupleElem{
		Name: "signature",
		Elem: bytesType,
	},
	}

	// // 创建 ABI 编码器
	// arguments := abi.Arguments{
	// 	{Type: addressType},
	// 	{Type: uintType},
	// 	{Type: bytesType},
	// }

	// Construct ABI types
	tupleType := abi.NewTupleType(types)

	// // 编码参数
	// encodedData, err := arguments.Pack(parameters...)
	encodedData, err := abi.Encode([]interface{}{relayRequest.Request.From, relayRequest.Request.Nonce, signature}, tupleType)

	if err != nil {
		return "", fmt.Errorf("failed to encode parameters: %v", err)
	}

	// 计算哈希值
	hash := crypto.Keccak256(encodedData)
	rawRelayRequestId := padStart(hex.EncodeToString(hash), 64, "0")
	prefixSize := 8

	// 创建前缀
	zeroPrefix := "0" + strings.Repeat("0", prefixSize-1)
	prefixedRelayRequestId := regexp.MustCompile(fmt.Sprintf("^.{%d}", prefixSize)).ReplaceAllString(rawRelayRequestId, zeroPrefix)
	return "0x" + prefixedRelayRequestId, nil
}
