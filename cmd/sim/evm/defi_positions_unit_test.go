package evm

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// apiResponseJSON is the exact JSON returned by the defi-positions API endpoint,
// used to verify that our structs can unmarshal every field correctly.
const apiResponseJSON = `{"positions":[{"type":"Erc4626","chain":"ethereum","chain_id":1,"token":{"address":"0xa3931d71877c0e7a3148cb7eb4463524fec27fbd","name":"Savings USDS","symbol":"sUSDS"},"underlying_token":{"address":"0xdc035d45d973e3ec169d2276ddab16f1e407384f","name":"USDS Stablecoin","symbol":"USDS","decimals":18,"holdings":47.00372505463423},"balance":43.091714709517426,"price_usd":1.0906557333153313,"value_usd":46.99822570632377,"logo":"https://api.sim.dune.com/beta/token/logo/1/0xdc035d45d973e3ec169d2276ddab16f1e407384f"},{"type":"Tokenized","chain":"ethereum","chain_id":1,"token_type":"AtokenV2","token":{"address":"0x030ba81f1c18d280636f32af80b9aad02cf0854e","name":"Aave interest bearing WETH","symbol":"aWETH"},"underlying_token":{"address":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","holdings":0.0515820177322061},"lending_pool":"0x7d2768de32b0b80b7a3454c06bdac94a69ddc7a9","balance":0.04961716041594229,"price_usd":2257.7211358982086,"value_usd":112.02171177432484,"logo":"https://api.sim.dune.com/beta/token/logo/1/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"},{"type":"UniswapV2","chain":"ethereum","chain_id":1,"protocol":"ShibaSwapV2","pool":"0x76ec974feaf0293f64cf8643e0f42dea5b71689b","token0":{"address":"0x198065e69a86cb8a9154b333aad8efe7a3c256f8","name":"KOYO","symbol":"KOY","decimals":18,"price_usd":0.00009108374058938265,"holdings":72267.17195098454},"token1":{"address":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","name":"Wrapped Ether","symbol":"WETH","decimals":18,"price_usd":2171.720237,"holdings":0.00297532741832308},"lp_balance":"0x8ac7230489e80000","balance":10.0,"price_usd":1.3043943109184983,"value_usd":13.043943109184983,"logo":"https://api.sim.dune.com/beta/token/logo/1/0x198065e69a86cb8a9154b333aad8efe7a3c256f8"},{"type":"UniswapV2","chain":"ethereum","chain_id":1,"protocol":"UniswapV2","pool":"0x09c29277d081a1b347f41277ff53116a30d4ddff","token0":{"address":"0x4206975c6d7135ad73129476ebe2b06e42f41f50","name":"FWOG","symbol":"FWOG","decimals":18,"price_usd":2.3754275952306002e-11,"holdings":609179876998.8339},"token1":{"address":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","name":"Wrapped Ether","symbol":"WETH","decimals":18,"price_usd":2171.720237,"holdings":0.006683389542586471},"lp_balance":"0xca7455529bd53680000","balance":59754.0,"price_usd":0.00048507345490195365,"value_usd":28.98507922421134,"logo":null},{"type":"Nft","chain":"ethereum","chain_id":1,"protocol":"UniswapV3","pool":"0x7625d7f67e4e44341ddfb1e698801fd5a1574b48","token0":{"address":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","name":"Wrapped Ether","symbol":"WETH","decimals":18},"token1":{"address":"0xd78959df1ff28b45e6b4ea234bdcf9d6609d16e1","name":"Moneda de Caca","symbol":"Mierda","decimals":18},"positions":[{"tick_lower":0,"tick_upper":184200,"token_id":"0xe8ac0","token0":{"price_usd":2171.720237,"holdings":0.0,"rewards":0.000201014351146556},"token1":{"price_usd":0.0,"holdings":100000000.0,"rewards":19678.323702556045}}],"logo":"https://api.sim.dune.com/beta/token/logo/1/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","value_usd":0.43654693431239977},{"type":"NftV4","chain":"ethereum","chain_id":1,"protocol":"UniswapV4","pool_id":"0x21fb293b9dc53b42fa6e63fa24e1212de76c88eb7a15b94cd220fc66274851bf","pool_manager":"0x000000000004444c5dc75cb358380d2e3de08a90","salt":"0x0000000000000000000000000000000000000000000000000000000000002118","token0":{"address":"0x0000000000000000000000000000000000000000","name":"Ether","symbol":"ETH","decimals":18},"token1":{"address":"0xf9c8631fba291bac14ed549a2dde7c7f2ddff1a8","name":"Mighty Morphin Power Rangers","symbol":"GoGo","decimals":18},"positions":[{"tick_lower":-184220,"tick_upper":207220,"token_id":"0x2118","token0":{"price_usd":2171.720237,"holdings":0.000252141460675072,"rewards":8.852246853764e-6},"token1":{"price_usd":0.0,"holdings":479748570.0393271,"rewards":8399.799985973925}}],"logo":"https://api.sim.dune.com/beta/token/logo/1/0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","value_usd":0.5668053163700324}],"aggregations":{"total_value_usd":6153.428535761592,"total_by_chain":{"1":4206.858757536072,"8453":1946.5539365060883,"42161":0.015841719431772122}}}`

