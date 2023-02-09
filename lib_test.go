//go:build cgo

package cosmwasm

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CosmWasm/wasmvm/internal/api"
	"github.com/CosmWasm/wasmvm/types"
)

const (
	TESTING_FEATURES     = "staking,stargate,iterator"
	TESTING_PRINT_DEBUG  = false
	TESTING_GAS_LIMIT    = uint64(500_000_000_000) // ~0.5ms
	TESTING_MEMORY_LIMIT = 32                      // MiB
	TESTING_CACHE_SIZE   = 100                     // MiB
)

const (
	CYBERPUNK_TEST_CONTRACT = "./testdata/cyberpunk.wasm"
	HACKATOM_TEST_CONTRACT  = "./testdata/hackatom.wasm"
	ITERATIONS              = 10000
)

func withVM(b *testing.B) *VM {
	tmpdir, err := ioutil.TempDir("", "wasmvm-testing")
	require.NoError(b, err)
	vm, err := NewVM(tmpdir, TESTING_FEATURES, TESTING_MEMORY_LIMIT, TESTING_PRINT_DEBUG, TESTING_CACHE_SIZE)
	require.NoError(b, err)

	b.Cleanup(func() {
		vm.Cleanup()
		os.RemoveAll(tmpdir)
	})
	return vm
}

func createTestContract(b *testing.B, vm *VM, path string) Checksum {
	wasm, err := ioutil.ReadFile(path)
	require.NoError(b, err)
	checksum, err := vm.StoreCode(wasm)
	require.NoError(b, err)
	return checksum
}

func BenchmarkHappyPathOneVM(b *testing.B) {
	vm := withVM(b)
	checksum := createTestContract(b, vm, HACKATOM_TEST_CONTRACT)

	deserCost := types.UFraction{1, 1}
	gasMeter1 := api.NewMockGasMeter(TESTING_GAS_LIMIT)
	// instantiate it with this store
	store := api.NewLookup(gasMeter1)
	goapi := api.NewMockAPI()
	balance := types.Coins{types.NewCoin(250, "ATOM")}
	querier := api.DefaultQuerier(api.MOCK_CONTRACT_ADDR, balance)

	// instantiate
	env := api.MockEnv()
	info := api.MockInfo("creator", nil)
	msg := []byte(`{"verifier": "fred", "beneficiary": "bob"}`)
	ires, _, err := vm.Instantiate(checksum, env, info, msg, store, *goapi, querier, gasMeter1, TESTING_GAS_LIMIT, deserCost)
	require.NoError(b, err)
	require.Equal(b, 0, len(ires.Messages))

	for i := 0; i < ITERATIONS; i++ {
		// execute
		gasMeter2 := api.NewMockGasMeter(TESTING_GAS_LIMIT)
		store.SetGasMeter(gasMeter2)
		env = api.MockEnv()
		info = api.MockInfo("fred", nil)
		hres, _, err := vm.Execute(checksum, env, info, []byte(`{"release":{}}`), store, *goapi, querier, gasMeter2, TESTING_GAS_LIMIT, deserCost)
		require.NoError(b, err)
		require.Equal(b, 1, len(hres.Messages))
		hres, _, err = vm.Execute(checksum, env, info, []byte(`{"release":{}}`), store, *goapi, querier, gasMeter2, TESTING_GAS_LIMIT, deserCost)
		require.NoError(b, err)
		require.Equal(b, 1, len(hres.Messages))
	}
}

func BenchmarkHappyPathTwoVMs(b *testing.B) {
	vm1 := withVM(b)
	vm2 := withVM(b)
	checksum := createTestContract(b, vm1, HACKATOM_TEST_CONTRACT)
	checksum2 := createTestContract(b, vm2, HACKATOM_TEST_CONTRACT)

	deserCost := types.UFraction{1, 1}
	gasMeter1 := api.NewMockGasMeter(TESTING_GAS_LIMIT)
	// instantiate it with this store
	store := api.NewLookup(gasMeter1)
	goapi := api.NewMockAPI()
	balance := types.Coins{types.NewCoin(250, "ATOM")}
	querier := api.DefaultQuerier(api.MOCK_CONTRACT_ADDR, balance)

	// instantiate
	env := api.MockEnv()
	info := api.MockInfo("creator", nil)
	msg := []byte(`{"verifier": "fred", "beneficiary": "bob"}`)
	ires1, _, err := vm1.Instantiate(checksum, env, info, msg, store, *goapi, querier, gasMeter1, TESTING_GAS_LIMIT, deserCost)
	require.NoError(b, err)
	require.Equal(b, 0, len(ires1.Messages))
	ires2, _, err := vm2.Instantiate(checksum2, env, info, msg, store, *goapi, querier, gasMeter1, TESTING_GAS_LIMIT, deserCost)
	require.NoError(b, err)
	require.Equal(b, 0, len(ires2.Messages))

	for i := 0; i < ITERATIONS; i++ {
		// execute
		gasMeter2 := api.NewMockGasMeter(TESTING_GAS_LIMIT)
		store.SetGasMeter(gasMeter2)
		env = api.MockEnv()
		info = api.MockInfo("fred", nil)
		hres, _, err := vm1.Execute(checksum, env, info, []byte(`{"release":{}}`), store, *goapi, querier, gasMeter2, TESTING_GAS_LIMIT, deserCost)
		require.NoError(b, err)
		require.Equal(b, 1, len(hres.Messages))
		hres, _, err = vm2.Execute(checksum, env, info, []byte(`{"release":{}}`), store, *goapi, querier, gasMeter2, TESTING_GAS_LIMIT, deserCost)
		require.NoError(b, err)
		require.Equal(b, 1, len(hres.Messages))
	}
}
