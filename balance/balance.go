package balance

import (
	"cmp"
	"errors"
	"math"
	"os"
	"slices"

	"github.com/goccy/go-yaml"
)

var DesiredAllocationsFile = "desiredAllocations.yaml"

func AlmostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func LoadDesiredAllocations(filepath string) (map[string]float64, error) {

	type DesiredAllocations struct {
		Ticker     string  `yaml:"ticker"`
		Proportion float64 `yaml:"proportion"`
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, errors.New("failed to read allocation file: " + err.Error())
	}

	var desiredAllocations []DesiredAllocations

	err = yaml.Unmarshal(data, &desiredAllocations)
	if err != nil {
		return nil, errors.New("failed to parse allocation file: " + err.Error())
	}

	result := make(map[string]float64, len(desiredAllocations))
	for _, v := range desiredAllocations {
		result[v.Ticker] = v.Proportion
	}

	sum := 0.0
	for _, v := range result {
		sum += v
	}
	if !AlmostEqual(sum, 1.0, 1e-7) {
		return nil, errors.New("allocations do not sum to 1.0")
	}

	return result, err
}

type Holding struct {
	Ticker string
	Count  float64
}

// sort by deviation from expected proportion
func PurchasePriorityFunc(totalHoldingsValue float64, prices map[string]float64, desiredAllocations map[string]float64) func(a, b Holding) int {
	return func(a, b Holding) int {
		va := float64(a.Count) * prices[a.Ticker]
		vb := float64(b.Count) * prices[b.Ticker]
		da := va/totalHoldingsValue - desiredAllocations[a.Ticker]
		db := vb/totalHoldingsValue - desiredAllocations[b.Ticker]
		return cmp.Compare(da, db)
	}
}

// returns purchases to be made and remaining cash
func BalancePurchase(cash float64, holdings map[string]float64, prices map[string]float64, desiredAllocations map[string]float64) (map[string]int64, float64) {
	holdingsSlice := make([]Holding, 0, len(desiredAllocations))
	for k := range desiredAllocations {
		holdingsSlice = append(holdingsSlice, Holding{k, holdings[k]})
	}
	minPrice := math.MaxFloat64
	for _, v := range prices {
		if v < minPrice {
			minPrice = v
		}
	}
	purchases := make(map[string]int64, 0)
	for cash >= minPrice {
		totalHoldingsValue := 0.0
		for _, v := range holdingsSlice {
			totalHoldingsValue += float64(v.Count) * prices[v.Ticker]
		}
		slices.SortFunc(holdingsSlice, PurchasePriorityFunc(totalHoldingsValue, prices, desiredAllocations))
		// buy the asset with the most negative deviation that can be afforded
		for i, v := range holdingsSlice {
			if prices[v.Ticker] > cash {
				continue
			}
			purchases[v.Ticker] += 1
			holdingsSlice[i].Count += 1
			cash -= prices[v.Ticker]
			break
		}
	}
	return purchases, cash
}

// returns purchases and sales to be made and remaining cash
func RebalanceWithSelling(cash float64, holdings map[string]int64, prices map[string]float64, desiredAllocations map[string]float64) (map[string]int64, float64) {
	// simulate selling all stocks and buying at proper proportions
	for k, v := range holdings {
		cash += float64(v) * prices[k]
	}
	newHoldings, cash := BalancePurchase(cash, nil, prices, desiredAllocations)
	purchasesAndSales := make(map[string]int64, 0)
	for k, v := range holdings {
		purchasesAndSales[k] = newHoldings[k] - v
	}
	return purchasesAndSales, cash
}
