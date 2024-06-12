package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/openwallet1/gsn-go/common/errs"
	trans "github.com/openwallet1/gsn-go/common/model/struct"
)

func printCallerNameAndLine() string {
	pc, _, line, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name() + "()@" + strconv.Itoa(line) + ": "
}

func StructToJsonString(param interface{}) string {
	dataType, err := json.Marshal(param)
	if err != nil {
		panic(err)
	}
	dataString := string(dataType)
	return dataString
}

func JsonStringToStruct(s string, args interface{}) error {
	return errs.Wrap(json.Unmarshal([]byte(s), args), "json Unmarshal failed")
}

func Batch[T any, V any](fn func(T) V, ts []T) []V {
	if ts == nil {
		return nil
	}
	res := make([]V, 0, len(ts))
	for i := range ts {
		res = append(res, fn(ts[i]))
	}
	return res
}

func StringToBigInt(s string, precision int) (*big.Int, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("empty string")
	}

	dotIndex := strings.Index(s, ".")
	var intStr string
	scale := 1
	if dotIndex != -1 {
		scale = len(s) - dotIndex - 1
		intStr = s[:dotIndex] + s[dotIndex+1:]
	} else {
		intStr = s
		scale = 0
	}

	scaleFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision-scale)), nil)

	intVal := new(big.Int)
	_, success := intVal.SetString(intStr, 10)
	if !success {
		return nil, fmt.Errorf("failed to parse string as big.Int")
	}

	intVal.Mul(intVal, scaleFactor)

	return intVal, nil
}
func IsSameAddress(address1 string, address2 string) bool {
	return strings.ToLower(address1) == strings.ToLower(address2)
}

func CopyNonZeroValues(A, B interface{}) error {
	aVal := reflect.ValueOf(A)
	bVal := reflect.ValueOf(B)

	// Ensure A and B are pointers to structs
	if aVal.Kind() != reflect.Ptr || bVal.Kind() != reflect.Ptr {
		return fmt.Errorf("A and B must be pointers to structs")
	}

	aVal = aVal.Elem()
	bVal = bVal.Elem()

	// Ensure A and B are structs
	if aVal.Kind() != reflect.Struct || bVal.Kind() != reflect.Struct {
		return fmt.Errorf("A and B must be structs")
	}

	// Iterate over the fields of struct A
	for i := 0; i < aVal.NumField(); i++ {
		aField := aVal.Field(i)
		bField := bVal.Field(i)

		// Check if the field in A is non-zero
		if !isZeroValue(aField) {
			// Set the corresponding field in B
			if bField.CanSet() {
				bField.Set(aField)
			}
		}
	}

	return nil
}

func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Bool:
		return !v.Bool()
	case reflect.Map, reflect.Slice, reflect.Array:
		return v.Len() == 0
	case reflect.Struct:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	default:
		return v.Interface() == reflect.Zero(v.Type()).Interface()
	}
}

// IsAddress checks if a string is a valid Ethereum address
func IsAddress(value string) bool {
	return strings.HasPrefix(value, "0x") && len(value) == 42
}

// Utf8ToHex converts a UTF-8 string to its hex representation
func Utf8ToHex(value string) string {
	return "0x" + hex.EncodeToString([]byte(value))
}

// NumberToHex converts a number (string, int64, *big.Int) to its hex representation
func NumberToHex(value interface{}) string {
	var num *big.Int
	switch v := value.(type) {
	case string:
		num = new(big.Int)
		num.SetString(v, 10)
	case int64:
		num = big.NewInt(v)
	case *big.Int:
		num = v
	default:
		return ""
	}
	return "0x" + num.Text(16)
}

// ToHex converts different types of values to their hex representation
func ToHex(value interface{}, returnType ...string) string {
	if value == nil {
		return ""
	}

	rt := ""
	if len(returnType) > 0 {
		rt = returnType[0]
	}

	switch v := value.(type) {
	case string:
		if IsAddress(v) {
			if rt == "address" {
				return v
			}
			return "0x" + strings.ToLower(strings.TrimPrefix(v, "0x"))
		}
		if strings.HasPrefix(v, "0x") || strings.HasPrefix(v, "0X") {
			return v
		}
		if _, success := new(big.Int).SetString(v, 10); success {
			if strings.HasPrefix(v, "-0x") || strings.HasPrefix(v, "-0X") {
				return NumberToHex(v)
			}
			return NumberToHex(v)
		}
		if !isFinite(v) {
			return Utf8ToHex(v)
		}
		return Utf8ToHex(v)
	case bool:
		if rt == "bool" {
			return "bool"
		}
		if v {
			return "0x01"
		}
		return "0x00"
	case []byte:
		return "0x" + hex.EncodeToString(v)
	case *big.Int:
		if rt == "uint256" || rt == "int256" {
			return rt
		}
		return "0x" + v.Text(16)
	default:
		if reflect.ValueOf(value).Kind() == reflect.Struct {
			jsonBytes, _ := json.Marshal(value)
			return Utf8ToHex(string(jsonBytes))
		}
		return NumberToHex(value)
	}
}

