package v1

type AssetAmount struct {
	asset  *Asset
	amount uint64
}
type Asset struct {
	id        uint64
	name      string
	unit_name string
	decimals  uint32
}

func FetchAsset(id uint64) *Asset {
	return new(Asset)
}

func (a *Asset) createAssetAmount(amt uint64) AssetAmount {
	return AssetAmount{a, amt}
}
