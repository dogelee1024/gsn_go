package artifact

import (
	"encoding/hex"
	"fmt"

	"github.com/umbracle/ethgo/abi"
)

var abiIPaymaster *abi.ABI

// IPaymasterAbi returns the abi of the IPaymaster contract
func IPaymasterAbi() *abi.ABI {
	return abiIPaymaster
}

var binIPaymaster []byte

func init() {
	var err error
	abiIPaymaster, err = abi.NewABI(abiIPaymasterStr)
	if err != nil {
		panic(fmt.Errorf("cannot parse IPaymaster abi: %v", err))
	}
	if len(binIPaymasterStr) != 0 {
		binIPaymaster, err = hex.DecodeString(binIPaymasterStr[2:])
		if err != nil {
			panic(fmt.Errorf("cannot parse IPaymaster bin: %v", err))
		}
	}
}

var binIPaymasterStr = ""

var abiIPaymasterStr = `[
	{
		"inputs": [],
		"name": "getGasAndDataLimits",
		"outputs": [
			{
				"components": [
					{
						"internalType": "uint256",
						"name": "acceptanceBudget",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "preRelayedCallGasLimit",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "postRelayedCallGasLimit",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "calldataSizeLimit",
						"type": "uint256"
					}
				],
				"internalType": "struct IPaymaster.GasAndDataLimits",
				"name": "limits",
				"type": "tuple"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "getRelayHub",
		"outputs": [
			{
				"internalType": "address",
				"name": "relayHub",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "getTrustedForwarder",
		"outputs": [
			{
				"internalType": "address",
				"name": "trustedForwarder",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "bytes",
				"name": "context",
				"type": "bytes"
			},
			{
				"internalType": "bool",
				"name": "success",
				"type": "bool"
			},
			{
				"internalType": "uint256",
				"name": "gasUseWithoutPost",
				"type": "uint256"
			},
			{
				"components": [
					{
						"internalType": "uint256",
						"name": "maxFeePerGas",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "maxPriorityFeePerGas",
						"type": "uint256"
					},
					{
						"internalType": "uint256",
						"name": "transactionCalldataGasUsed",
						"type": "uint256"
					},
					{
						"internalType": "address",
						"name": "relayWorker",
						"type": "address"
					},
					{
						"internalType": "address",
						"name": "paymaster",
						"type": "address"
					},
					{
						"internalType": "address",
						"name": "forwarder",
						"type": "address"
					},
					{
						"internalType": "bytes",
						"name": "paymasterData",
						"type": "bytes"
					},
					{
						"internalType": "uint256",
						"name": "clientId",
						"type": "uint256"
					}
				],
				"internalType": "struct GsnTypes.RelayData",
				"name": "relayData",
				"type": "tuple"
			}
		],
		"name": "postRelayedCall",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"components": [
					{
						"components": [
							{
								"internalType": "address",
								"name": "from",
								"type": "address"
							},
							{
								"internalType": "address",
								"name": "to",
								"type": "address"
							},
							{
								"internalType": "uint256",
								"name": "value",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "gas",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "nonce",
								"type": "uint256"
							},
							{
								"internalType": "bytes",
								"name": "data",
								"type": "bytes"
							},
							{
								"internalType": "uint256",
								"name": "validUntilTime",
								"type": "uint256"
							}
						],
						"internalType": "struct IForwarder.ForwardRequest",
						"name": "request",
						"type": "tuple"
					},
					{
						"components": [
							{
								"internalType": "uint256",
								"name": "maxFeePerGas",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "maxPriorityFeePerGas",
								"type": "uint256"
							},
							{
								"internalType": "uint256",
								"name": "transactionCalldataGasUsed",
								"type": "uint256"
							},
							{
								"internalType": "address",
								"name": "relayWorker",
								"type": "address"
							},
							{
								"internalType": "address",
								"name": "paymaster",
								"type": "address"
							},
							{
								"internalType": "address",
								"name": "forwarder",
								"type": "address"
							},
							{
								"internalType": "bytes",
								"name": "paymasterData",
								"type": "bytes"
							},
							{
								"internalType": "uint256",
								"name": "clientId",
								"type": "uint256"
							}
						],
						"internalType": "struct GsnTypes.RelayData",
						"name": "relayData",
						"type": "tuple"
					}
				],
				"internalType": "struct GsnTypes.RelayRequest",
				"name": "relayRequest",
				"type": "tuple"
			},
			{
				"internalType": "bytes",
				"name": "signature",
				"type": "bytes"
			},
			{
				"internalType": "bytes",
				"name": "approvalData",
				"type": "bytes"
			},
			{
				"internalType": "uint256",
				"name": "maxPossibleGas",
				"type": "uint256"
			}
		],
		"name": "preRelayedCall",
		"outputs": [
			{
				"internalType": "bytes",
				"name": "context",
				"type": "bytes"
			},
			{
				"internalType": "bool",
				"name": "rejectOnRecipientRevert",
				"type": "bool"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "bytes4",
				"name": "interfaceId",
				"type": "bytes4"
			}
		],
		"name": "supportsInterface",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "versionPaymaster",
		"outputs": [
			{
				"internalType": "string",
				"name": "",
				"type": "string"
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]
`