// Check if a string is finite
func isFinite(value string) bool {
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		return false
	}
	return true
}

func ParseSignedTransaction(signedTxHex string) (*trans.Transaction, error) {
	// Decode the hex string to bytes
	txBytes, err := hex.DecodeString(signedTxHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %v", err)
	}

	// RLP decode the transaction bytes
	var tx *trans.Transaction
	err = rlp.DecodeBytes(txBytes, &tx)
	if err != nil {
		return nil, fmt.Errorf("failed to RLP decode transaction bytes: %v", err)
	}
	return tx, nil
}

func Deraw(rawTx string) (string, error) {

	// eip-1559 tx: 0xc2163f50770bd4bfd3c13848b405a56a451ae2a39cfa5a236ea2738ce44aa9df
	//rawTx := "02f8740181bf8459682f00851191460ee38252089497e542ec6b81dea28f212775ce8ac436ab77a7df880de0b6b3a764000080c080a02bc11202cee115fe22558ce2edb25c621266ce75f75e9b10da9a2ae72460ad4ea07d573eef31fdebf0f5f93eb7721924a082907419eb97a8dda0dd20a4a5b954a1"
	//rawTx := "02f904b58222da820c68843b9aca00844798dcae83071368943f1f0d5f8a50c82e71dae33a7dbe1e6207aa991880b904446ca862e200000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000045a4400000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000003a00000000000000000000000000000000000000000000000000000000000000420000000000000000000000000000000000000000000000000000000000000001747534e2052656c61796564205472616e73616374696f6e000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000001a0000000000000000000000000d3cff20db7af99bf6caf0a1247afe5bc63664172000000000000000000000000a6c1a5b6b52fb299edd8e9cf02b63c980176bfe30000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a84e000000000000000000000000000000000000000000000000000000000000017b00000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000666ac1090000000000000000000000000000000000000000000000000000000000000044a9059cbb000000000000000000000000693438ff63fedd1559870fbe2b1ad4128b74468e000000000000000000000000000000000000000000000000a6cf9e50b832000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004798dcae000000000000000000000000000000000000000000000000000000003b9aca000000000000000000000000000000000000000000000000000000000000006f5400000000000000000000000023bbd3ed8d813d087b122bb977cacf8e7f60d247000000000000000000000000cb798e55717978669bf3f9a29cc117c7ea37a4a2000000000000000000000000cfbfcfa99f4a8cebfdb90a1e424bb146c16a182b0000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041cf1a9e32880281a91593a0e81d2a0b0d06dd772f7c62082a92118d7d9bfcb07203eab3f6bb321fe9adea0e39120bd814a5443501570c7e253e77f288a4dc6e421b000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c001a09768cf6563478cc95a6c4e192c8055f1a5a751eb403ffae1eb05f27035eb39c2a01e681bf981bdf7c78ddea1acde4ac4ac751589fc25b58d7b28f096132e3e09b7"
	// legacy tx
	//rawTx := "f86d8202b28477359400825208944592d8f8d7b001e72cb26a73e4fa1806a51ac79d880de0b6b3a7640000802ca05924bde7ef10aa88db9c66dd4f5fb16b46dff2319b9968be983118b57bb50562a001b24b31010004f13d9a26b320845257a6cfc2bf819a3d55e3fc86263c5f0772"

	tx := &types.Transaction{}
	rawTxBytes, err := hex.DecodeString(rawTx)
	if err != nil {
		fmt.Println("err:", err)
		return "", err
	}

	err = tx.UnmarshalBinary(rawTxBytes)
	if err != nil {
		fmt.Println("err:", err)
		return "", err
	}

	spew.Dump(tx)

	hash := tx.Hash().Hex()
	fmt.Println(hash)

	return hash, nil
}

func main() {
	fmt.Println(ToHex("0x1234"))
	fmt.Println(ToHex("1234"))
	fmt.Println(ToHex(true))
	fmt.Println(ToHex([]byte{0x12, 0x34}))
	fmt.Println(ToHex(big.NewInt(1234)))
	fmt.Println(ToHex(map[string]string{"key": "value"}))
}