func TestUnmarshal_FullAPIResponse(t *testing.T) {
	var resp defiPositionsResponse
	err := json.Unmarshal([]byte(apiResponseJSON), &resp)
	require.NoError(t, err, "unmarshal must not fail on real API response")

	require.Len(t, resp.Positions, 6)

	// --- Erc4626 ---
	erc := resp.Positions[0]
	assert.Equal(t, "Erc4626", erc.Type)
	assert.Equal(t, "ethereum", erc.Chain)
	assert.Equal(t, int64(1), erc.ChainID)
	assert.InDelta(t, 46.998, erc.ValueUSD, 0.01)
	assert.InDelta(t, 43.0917, erc.Balance, 0.001)
	assert.InDelta(t, 1.0906, erc.PriceUSD, 0.001)
	require.NotNil(t, erc.Logo)
	assert.Contains(t, *erc.Logo, "api.sim.dune.com")
	// token
	require.NotNil(t, erc.Token)
	assert.Equal(t, "0xa3931d71877c0e7a3148cb7eb4463524fec27fbd", erc.Token.Address)
	assert.Equal(t, "Savings USDS", erc.Token.Name)
	assert.Equal(t, "sUSDS", erc.Token.Symbol)
	// underlying_token
	require.NotNil(t, erc.UnderlyingToken)
	assert.Equal(t, "0xdc035d45d973e3ec169d2276ddab16f1e407384f", erc.UnderlyingToken.Address)
	assert.Equal(t, "USDS Stablecoin", erc.UnderlyingToken.Name)
	assert.Equal(t, "USDS", erc.UnderlyingToken.Symbol)
	assert.Equal(t, 18, erc.UnderlyingToken.Decimals)
	assert.InDelta(t, 47.003, erc.UnderlyingToken.Holdings, 0.001)

	// --- Tokenized ---
	tok := resp.Positions[1]
	assert.Equal(t, "Tokenized", tok.Type)
	assert.Equal(t, "AtokenV2", tok.TokenType)
	assert.Equal(t, "0x7d2768de32b0b80b7a3454c06bdac94a69ddc7a9", tok.LendingPool)
	assert.InDelta(t, 0.04961, tok.Balance, 0.0001)
	assert.InDelta(t, 2257.72, tok.PriceUSD, 0.01)
	assert.InDelta(t, 112.02, tok.ValueUSD, 0.01)
	require.NotNil(t, tok.Token)
	assert.Equal(t, "aWETH", tok.Token.Symbol)
	assert.Equal(t, "Aave interest bearing WETH", tok.Token.Name)
	require.NotNil(t, tok.UnderlyingToken)
	assert.Equal(t, "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2", tok.UnderlyingToken.Address)
	assert.InDelta(t, 0.05158, tok.UnderlyingToken.Holdings, 0.0001)

	// --- UniswapV2 (with logo) ---
	uni := resp.Positions[2]
	assert.Equal(t, "UniswapV2", uni.Type)
	assert.Equal(t, "ShibaSwapV2", uni.Protocol)
	assert.Equal(t, "0x76ec974feaf0293f64cf8643e0f42dea5b71689b", uni.Pool)
	assert.Equal(t, "0x8ac7230489e80000", uni.LPBalance)
	assert.InDelta(t, 10.0, uni.Balance, 0.001)
	require.NotNil(t, uni.Token0)
	assert.Equal(t, "KOY", uni.Token0.Symbol)
	assert.Equal(t, 18, uni.Token0.Decimals)
	assert.InDelta(t, 0.0000911, uni.Token0.PriceUSD, 0.00001)
	assert.InDelta(t, 72267.17, uni.Token0.Holdings, 0.01)
	require.NotNil(t, uni.Token1)
	assert.Equal(t, "WETH", uni.Token1.Symbol)
	assert.InDelta(t, 2171.72, uni.Token1.PriceUSD, 0.01)

	// --- UniswapV2 (logo: null) ---
	uniNull := resp.Positions[3]
	assert.Equal(t, "UniswapV2", uniNull.Type)
	assert.Nil(t, uniNull.Logo, "null logo should be nil pointer")
	assert.InDelta(t, 59754.0, uniNull.Balance, 0.1)

	// --- Nft (UniswapV3) ---
	nft := resp.Positions[4]
	assert.Equal(t, "Nft", nft.Type)
	assert.Equal(t, "UniswapV3", nft.Protocol)
	assert.Equal(t, "0x7625d7f67e4e44341ddfb1e698801fd5a1574b48", nft.Pool)
	require.NotNil(t, nft.Token0)
	assert.Equal(t, "WETH", nft.Token0.Symbol)
	require.NotNil(t, nft.Token1)
	assert.Equal(t, "Mierda", nft.Token1.Symbol)
	assert.InDelta(t, 0.4365, nft.ValueUSD, 0.001)
	// NFT position details
	require.Len(t, nft.Positions, 1)
	nftPos := nft.Positions[0]
	assert.Equal(t, 0, nftPos.TickLower)
	assert.Equal(t, 184200, nftPos.TickUpper)
	assert.Equal(t, "0xe8ac0", nftPos.TokenID)
	require.NotNil(t, nftPos.Token0)
	assert.InDelta(t, 2171.72, nftPos.Token0.PriceUSD, 0.01)
	assert.InDelta(t, 0.0, nftPos.Token0.Holdings, 0.001)
	assert.InDelta(t, 0.000201, nftPos.Token0.Rewards, 0.00001)
	require.NotNil(t, nftPos.Token1)
	assert.InDelta(t, 0.0, nftPos.Token1.PriceUSD, 0.001)
	assert.InDelta(t, 100000000.0, nftPos.Token1.Holdings, 1.0)
	assert.InDelta(t, 19678.32, nftPos.Token1.Rewards, 0.01)

	// --- NftV4 (UniswapV4) ---
	nft4 := resp.Positions[5]
	assert.Equal(t, "NftV4", nft4.Type)
	assert.Equal(t, "UniswapV4", nft4.Protocol)
	assert.Equal(t, "0x21fb293b9dc53b42fa6e63fa24e1212de76c88eb7a15b94cd220fc66274851bf", nft4.PoolID)
	assert.Equal(t, "0x000000000004444c5dc75cb358380d2e3de08a90", nft4.PoolManager)
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000002118", nft4.Salt)
	require.NotNil(t, nft4.Token0)
	assert.Equal(t, "ETH", nft4.Token0.Symbol)
	assert.Equal(t, "0x0000000000000000000000000000000000000000", nft4.Token0.Address)
	require.NotNil(t, nft4.Token1)
	assert.Equal(t, "GoGo", nft4.Token1.Symbol)
	assert.InDelta(t, 0.5668, nft4.ValueUSD, 0.001)
	require.Len(t, nft4.Positions, 1)
	nft4Pos := nft4.Positions[0]
	assert.Equal(t, -184220, nft4Pos.TickLower)
	assert.Equal(t, 207220, nft4Pos.TickUpper)
	assert.Equal(t, "0x2118", nft4Pos.TokenID)
	require.NotNil(t, nft4Pos.Token0)
	assert.InDelta(t, 0.000252, nft4Pos.Token0.Holdings, 0.00001)
	assert.InDelta(t, 8.852e-6, nft4Pos.Token0.Rewards, 1e-7)

	// --- Aggregations ---
	require.NotNil(t, resp.Aggregations)
	assert.InDelta(t, 6153.43, resp.Aggregations.TotalValueUSD, 0.01)
	assert.Len(t, resp.Aggregations.TotalByChain, 3)
	assert.InDelta(t, 4206.86, resp.Aggregations.TotalByChain["1"], 0.01)
	assert.InDelta(t, 1946.55, resp.Aggregations.TotalByChain["8453"], 0.01)
	assert.InDelta(t, 0.01584, resp.Aggregations.TotalByChain["42161"], 0.0001)
}

