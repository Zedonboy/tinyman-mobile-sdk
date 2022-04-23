package v1

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math"
	"strconv"
	"strings"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

type Pool struct {
	validator_app_id                int32
	asset_a                         *Asset
	asset_b                         *Asset
	client                          *TinymanClient
	asset1_reserves                 uint64
	asset2_reserves                 uint64
	address                         string
	exists                          bool
	liquidityAsset                  *Asset
	issuedLiquidity                 int64
	unclaimedProtocolFees           int64
	outstandingAsset1Amount         uint64
	outstandingAsset2Amount         uint64
	outstandingLiquidityAssetAmount uint64
	algoBalance                     uint64
	round                           uint64
}

var contract map[string]interface{}
var pools_json map[string]interface{}
var poolCache map[string]string

func init() {
	data, err := ioutil.ReadFile("asc.json")

	if err == nil {
		json.Unmarshal(data, &contract)
	}

	_poolJson, err := ioutil.ReadFile("pool.json")

	if err == nil {
		json.Unmarshal(_poolJson, &pools_json)
	}

	poolCache = make(map[string]string)
	for _, v := range pools_json["pools"].([]interface{}) {
		key := v.(map[string]interface{})["key"].(string)
		addr := v.(map[string]interface{})["address"].(string)
		poolCache[key] = addr
	}

}

func ShowContract() map[string]interface{} {
	return contract
}

func updatePool(pool *Pool, account models.Account) {

	vappid := account.AppsLocalState[0].Id
	if vappid == 0 {
		panic("Validator id is null")
	}

	stateArr := account.AppsLocalState[0].KeyValue

	dict := TealValueArrayToMap(stateArr)
	asset_1_id := ExtractInt(dict, "a1")
	asset_2_id := ExtractInt(dict, "a2")

	poolAddr := GetPoolAddress(int(vappid), int(asset_1_id), int(asset_2_id))

	asset1Reserves := ExtractInt(dict, "s1")
	asset2Reserves := ExtractInt(dict, "s2")
	issuedLiquidity := ExtractInt(dict, "ilt")
	unclaimedProtocolFees := ExtractInt(dict, "p")

	liquidityAssetId := account.CreatedAssets[0].Index

	exists := liquidityAssetId != 0
	liquidityAsset := FetchAsset(liquidityAssetId)
	outstandingAsset1Amount := ExtractInt(dict, IntToStateKey(int64(asset_1_id)))

	outstandingAsset2Amount := ExtractInt(dict, IntToStateKey(int64(asset_2_id)))

	outstandingLiquidityAssetAmount := ExtractInt(dict, IntToStateKey(int64(liquidityAssetId)))

	// pool := new(Pool)
	asset1Obj, err := pool.client.Fetch_asset(asset_1_id)
	if err != nil {
		asset1Obj = &Asset{id: asset_1_id}
	}

	asset2Obj, err := pool.client.Fetch_asset(asset_2_id)
	if err != nil {
		asset2Obj = &Asset{id: asset_2_id}
	}
	pool.address = poolAddr
	pool.exists = exists
	pool.liquidityAsset = liquidityAsset
	pool.asset1_reserves = asset1Reserves
	pool.asset2_reserves = asset2Reserves
	pool.issuedLiquidity = int64(issuedLiquidity)
	pool.unclaimedProtocolFees = int64(unclaimedProtocolFees)
	pool.outstandingAsset1Amount = outstandingAsset1Amount
	pool.outstandingAsset2Amount = outstandingAsset2Amount
	pool.outstandingLiquidityAssetAmount = outstandingLiquidityAssetAmount
	pool.validator_app_id = int32(vappid)
	pool.algoBalance = account.Amount
	pool.round = account.Round
	pool.asset_a = asset1Obj
	pool.asset_b = asset2Obj
}

func CreatePoolFromAccountInfo(account models.Account, asset_a Asset, asset_b Asset) (p *Pool, err error) {
	pool := new(Pool)
	updatePool(pool, account)

	if asset_a.id != p.asset_a.id {
		err = &LibraryError{"Asset 1 is not equal to founded pool address"}
		p = pool
		return
	}

	if asset_b.id != p.asset_b.id {
		err = &LibraryError{"Asset 2 is not equal to founded pool address"}
		p = pool
		return
	}

	p = pool
	return
}

