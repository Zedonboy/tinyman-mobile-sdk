package v1

import (
	"testing"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
)

func TestClient(t *testing.T) {
	algodClient, err := algod.MakeClient("https://testnet-api.algonode.cloud", "")
	tinyClient, err := MakeTinyManClient(algodClient, 62368684)

	tinyusdc, err := tinyClient.Fetch_asset(21582668)
	algo, err := tinyClient.Fetch_asset(0)

	pool, err := tinyClient.Fetch_pool(*tinyusdc, *algo)
	if err != nil {
		t.Error("Pool is not cool")
	}
	t.Log(poolCache)
	t.Log(pool.address)
}