func TestPositionDetails_Erc4626(t *testing.T) {
	p := defiPosition{
		Type:            "Erc4626",
		Token:           &defiTokenInfo{Symbol: "alUSD"},
		UnderlyingToken: &defiTokenInfo{Symbol: "USDC"},
		Balance:         0.0673736869415349,
	}
	got := positionDetails(p)
	assert.Contains(t, got, "alUSD")
	assert.Contains(t, got, "-> USDC")
	assert.Contains(t, got, "bal=")
}

func TestPositionDetails_Erc4626_NoBalance(t *testing.T) {
	p := defiPosition{
		Type:            "Erc4626",
		Token:           &defiTokenInfo{Symbol: "yvDAI"},
		UnderlyingToken: &defiTokenInfo{Symbol: "DAI"},
	}
	assert.Equal(t, "yvDAI -> DAI", positionDetails(p))
}

func TestPositionDetails_Tokenized(t *testing.T) {
	p := defiPosition{
		Type:      "Tokenized",
		TokenType: "AtokenV2",
		Token:     &defiTokenInfo{Symbol: "aWETH"},
		Balance:   0.0496171604159423,
	}
	got := positionDetails(p)
	assert.Contains(t, got, "AtokenV2")
	assert.Contains(t, got, "aWETH")
	assert.Contains(t, got, "bal=")
}

