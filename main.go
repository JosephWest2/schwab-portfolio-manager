package main

import (
	"math"
	"sort"
)

func main() {

}

type Asset struct {
	Ticker string
	Amount float64
	Price  float64
}

var tickerAllocations map[string]float64 = map[string]float64{
	"DFAC": 0.64,
	"DFIC": 0.27,
	"DFEM": 0.09,
}

func minAssetPrice(assets []Asset) float64 {
	if len(assets) == 0 {
		panic("empty asset array passed to assetMinPrice")
	}
	min := assets[0].Price
	for _, v := range assets {
		if v.Price < min {
			min = v.Price
		}
	}
	return min
}

func sumAssetValues(assets []Asset) float64 {
	total := 0.0
	for _, v := range assets {
		total += v.Price * v.Amount
	}
	return total
}

func sortByPurchasePriority(assets []Asset, proportions map[string]float64) {
	totalValue := sumAssetValues(assets)
	sort.Slice(assets, func(i, j int) bool {
		vi := assets[i].Amount * assets[i].Price
		vj := assets[j].Amount * assets[j].Price
		return vi/totalValue-proportions[assets[i].Ticker] < vj/totalValue-proportions[assets[j].Ticker]
	})
}

// returns purchases to be made, remaining cash, and new assets
func balanceAllocation(cash float64, assets []Asset, proportions map[string]float64) (map[string]int64, float64, []Asset) {
	newAssets := make([]Asset, len(assets))
	copy(newAssets, assets)
	assets = newAssets
	purchases := make(map[string]int64, 0)
	minAssetPrice := minAssetPrice(assets)
	for cash >= minAssetPrice {
		sortByPurchasePriority(assets, proportions)
		for i := range assets {
			if assets[i].Price > cash {
				continue
			}
			purchases[assets[i].Ticker] += 1
			assets[i].Amount += 1
			cash -= assets[i].Price
			break
		}
	}
	return purchases, cash, assets
}
func getAssetDeviation(asset Asset, total float64, proportion float64) float64 {
	return math.Pow(asset.Amount*asset.Price/total-proportion, 2)
}

func getAssetsDeviation(assets []Asset, proportions map[string]float64) float64 {
	totalValue := sumAssetValues(assets)
	deviation := 0.0
	for _, v := range assets {
		deviation += getAssetDeviation(v, totalValue, proportions[v.Ticker])
	}
	return deviation
}

// returns purchases to be made, remaining cash, and new assets
func rebalanceWithSelling(cash float64, assets []Asset, proportions map[string]float64) (map[string]int64, float64, []Asset) {
	purchasesAndSales, cash, assets := balanceAllocation(cash, assets, proportions)
	deviation := getAssetsDeviation(assets, proportions)
outer:
	for {
		sortByPurchasePriority(assets, proportions)
		j := len(assets) - 1
		for j > 0 {
			i := 0
			for i < j {
				if assets[i].Price+cash <= assets[j].Price {
					newAssets := make([]Asset, len(assets))
					copy(newAssets, assets)
					newAssets[i].Amount++
					newAssets[j].Amount--
					newDeviation := getAssetsDeviation(newAssets, proportions)
					if newDeviation < deviation {
						deviation = newDeviation
						assets = newAssets
						cashChange := assets[j].Price - assets[i].Price
						cash += cashChange
						purchasesAndSales[assets[i].Ticker]++
						purchasesAndSales[assets[j].Ticker]--
						continue outer
					}
					break
				}
				i++
			}
			j--
		}
		break
	}
	return purchasesAndSales, cash, assets
}
