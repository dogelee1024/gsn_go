package provider

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"sync"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/openwallet1/gsn-go/common/errs"
	"github.com/openwallet1/gsn-go/common/model/eip712"
	"github.com/umbracle/ethgo/signing"
	"github.com/umbracle/ethgo/wallet"
)

type AccountKeypair struct {
	PrivateKey string
	Address    string
}

type AccountManager struct {
	signer   *wallet.EIP1155Signer
	accounts []*AccountKeypair
	chainID  uint64

	mu sync.Mutex
}

func NewAccountManager(chainID uint64) *AccountManager {
	accountManager := &AccountManager{
		chainID: chainID,
		signer:  wallet.NewEIP155Signer(chainID),
	}
	return accountManager
}

// RemoveHexPrefix removes the "0x" prefix from a hex string.
func removeHexPrefix(hexStr string) string {
	if strings.HasPrefix(hexStr, "0x") {
		return hexStr[2:]
	}
	return hexStr
}

func toAddress(privateKey string) (string, error) {
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return "", err
	}

	// Get the private key bytes
	privateKeyBytes := privateKeyECDSA.D.Bytes()

	walletKey, err := wallet.NewWalletFromPrivKey(privateKeyBytes)
	if err != nil {
		return "", err
	}
	return walletKey.Address().String(), nil
}

// AddAccount adds an account to the wallet using the given private key.
func (a *AccountManager) AddAccount(privateKey string) (AccountKeypair, error) {
	address, err := toAddress(privateKey)
	if err != nil {
		return AccountKeypair{}, err
	}

	keypair := AccountKeypair{
		PrivateKey: privateKey,
		Address:    strings.ToLower(address),
	}

	a.mu.Lock()
	a.accounts = append(a.accounts, &keypair)
	a.mu.Unlock()

	return keypair, nil
}

// NewAccount generates a new account and adds it to the wallet.
func (a *AccountManager) NewAccount() (AccountKeypair, error) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		return AccountKeypair{}, fmt.Errorf("failed to generate new private key: %w", err)
	}

	privateKey := hex.EncodeToString(crypto.FromECDSA(privKey))
	keypair, err := a.AddAccount(privateKey)
	if err != nil {
		return AccountKeypair{}, err
	}

	return keypair, nil
}

func (a *AccountManager) findAccountKey(from string) (*ecdsa.PrivateKey, error) {
	// Find the account by address
	var keypair *AccountKeypair
	for _, account := range a.accounts {
		if account.Address == strings.ToLower(from) {
			keypair = account
			break
		}
	}
	if keypair == nil {
		return nil, errs.ErrPrivateKeyNotFound
	}

	// Decode the private key
	privateKey := removeHexPrefix(keypair.PrivateKey)
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, errs.ErrPrivateKeyNotFound
	}

	privKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, errs.ErrPrivateKeyNotFound
	}

	return privKey, nil
}

func (a *AccountManager) Sign(domainSeparatorName string, relayRequest *eip712.RelayRequest) (string, error) {
	var signature string

	forwarder := relayRequest.RelayData.Forwarder
	cloneRequest := relayRequest

	signedData := eip712.NewTypedRequestData(domainSeparatorName, big.NewInt(int64(a.chainID)), forwarder, *cloneRequest)

	//privateKey, err := a.findAccountKey(string(relayRequest.Request.From))
	signature, err := a.SignTypedData(string(relayRequest.Request.From), signedData)
	if err != nil {
		return "", errs.ErrPrivateKeyNotFound
	}

	return signature, nil
}

// SignMessage signs a message using the private key of the specified address.
func (a *AccountManager) SignMessage(message string, from string) (string, error) {
	privKey, err := a.findAccountKey(from)
	if err != nil {
		return "", err
	}

	// Sign the message
	msg := accounts.TextHash([]byte(message))
	signature, err := crypto.Sign(msg, privKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	return hex.EncodeToString(signature), nil
}

func (a *AccountManager) SignTypedData(from string, typedData *signing.EIP712TypedData) (string, error) {
	privKey, err := a.findAccountKey(from)
	if err != nil {
		return "", err
	}

	// Encode the typed data
	// data, err := ethgo.e
	// if err != nil {
	// 	fmt.Println("Error encoding typed data:", err)
	// 	return
	// }

	// domainSeparator, err := hashStruct("EIP712Domain", typedData.Domain, typedData.Types)
	// if err != nil {
	// 	return "", err
	// }

	// messageHash, err := hashStruct(typedData.PrimaryType, typedData.Message, typedData.Types)
	// if err != nil {
	// 	return "", err
	// }

	// data := []byte(fmt.Sprintf("\x19\x01%s%s", domainSeparator.Hex(), messageHash.Hex()))
	data, err := typedData.Hash()
	if err != nil {
		fmt.Println("Error encoding typed data:", err)
		return "", err
	}

	//hash := crypto.Keccak256Hash(data)
	signature, err := sign(data, false, privKey)
	if err != nil {
		return "", err
	}

	return "0x" + hex.EncodeToString(signature), nil
}

func sign(sighash []byte, isCompressedKey bool, key *ecdsa.PrivateKey) ([]byte, error) {
	signature, err := btcec.SignCompact(btcec.S256(), (*btcec.PrivateKey)(key), sighash, false)
	if err != nil {
		return nil, err
	}

	// Convert to Ethereum signature format with 'recovery id' v at the end.
	v := signature[0]
	copy(signature, signature[1:])
	signature[64] = v
	return signature, nil
}

func hashStruct(primaryType string, data interface{}, types map[string][]eip712.TypedDataField) (common.Hash, error) {
	typeHash := crypto.Keccak256Hash([]byte(encodeType(primaryType, types)))
	encodedData, err := encodeData(primaryType, data, types)
	if err != nil {
		return common.Hash{}, err
	}

	hash := crypto.Keccak256Hash(append(typeHash.Bytes(), encodedData...))
	return hash, nil
}

func encodeType(primaryType string, types map[string][]eip712.TypedDataField) string {
	var result string
	result += primaryType + "("
	for _, field := range types[primaryType] {
		result += field.Type + " " + field.Name + ","
	}
	result = result[:len(result)-1] + ")"
	return result
}

func encodeData(primaryType string, data interface{}, types map[string][]eip712.TypedDataField) ([]byte, error) {
	fields := types[primaryType]
	encoded := []byte{}
	for _, field := range fields {
		value := reflect.ValueOf(data).MapIndex(reflect.ValueOf(field.Name)).Interface()
		encodedField, err := encodeField(field.Type, value, types)
		if err != nil {
			return nil, err
		}
		encoded = append(encoded, encodedField...)
	}
	return encoded, nil
}

func encodeField(fieldType string, value interface{}, types map[string][]eip712.TypedDataField) ([]byte, error) {
	switch fieldType {
	case "string":
		return crypto.Keccak256Hash([]byte(value.(string))).Bytes(), nil
	case "uint256":
		return common.LeftPadBytes(big.NewInt(value.(int64)).Bytes(), 32), nil
	case "address":
		return common.HexToAddress(value.(string)).Bytes(), nil
	default:
		if _, ok := types[fieldType]; ok {
			hash, err := hashStruct(fieldType, value, types)
			if err != nil {
				return nil, err
			}
			return hash.Bytes(), nil
		}
		return nil, fmt.Errorf("unsupported type: %s", fieldType)
	}
}
