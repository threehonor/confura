package test

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	sdk "github.com/Conflux-Chain/go-conflux-sdk"
	"github.com/Conflux-Chain/go-conflux-sdk/types"
	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	fullnode, infura sdk.ClientOperator

	wcfx     = cfxaddress.MustNewFromBase32("cfx:acg158kvr8zanb1bs048ryb6rtrhr283ma70vz70tx")
	wcfxTest = cfxaddress.MustNewFromBase32("cfxtest:achs3nehae0j6ksvy1bhrffsh1rtfrw1f6w1kzv46t")
)

// Please set the following environments to start:
// `TEST_CFX_FULL_NODE`: Core space full node endpoint as benchmarking data source.
// `TEST_CFX_INFURA_NODE`: Core space infura service endpoint to be validated against.

func TestMain(m *testing.M) {
	if err := setup(); err != nil {
		panic(errors.WithMessage(err, "failed to setup"))
	}

	code := 0
	if fullnode != nil && infura != nil {
		code = m.Run()
	}

	if err := teardown(); err != nil {
		panic(errors.WithMessage(err, "failed to tear down"))
	}

	os.Exit(code)
}

func setup() error {
	cfxHttpNode := os.Getenv("TEST_CFX_FULL_NODE")
	if len(cfxHttpNode) > 0 {
		fullnode = sdk.MustNewClient(cfxHttpNode)
	}

	cfxHttpNode = os.Getenv("TEST_CFX_INFURA_NODE")
	if len(cfxHttpNode) > 0 {
		infura = sdk.MustNewClient(cfxHttpNode)
	}

	return nil
}

func teardown() (err error) {
	if fullnode != nil {
		fullnode.Close()
	}

	if infura != nil {
		infura.Close()
	}

	return nil
}

func mustGetTestEpoch(t require.TestingT, epoch *types.Epoch, deltaToLatestState int64) *types.Epoch {
	num, err := fullnode.GetEpochNumber(epoch)
	require.NoError(t, err)
	targetEpoch := num.ToInt().Int64() + deltaToLatestState
	return types.NewEpochNumberBig(big.NewInt(targetEpoch))
}

func mustMarshalJSON(t require.TestingT, v interface{}) string {
	encoded, err := json.Marshal(v)
	require.NoError(t, err)
	return string(encoded)
}

func TestGetBalance(t *testing.T) {
	epoch := mustGetTestEpoch(t, types.EpochLatestConfirmed, 0)

	b1, err := fullnode.GetBalance(wcfx, types.NewEpochOrBlockHashWithEpoch(epoch))
	require.NoError(t, err)
	require.NotNil(t, b1)

	b2, err := infura.GetBalance(wcfx, types.NewEpochOrBlockHashWithEpoch(epoch))
	require.NoError(t, err)
	require.NotNil(t, b2)

	assert.Equal(t, b1.ToInt().Int64(), b2.ToInt().Int64())
}

func TestGetAdmin(t *testing.T) {
	admin1, err := fullnode.GetAdmin(wcfx)
	require.NoError(t, err)
	require.NotNil(t, admin1)

	admin2, err := infura.GetAdmin(wcfx)
	require.NoError(t, err)
	require.NotNil(t, admin2)

	assert.Equal(t, admin1.String(), admin2.String())
}

func testGetLogs(t *testing.T, expectedCount int, filter types.LogFilter) {
	logs1, err := fullnode.GetLogs(filter)
	require.NoError(t, err)
	assert.Equal(t, expectedCount, len(logs1))

	logs2, err := infura.GetLogs(filter)
	require.NoError(t, err)
	assert.Equal(t, expectedCount, len(logs2))

	json1 := mustMarshalJSON(t, logs1)
	json2 := mustMarshalJSON(t, logs2)
	assert.Equal(t, json1, json2)
}

func TestGetLogs(t *testing.T) {
	// filter: epoch range
	testGetLogs(t, 55, types.LogFilter{
		FromEpoch: types.NewEpochNumberUint64(10477303),
		ToEpoch:   types.NewEpochNumberUint64(10477315),
	})

	// filter: contract address
	testGetLogs(t, 22, types.LogFilter{
		FromEpoch: types.NewEpochNumberUint64(10477303),
		ToEpoch:   types.NewEpochNumberUint64(10477315),
		Address: []types.Address{
			cfxaddress.MustNewFromBase32("cfx:acckucyy5fhzknbxmeexwtaj3bxmeg25b2b50pta6v"),
			cfxaddress.MustNewFromBase32("cfx:acc7uawf5ubtnmezvhu9dhc6sghea0403y2dgpyfjp"),
		},
	})

	// filter: blocks
	testGetLogs(t, 17, types.LogFilter{
		BlockHashes: []types.Hash{
			types.Hash("0x9071c4446dfe9a8ce22175863c53b3b99bd596d89470423c5bb4a262c4a8716c"),
			types.Hash("0x3e722a9a61ada255c334d7fea10179b6ae6f084af293e1ef136a7b6f856edbcf"),
		},
	})

	// filter: contract address + blocks
	testGetLogs(t, 6, types.LogFilter{
		Address: []types.Address{
			cfxaddress.MustNewFromBase32("cfx:acckucyy5fhzknbxmeexwtaj3bxmeg25b2b50pta6v"),
			cfxaddress.MustNewFromBase32("cfx:acdrf821t59y12b4guyzckyuw2xf1gfpj2ba0x4sj6"),
		},
		BlockHashes: []types.Hash{
			types.Hash("0x9071c4446dfe9a8ce22175863c53b3b99bd596d89470423c5bb4a262c4a8716c"),
			types.Hash("0x3e722a9a61ada255c334d7fea10179b6ae6f084af293e1ef136a7b6f856edbcf"),
		},
	})

	// filter: topics
	testGetLogs(t, 16, types.LogFilter{
		FromEpoch: types.NewEpochNumberUint64(10477303),
		ToEpoch:   types.NewEpochNumberUint64(10477315),
		Topics: [][]types.Hash{
			{types.Hash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")},
			{types.Hash("0x0000000000000000000000001e61c5dab363c1fdb903b61178b380d2cc7df999")},
		},
	})
}

func TestErrorWithData(t *testing.T) {
	_, err1 := fullnode.GetBalance(wcfxTest)
	_, err2 := infura.GetBalance(wcfxTest)
	assert.Equal(t, mustMarshalJSON(t, err1), mustMarshalJSON(t, err2))
}