func GetPoolAddress(validator_id int, asset_id_1 int, asset_id_2 int) string {
	as1 := max(asset_id_1, asset_id_2)
	as2 := min(asset_id_1, asset_id_2)

	key := strconv.Itoa(validator_id) + "-" + strconv.Itoa(as1) + "-" + strconv.Itoa(as2)

	elem, ok := poolCache[key]
	if ok {
		return elem
	}

	poolAddr := getLogicContractSignature(validator_id, as1, as2)
	poolCache[key] = poolAddr
	print(poolCache)
	return poolAddr
}

func getLogicContractSignature(v_app_id int, asset_1 int, asset_2 int) string {
	logic := contract["contracts"].(map[string]interface{})["pool_logicsig"].(map[string]interface{})["logic"].(map[string]interface{})
	programBytes := getProgram(logic, map[string]interface{}{
		"validator_app_id": v_app_id,
		"asset_id_1":       asset_1,
		"asset_id_2":       asset_1,
	})

	logic_prefix := []byte("Program")
	checkSum := checkSum(append(logic_prefix, programBytes...))
	_poolAdress, _ := encode_address(checkSum)

	return _poolAdress
}

func MakePool(client *TinymanClient, asset1 Asset, asset2 Asset) (p Pool, e error) {

	pool := &Pool{}
	addr := GetPoolAddress(int(client.validator_app_id), int(asset1.id), int(asset2.id))
	pool.address = addr
	pool.asset_a = &asset1
	pool.asset_b = &asset2
	p = *pool
	return
}

func (p *Pool) refresh() error {
	accInfoAction := p.client.algod_client.AccountInformation(p.address)
	acc, err := accInfoAction.Do(context.Background())
	if err != nil {
		return err
	}
	updatePool(p, acc)
	return nil
}

func getProgram(logic map[string]interface{}, dict map[string]interface{}) []byte {
	base64Str := logic["bytecode"].(string)
	variables := logic["variables"].([]interface{})
	templateBytes, _ := base64.StdEncoding.DecodeString(base64Str)
	var offset = 0

	for _, v := range variables {
		value := dict["asset_id_1"].(int)
		start := int(v.(map[string]interface{})["index"].(float64)) - offset
		end := start + int(v.(map[string]interface{})["length"].(float64))
		valueEncoded, _ := encodeValue(value, v.(map[string]interface{})["type"].(string))
		diff := int(v.(map[string]interface{})["length"].(float64)) - len(valueEncoded)
		offset += diff

		// updateList(start, end, templateBytes, valueEncoded)
		head := templateBytes[:start-1]
		tail := templateBytes[end-1:]
		templateBytes = append(head, valueEncoded...)
		templateBytes = append(templateBytes, tail...)
	}

	return templateBytes
}

func updateList(start int, end int, in []byte, out []byte) {
	count := 0
	for i := start; i <= end; i++ {
		if count >= len(out) {
			break
		}
		in[i] = out[count]
		count++
	}
}

func encodeValue(value int, _type string) (data []byte, e error) {
	if strings.ToLower(_type) != "int" {
		e = &LibraryError{message: "Unsupported"}
		return
	}

	result := make([]byte, 0)
	number := uint64(value)
	for {
		towrite := number & 0x7F

		number >>= 7
		if number > 0 {
			result = append(result, byte(towrite|0x80))
		} else {
			result = append(result, byte(towrite))
			break
		}
	}

	data = result
	return
}

func max(a int, b int) int {
	if a >= b {
		return a
	}

	return b
}

func min(a int, b int) int {
	if a <= b {
		return a
	}

	return b
}

