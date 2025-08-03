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

type DesiredAllocations struct {
	Proportions      map[string]float64
	FixedCashAmounts map[string]float64
}

func NewDesiredAllocations() *DesiredAllocations {
	return &DesiredAllocations{
		Proportions:      make(map[string]float64),
		FixedCashAmounts: make(map[string]float64),
	}
}

func LoadDesiredAllocations(filepath string) (*DesiredAllocations, error) {

	type DesiredAllocationsYaml struct {
		Ticker          string  `yaml:"ticker"`
		Proportion      float64 `yaml:"proportion"`
		FixedCashAmount float64 `yaml:"fixedCashAmount"`
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, errors.New("failed to read allocation file: " + err.Error())
	}

	var desiredAllocations []DesiredAllocationsYaml

	err = yaml.Unmarshal(data, &desiredAllocations)
	if err != nil {
		return nil, errors.New("failed to parse allocation file: " + err.Error())
	}

	result := NewDesiredAllocations()
	for _, v := range desiredAllocations {
		if v.FixedCashAmount > 0 {
			result.FixedCashAmounts[v.Ticker] = v.FixedCashAmount
		} else {
			result.Proportions[v.Ticker] = v.Proportion
		}
	}

	sum := 0.0
	for _, v := range result.Proportions {
		sum += v
	}
	if !AlmostEqual(sum, 1.0, 1e-7) {
		return nil, errors.New("allocation proportions do not sum to 1.0")
	}

	return result, err
}

type Holding struct {
	Ticker string
	Count  float64
}

// sort by deviation from expected proportion
func PurchasePriorityFunc(totalHoldingsValue float64, prices map[string]float64, desiredAllocations *DesiredAllocations) func(a, b Holding) int {
	return func(a, b Holding) int {
		va := float64(a.Count) * prices[a.Ticker]
		vb := float64(b.Count) * prices[b.Ticker]
		da := va/totalHoldingsValue - desiredAllocations.Proportions[a.Ticker]
		db := vb/totalHoldingsValue - desiredAllocations.Proportions[b.Ticker]
		return cmp.Compare(da, db)
	}
}

// returns purchases to be made and remaining cash
func BalancePurchase(cash float64, holdings map[string]float64, prices map[string]float64, desiredAllocations *DesiredAllocations) (map[string]int64, float64) {
	purchases, cash := FillFixedAmounts(cash, holdings, prices, desiredAllocations)
	holdingsSlice := make([]Holding, 0, len(desiredAllocations.Proportions))
	for k := range desiredAllocations.Proportions {
		holdingsSlice = append(holdingsSlice, Holding{k, holdings[k]})
	}
	minPrice := math.MaxFloat64
	availablePrices := make(map[string]float64)
	for k := range desiredAllocations.Proportions {
		availablePrices[k] = prices[k]
	}
	for _, v := range availablePrices {
		if v < minPrice {
			minPrice = v
		}
	}
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

// returns purchases to be made and remaining cash
func FillFixedAmounts(cash float64, holdings map[string]float64, prices map[string]float64, desiredAllocations *DesiredAllocations) (map[string]int64, float64) {
	result := make(map[string]int64, 0)
	for k, v := range desiredAllocations.FixedCashAmounts {
		diff := v - holdings[k]*prices[k]
		if diff > 0 {
			result[k] = int64(math.Ceil(diff / prices[k]))
			cash -= float64(result[k]) * prices[k]
		}
	}
	return result, cash
}

// returns sales to be made
func SellExcessFixed(holdings map[string]float64, prices map[string]float64, desiredAllocations *DesiredAllocations) map[string]int64 {
	result := make(map[string]int64, 0)
	for k, v := range desiredAllocations.FixedCashAmounts {
		heldValue := holdings[k] * prices[k]
		if heldValue > v {
			result[k] = int64(math.Floor((heldValue - v) / prices[k]))
		}
	}
	return result
}

// returns purchases and sales to be made and remaining cash
func RebalanceWithSelling(cash float64, holdings map[string]int64, prices map[string]float64, desiredAllocations *DesiredAllocations) (map[string]int64, float64) {
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
