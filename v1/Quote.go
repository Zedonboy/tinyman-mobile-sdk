package v1

const (
	FIXED_OUTPUT = "fixed-output"
	FIXED_INPUT  = "fixed-input"
)

type SwapQuote struct {
	swap_type       string
	amount_in       AssetAmount
	amount_out      AssetAmount
	swap_fees_asset AssetAmount
	slippage        float32
}

type BurnQuote struct {
	asset_amount_1_out   *AssetAmount
	asset_amount_2_out   *AssetAmount
	slippage             float32
	liquidityAssetAmount *AssetAmount
}

type RedeemQuote struct {
	amount      *AssetAmount
	poolAddress string
	pool        *Pool
}

type MintQuote struct {
	pool                 *Pool
	slippage             float32
	asset_amt_1          *AssetAmount
	asset_amt_2          *AssetAmount
	liquidityAssetAmount *AssetAmount
}

func (mq *MintQuote) GetLiquidityAssetAmountWithSlippage(rhs *AssetAmount) AssetAmount {

	amt := float32(rhs.amount) * float32(mq.slippage)
	n_amt := mq.slippage - amt
	return AssetAmount{amount: uint64(n_amt)}
}
