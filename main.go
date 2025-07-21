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

// returns purchases to be made, alters assets to reflect changes
func balanceAllocation(cash float64, assets []Asset, proportions map[string]float64) map[string]int64 {
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
	return purchases
}

func deviation(assets []Asset, proportions map[string]float64) float64 {
	totalValue := sumAssetValues(assets)
	deviation := 0.0
	for _, v := range assets {
		deviation += math.Pow(v.Amount * v.Price / totalValue - proportions[v.Ticker], 2)
	}
	return deviation
}

// returns purchases and sales to be made, alters assets to reflect changes
func rebalanceWithSelling(cash float64, assets []Asset, proportions map[string]float64) map[string]int64 {
	purchasesAndSales := balanceAllocation(cash, assets, proportions)
	cashSpent := 0.0
	for _, v := range assets {
		cashSpent += v.Price * float64(purchasesAndSales[v.Ticker])
	}
	cash -= cashSpent
	deviation := deviation(assets, proportions)
	newDeviation := 0.0
	for newDeviation < deviation {
		i := 0
		j := len(assets) - 1
		for i < j {

		}
	}



}
