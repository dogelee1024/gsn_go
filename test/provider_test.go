package test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/openwallet1/gsn-go/common/utils"
	"github.com/openwallet1/gsn-go/provider"
	"github.com/stretchr/testify/assert"
)

var (
	ctx   = context.Background()
	input = &provider.GSNUnresolvedConstructorInput{
		Config: &provider.GSNConfig{
			PaymasterAddress:        "0xCB798E55717978669Bf3F9A29cC117c7ea37a4A2",
			PreferredRelays:         []string{"https://relay.alg2-test.algen.network"},
			MinMaxPriorityFeePerGas: big.NewInt(1000252),
		},
		RelayParams: &provider.RelayParams{
			NodeUrl:    "https://rpc.alg2-test.algen.network",
			PrivateKey: "154279322921b6d9101e9c5178a31976504c24851c89d4f071567d3316c54c29",
			WalletType: 0,
		},
	}
)

var (
	param = &provider.Payload{
		From: "0xd3cff20db7af99bf6caf0a1247afe5bc63664172",
		To:   "0xa6c1a5b6b52fb299edd8e9cf02b63c980176bfe3",
		Gas:  "0xa84e",
		Data: "0xa9059cbb000000000000000000000000693438ff63fedd1559870fbe2b1ad4128b74468e000000000000000000000000000000000000000000000000a6f3254327f30000",
	}
)

func Test_SendTransaction(t *testing.T) {

	relayProvider := provider.NewRelayProvider(ctx, input)
	err := relayProvider.Init()
	assert.Nil(t, err)
	if err != nil {
		fmt.Println(utils.StructToJsonString(err))
	}

	realyingResult, err := relayProvider.SendTransaction(param)
	assert.Nil(t, err)
	fmt.Println(utils.StructToJsonString(realyingResult))
}

func Test_Parse(t *testing.T) {

	utils.Deraw("")

}