func TestPositionDetails_Tokenized_NoBalance(t *testing.T) {
	p := defiPosition{
		Type:      "Tokenized",
		TokenType: "AtokenV2",
		Token:     &defiTokenInfo{Symbol: "aWBTC"},
	}
	assert.Equal(t, "AtokenV2 aWBTC", positionDetails(p))
}

func TestPositionDetails_UniswapV2(t *testing.T) {
	p := defiPosition{
		Type:    "UniswapV2",
		Token0:  &defiTokenInfo{Symbol: "FWOG"},
		Token1:  &defiTokenInfo{Symbol: "WETH"},
		Balance: 59754,
	}
	got := positionDetails(p)
	assert.Contains(t, got, "FWOG/WETH")
	assert.Contains(t, got, "bal=59754")
}

func TestPositionDetails_UniswapV2_NoBalance(t *testing.T) {
	p := defiPosition{
		Type:   "UniswapV2",
		Token0: &defiTokenInfo{Symbol: "USDC"},
		Token1: &defiTokenInfo{Symbol: "WETH"},
	}
	assert.Equal(t, "USDC/WETH", positionDetails(p))
}

func TestPositionDetails_Nft(t *testing.T) {
	p := defiPosition{
		Type:   "Nft",
		Token0: &defiTokenInfo{Symbol: "WETH"},
		Token1: &defiTokenInfo{Symbol: "USDC"},
		Positions: []nftPositionDetails{
			{TickLower: -100, TickUpper: 100, TokenID: "0x1"},
			{TickLower: -200, TickUpper: 200, TokenID: "0x2"},
			{TickLower: -300, TickUpper: 300, TokenID: "0x3"},
		},
	}
	assert.Equal(t, "WETH/USDC (3 positions)", positionDetails(p))
}

func TestPositionDetails_NftV4(t *testing.T) {
	p := defiPosition{
		Type:   "NftV4",
		Token0: &defiTokenInfo{Symbol: "WBTC"},
		Token1: &defiTokenInfo{Symbol: "WETH"},
		Positions: []nftPositionDetails{
			{TickLower: -50, TickUpper: 50, TokenID: "0xabc"},
		},
	}
	assert.Equal(t, "WBTC/WETH (1 position)", positionDetails(p))
}

func TestPositionDetails_NftNoPositions(t *testing.T) {
	p := defiPosition{
		Type:   "Nft",
		Token0: &defiTokenInfo{Symbol: "DAI"},
		Token1: &defiTokenInfo{Symbol: "USDC"},
	}
	assert.Equal(t, "DAI/USDC", positionDetails(p))
}

func TestPositionDetails_Unknown(t *testing.T) {
	p := defiPosition{Type: "SomeNewType"}
	assert.Equal(t, "", positionDetails(p))
}

func TestFormatPair(t *testing.T) {
	assert.Equal(t, "WETH/USDC", formatPair("WETH", "USDC"))
	assert.Equal(t, "WETH", formatPair("WETH", ""))
	assert.Equal(t, "USDC", formatPair("", "USDC"))
	assert.Equal(t, "", formatPair("", ""))
}

func TestTokenSymbol_Nil(t *testing.T) {
	assert.Equal(t, "", tokenSymbol(nil))
}

func TestTokenSymbol_NonNil(t *testing.T) {
	assert.Equal(t, "ETH", tokenSymbol(&defiTokenInfo{Symbol: "ETH"}))
}
