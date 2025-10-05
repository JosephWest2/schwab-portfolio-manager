package balance

import (
	"cmp"
	"log"
	"maps"
	"math"
	"slices"

	"github.com/josephwest2/schwab-portfolio-manager/targetAllocation"
)

type Ticker = targetAllocation.Ticker
type ShareQuantity = float64

type Holding struct {
	Ticker Ticker
	Amount ShareQuantity
}

// sort by deviation from expected proportion
func PurchasePriorityFunc(totalHoldingsValue float64, prices map[Ticker]float64, proportionTargets map[Ticker]float64) func(a, b Holding) int {
	AssertValidDesiredAllocationPrices(slices.Collect(maps.Keys(proportionTargets)), prices)
	return func(a, b Holding) int {
		va := float64(a.Amount) * prices[a.Ticker]
		vb := float64(b.Amount) * prices[b.Ticker]
		da := va/totalHoldingsValue - proportionTargets[a.Ticker]
		db := vb/totalHoldingsValue - proportionTargets[b.Ticker]
		return cmp.Compare(da, db)
	}
}

// panics on error
func AssertValidHoldingPrices(holdings map[Ticker]float64, prices map[Ticker]float64) {
	for k := range holdings {
		if _, ok := prices[k]; !ok {
			log.Fatal("price for " + k + " in holdings not found")
		}
	}
}

// panics on error
func AssertValidDesiredAllocationPrices(tickers []Ticker, prices map[Ticker]float64) {
	for _, ticker := range tickers {
		if _, ok := prices[ticker]; !ok {
			log.Fatal("price for " + ticker + " not found")
		}
	}
}

// returns purchases to be made and remaining cash
func FillProportions(cash float64, holdings map[Ticker]float64, prices map[Ticker]float64, proportionTargets map[Ticker]float64) (map[Ticker]float64, float64) {
	holdingsSlice := make([]Holding, 0)
	for ticker := range proportionTargets {
		holdingsSlice = append(holdingsSlice, Holding{ticker, holdings[ticker]})
	}
	minPrice := math.MaxFloat64
	availablePrices := make(map[Ticker]float64)
	for ticker := range proportionTargets {
		availablePrices[ticker] = prices[ticker]
	}
	for _, v := range availablePrices {
		if v < minPrice {
			minPrice = v
		}
	}
	purchases := make(map[Ticker]float64, 0)
	for cash >= minPrice {
		totalHoldingsValue := 0.0
		for _, v := range holdingsSlice {
			totalHoldingsValue += v.Amount * prices[v.Ticker]
		}
		slices.SortFunc(holdingsSlice, PurchasePriorityFunc(totalHoldingsValue, prices, proportionTargets))
		// buy the asset with the most negative deviation that can be afforded
		for i, holding := range holdingsSlice {
			if prices[holding.Ticker] > cash {
				continue
			}
			purchases[holding.Ticker] += 1
			holdingsSlice[i].Amount += 1
			cash -= prices[holding.Ticker]
			break
		}
	}
	return purchases, cash
}

// Returns purchases to be made and remaining cash.
// Note that partial shares can not be bought
func BalancePurchase(cash float64, holdings map[Ticker]float64, prices map[Ticker]float64, targetAllocation targetAllocation.TargetAllocation) (map[Ticker]float64, float64) {

	AssertValidDesiredAllocationPrices(slices.Collect(maps.Keys(targetAllocation)), prices)
	AssertValidHoldingPrices(holdings, prices)

	fixedTargets := make(map[Ticker]float64, 0)
	for ticker, alloc := range targetAllocation {
		if alloc.FixedCashValue != 0 {
			fixedTargets[ticker] = alloc.FixedCashValue
		}
	}
	proportionTargets := make(map[Ticker]float64, 0)
	for ticker, alloc := range targetAllocation {
		if alloc.Proportion != 0 {
			proportionTargets[ticker] = alloc.Proportion
		}
	}

	fixedPurchases, cash := FillFixed(cash, holdings, prices, fixedTargets)
	proportionPurchases, cash := FillProportions(cash, holdings, prices, proportionTargets)

	purchases := make(map[Ticker]float64, 0)
	for k, v := range fixedPurchases {
		purchases[k] += v
	}
	for k, v := range proportionPurchases {
		purchases[k] += v
	}
	return purchases, cash
}

// returns purchases to be made and remaining cash
func FillFixed(cash float64, holdings map[Ticker]float64, prices map[Ticker]float64, fixedTargets map[Ticker]float64) (map[Ticker]float64, float64) {
	AssertValidDesiredAllocationPrices(slices.Collect(maps.Keys(fixedTargets)), prices)
	AssertValidHoldingPrices(holdings, prices)
	result := make(map[Ticker]float64, 0)
	for ticker, alloc := range fixedTargets {
		diff := alloc - holdings[ticker]*prices[ticker]
		if diff > 0 {
			spendAmount := math.Min(diff*prices[ticker], cash)
			r := math.Floor(spendAmount / prices[ticker])
			if r > 0 {
				result[ticker] = r
			}
			cash -= float64(r) * prices[ticker]
		}
	}
	return result, cash
}

// returns purchases and sales to be made and remaining cash
func RebalanceWithSelling(cash float64, holdings map[Ticker]float64, prices map[Ticker]float64, targetAllocation targetAllocation.TargetAllocation) (map[Ticker]float64, float64) {
	AssertValidDesiredAllocationPrices(slices.Collect(maps.Keys(targetAllocation)), prices)
	AssertValidHoldingPrices(holdings, prices)

	// simulate selling all stocks and buying at proper proportions
	// exclude fractional shares from selling logic
	newHoldings := make(map[Ticker]float64, 0)
	for ticker, quantity := range holdings {
		cash += math.Floor(quantity) * prices[ticker]
	}
	newHoldings, cash = BalancePurchase(cash, newHoldings, prices, targetAllocation)
	purchasesAndSales := make(map[Ticker]float64, 0)
	for ticker := range newHoldings {
		difference := float64(int64(newHoldings[ticker] - holdings[ticker]))
		if difference != 0 {
			purchasesAndSales[ticker] = difference
		}
	}
	return purchasesAndSales, cash
}
