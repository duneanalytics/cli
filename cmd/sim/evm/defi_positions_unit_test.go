package evm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPositionDetails_Erc4626(t *testing.T) {
	p := defiPosition{
		Type:                  "Erc4626",
		TokenSymbol:           "alUSD",
		UnderlyingTokenSymbol: "USDC",
		CalculatedBalance:     0.0673736869415349,
	}
	got := positionDetails(p)
	assert.Contains(t, got, "alUSD")
	assert.Contains(t, got, "-> USDC")
	assert.Contains(t, got, "bal=")
}

func TestPositionDetails_Erc4626_NoBalance(t *testing.T) {
	p := defiPosition{
		Type:                  "Erc4626",
		TokenSymbol:           "yvDAI",
		UnderlyingTokenSymbol: "DAI",
	}
	assert.Equal(t, "yvDAI -> DAI", positionDetails(p))
}

func TestPositionDetails_Tokenized(t *testing.T) {
	p := defiPosition{
		Type:              "Tokenized",
		TokenType:         "AtokenV2",
		TokenSymbol:       "aWETH",
		CalculatedBalance: 0.0496171604159423,
	}
	got := positionDetails(p)
	assert.Contains(t, got, "AtokenV2")
	assert.Contains(t, got, "aWETH")
	assert.Contains(t, got, "bal=")
}

func TestPositionDetails_Tokenized_NoBalance(t *testing.T) {
	p := defiPosition{
		Type:        "Tokenized",
		TokenType:   "AtokenV2",
		TokenSymbol: "aWBTC",
	}
	assert.Equal(t, "AtokenV2 aWBTC", positionDetails(p))
}

func TestPositionDetails_UniswapV2(t *testing.T) {
	p := defiPosition{
		Type:              "UniswapV2",
		Token0Symbol:      "FWOG",
		Token1Symbol:      "WETH",
		CalculatedBalance: 59754,
	}
	got := positionDetails(p)
	assert.Contains(t, got, "FWOG/WETH")
	assert.Contains(t, got, "bal=59754")
}

func TestPositionDetails_UniswapV2_NoBalance(t *testing.T) {
	p := defiPosition{
		Type:         "UniswapV2",
		Token0Symbol: "USDC",
		Token1Symbol: "WETH",
	}
	assert.Equal(t, "USDC/WETH", positionDetails(p))
}

func TestPositionDetails_Nft(t *testing.T) {
	p := defiPosition{
		Type:         "Nft",
		Token0Symbol: "WETH",
		Token1Symbol: "USDC",
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
		Type:         "NftV4",
		Token0Symbol: "WBTC",
		Token1Symbol: "WETH",
		Positions: []nftPositionDetails{
			{TickLower: -50, TickUpper: 50, TokenID: "0xabc"},
		},
	}
	assert.Equal(t, "WBTC/WETH (1 position)", positionDetails(p))
}

func TestPositionDetails_NftNoPositions(t *testing.T) {
	p := defiPosition{
		Type:         "Nft",
		Token0Symbol: "DAI",
		Token1Symbol: "USDC",
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