func (p *Pool) CalculateMintQuote(asset_amount AssetAmount, slippage float32) (ms *MintQuote, err error) {
	accInfoAction := p.client.algod_client.AccountInformation(p.address)
	acc, err := accInfoAction.Do(context.Background())
	updatePool(p, acc)

	if slippage == 0 {
		slippage = 0.05
	}

	var amount1 *AssetAmount
	var amount2 *AssetAmount

	if asset_amount.asset == p.asset_a {
		amount1 = &asset_amount
		amount2 = nil
	} else {
		amount2 = &asset_amount
		amount1 = nil
	}

	if p.exists {
		err = &LibraryError{message: "Pool is not Boostrapped yet"}
		return
	}

	var liquidityAssetAmount uint64

	if p.issuedLiquidity > 0 {
		if amount1 == nil {
			amount1 = p.convert(*amount2)
		} else if amount2 == nil {
			amount2 = p.convert(*amount1)
		}
		min_a := (amount1.amount * uint64(p.issuedLiquidity)) / uint64(p.asset1_reserves)
		min_b := amount2.amount * uint64(p.issuedLiquidity) / uint64(p.asset2_reserves)
		liquidityAssetAmount = uint64(min(int(min_a), int(min_b)))
	} else {
		if amount1 == nil || amount2 == nil {
			err = &LibraryError{"Amounts required for both assets for first mint!"}
			return
		}

		liquidityAssetAmount = uint64(math.Sqrt((float64(amount1.amount) * float64(amount2.amount))) - 1000)
	}

	return &MintQuote{
		pool:                 p,
		asset_amt_1:          amount1,
		asset_amt_2:          amount2,
		liquidityAssetAmount: &AssetAmount{asset: p.liquidityAsset, amount: liquidityAssetAmount},
		slippage:             slippage,
	}, nil
}

func (p *Pool) CalculateBurnQuote(amtIn AssetAmount, slippage float32) (bq *BurnQuote, err error) {
	if slippage == 0 {
		slippage = 0.05
	}
	accInfoAction := p.client.algod_client.AccountInformation(p.address)
	acc, err := accInfoAction.Do(context.Background())
	updatePool(p, acc)
	if p.liquidityAsset.id != amtIn.asset.id {
		err = &LibraryError{"Expected ## to be liquidity pool asset amount."}
		return
	}

	asset1Amount := (amtIn.amount * p.asset1_reserves) / uint64(p.issuedLiquidity)
	asset2Amount := (amtIn.amount * p.asset2_reserves) / uint64(p.issuedLiquidity)

	bq = &BurnQuote{
		asset_amount_1_out: &AssetAmount{
			amount: asset1Amount,
			asset:  p.asset_a,
		},
		asset_amount_2_out: &AssetAmount{
			asset:  p.asset_b,
			amount: asset2Amount,
		},

		liquidityAssetAmount: &amtIn,
		slippage:             slippage,
	}
	return

}

func (p *Pool) convert(asset AssetAmount) *AssetAmount {
	if asset.asset == p.asset_a {
		am := new(AssetAmount)
		am.asset = p.asset_b
		am.amount = asset.amount * (p.asset2_reserves / p.asset1_reserves)
		return am
	}

	if asset.asset == p.asset_b {
		am := new(AssetAmount)
		am.asset = p.asset_a
		am.amount = asset.amount * (p.asset1_reserves / p.asset2_reserves)
		return am
	}

	return nil
}

func (p *Pool) Fetch_fixed_input_swap_quote(asset_amount AssetAmount, slippage float32) (qoute SwapQuote, err error) {
	p.refresh()
	var input_supply, output_supply uint32
	var asset_out *Asset
	if p.asset_a.id == asset_amount.asset.id {
		asset_out = p.asset_b
		input_supply = uint32(p.asset1_reserves)
		output_supply = uint32(p.asset2_reserves)
	} else {
		asset_out = p.asset_a
		input_supply = uint32(p.asset2_reserves)
		output_supply = uint32(p.asset1_reserves)
	}

	if input_supply == 0 || output_supply == 0 {
		err = &LibraryError{"Pool has no liquidity"}
		return
	}

	k := input_supply * output_supply
	asset_in_amount_minus_fee := (asset_amount.amount * 997) / 1000
	swap_fees := asset_amount.amount - asset_in_amount_minus_fee
	asset_out_amount := uint64(output_supply) - (uint64(k) / (uint64(input_supply) + asset_in_amount_minus_fee))

	asset_amount_out := AssetAmount{asset_out, asset_out_amount}
	if slippage == 0 {
		slippage = 0.05
	}
	qoute = SwapQuote{
		swap_type:       FIXED_INPUT,
		amount_in:       asset_amount,
		amount_out:      asset_amount_out,
		slippage:        slippage,
		swap_fees_asset: AssetAmount{asset_amount.asset, swap_fees},
	}

	return
}
