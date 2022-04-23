package v1

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"strconv"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

type TinymanClient struct {
	algod_client     *algod.Client
	validator_app_id int32
	assets_cache     map[uint64]*Asset
	// user_address     string
}

func MakeTinyManClient(client *algod.Client, validator_app_id int32) (c *TinymanClient, err error) {
	c = &TinymanClient{
		algod_client:     client,
		validator_app_id: validator_app_id,
		assets_cache:     make(map[uint64]*Asset),
		// user_address:     user_address,
	}
	return
}

func (c *TinymanClient) Fetch_pool(asset_a Asset, asset_b Asset) (p *Pool, err error) {
	asset_id_1 := min(int(asset_a.id), int(asset_a.id))
	asset_id_2 := max(int(asset_a.id), int(asset_b.id))
	key := strconv.Itoa(asset_id_1) + "_" + strconv.Itoa(asset_id_2)
	poolAddr, ok := poolCache[key]
	if !ok {
		err = &LibraryError{"Pool does not exist, try using its adddress"}
		return
	}
	// poolAddr := GetPoolAddress(int(c.validator_app_id), int(asset_a.id), int(asset_b.id))
	accInfo := c.algod_client.AccountInformation(poolAddr)
	acc, err := accInfo.Do(context.Background())
	p, err = CreatePoolFromAccountInfo(acc, asset_a, asset_b)
	return
}

func (c *TinymanClient) Fetch_pool_by_address(address string) (p *Pool, err error) {
	accInfo := c.algod_client.AccountInformation(address)
	acc, err := accInfo.Do(context.Background())
	p, err = CreatePoolFromAccountInfo(acc, Asset{id: 0}, Asset{id: 0})
	err = nil
	return
}

func (c *TinymanClient) FetchExcessAmount(address string) []RedeemQuote {
	accInfo := c.algod_client.AccountInformation(address)
	acc, _ := accInfo.Do(context.Background())

	var vappAppState models.ApplicationLocalState

	for _, appStates := range acc.AppsLocalState {
		if appStates.Id == uint64(c.validator_app_id) {
			vappAppState = appStates
			break
		}
	}

	dict := TealValueArrayToMap(vappAppState.KeyValue)
	var results []RedeemQuote
	for k := range dict {
		utfBytes := []byte(k)
		var base64bytes []byte
		base64.StdEncoding.Decode(base64bytes, utfBytes)
		if base64bytes[len(base64bytes)-9] == byte('e') {
			value := dict[k].Uint
			poolAddress, _ := encode_address(base64bytes[:len(base64bytes)-9])
			var assetId uint64
			buf := new(bytes.Buffer)
			buf.Write(base64bytes[:(len(base64bytes) - 9)])
			binary.Read(buf, binary.BigEndian, &assetId)
			asset, _ := c.Fetch_asset(assetId)
			pool, _ := c.Fetch_pool_by_address(poolAddress)
			rq := RedeemQuote{
				pool: pool,
				amount: &AssetAmount{
					asset:  asset,
					amount: value,
				},
				poolAddress: poolAddress,
			}
			results = append(results, rq)
		}
	}

	return results
}

func (c *TinymanClient) Fetch_asset(id uint64) (a *Asset, err error) {
	elem, ok := c.assets_cache[id]
	if ok {
		return elem, nil
	} else {
		assetId := c.algod_client.GetAssetByID(uint64(id))
		asset, _err := assetId.Do(context.Background())
		err = _err
		a = new(Asset)
		a.id = asset.Index
		a.unit_name = asset.Params.UnitName
		a.name = asset.Params.Name
		a.decimals = uint32(asset.Params.Decimals)

		c.assets_cache[asset.Index] = a
	}

	return
}

func TestNetClient() *TinymanClient {
	algodClient, _ := algod.MakeClient("https://testnet-api.algonode.cloud", "")
	tinyClient, _ := MakeTinyManClient(algodClient, 62368684)

	return tinyClient
}

func MainNetClient() *TinymanClient {
	algodClient, _ := algod.MakeClient("https://mainnet-api.algonode.cloud", "")
	tinyClient, _ := MakeTinyManClient(algodClient, 552635992)

	return tinyClient
}
