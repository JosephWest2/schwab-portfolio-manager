package balance

import (
	"cmp"
	"errors"
	"fmt"
	"log"
	"maps"
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

func (da *DesiredAllocations) Tickers() []string {
	return append(slices.Collect(maps.Keys(da.Proportions)), slices.Collect(maps.Keys(da.FixedCashAmounts))...)
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
	AssertValidDesiredAllocationPrices(desiredAllocations, prices)
	return func(a, b Holding) int {
		va := float64(a.Count) * prices[a.Ticker]
		vb := float64(b.Count) * prices[b.Ticker]
		da := va/totalHoldingsValue - desiredAllocations.Proportions[a.Ticker]
		db := vb/totalHoldingsValue - desiredAllocations.Proportions[b.Ticker]
		return cmp.Compare(da, db)
	}
}

// panics on error
func AssertValidHoldingPrices(holdings map[string]float64, prices map[string]float64) {
	for k := range holdings {
		if _, ok := prices[k]; !ok {
			log.Fatal("price for " + k + " in holdings not found")
		}
	}
}

// panics on error
func AssertValidDesiredAllocationPrices(desiredAllocations *DesiredAllocations, prices map[string]float64) {
	for k := range desiredAllocations.Proportions {
		if _, ok := prices[k]; !ok {
			log.Fatal("price for " + k + " in desiredAllocations.Proportions not found")
		}
	}
	for k := range desiredAllocations.FixedCashAmounts {
		if _, ok := prices[k]; !ok {
			log.Fatal("price for " + k + " in desiredAllocaitons.FixedCashAmounts not found")
		}
	}
}

// Returns purchases to be made and remaining cash.
// Note that partial shares can not be bought
func BalancePurchase(cash float64, holdings map[string]float64, prices map[string]float64, desiredAllocations *DesiredAllocations) (map[string]float64, float64) {

	AssertValidDesiredAllocationPrices(desiredAllocations, prices)
	AssertValidHoldingPrices(holdings, prices)

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
			totalHoldingsValue += v.Count * prices[v.Ticker]
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
func FillFixedAmounts(cash float64, holdings map[string]float64, prices map[string]float64, desiredAllocations *DesiredAllocations) (map[string]float64, float64) {
	AssertValidDesiredAllocationPrices(desiredAllocations, prices)
	AssertValidHoldingPrices(holdings, prices)
	result := make(map[string]float64, 0)
	for k, v := range desiredAllocations.FixedCashAmounts {
		diff := v - holdings[k]*prices[k]
		if diff > 0 {
			spendAmount := math.Min(diff*prices[k], cash)
			r := math.Floor(spendAmount / prices[k])
			if r > 0 {
				result[k] = r
			}
			cash -= float64(r) * prices[k]
		}
	}
	return result, cash
}


// returns purchases and sales to be made and remaining cash
func RebalanceWithSelling(cash float64, holdings map[string]float64, prices map[string]float64, desiredAllocations *DesiredAllocations) (map[string]float64, float64) {
	AssertValidDesiredAllocationPrices(desiredAllocations, prices)
	AssertValidHoldingPrices(holdings, prices)

	initialCash := cash
	// simulate selling all stocks and buying at proper proportions
	// maintain fractional holdings outside this calculation
	fractionalHoldings := make(map[string]float64, 0)
	for k, v := range holdings {
		_, frac := math.Modf(v)
		cash += v * prices[k] - frac * prices[k]
	}
	newHoldings, cash := BalancePurchase(cash, fractionalHoldings, prices, desiredAllocations)
	for k := range newHoldings {
		newHoldings[k] += fractionalHoldings[k]
	}
	purchasesAndSales := make(map[string]float64, 0)
	for k, v := range holdings {
		ri := newHoldings[k] - v
		r := float64(int64(newHoldings[k]-v))
		fmt.Println(ri, r)
		if r != 0 {
			purchasesAndSales[k] = r
		}
	}
	purchaseCount := 0
	for _, v := range purchasesAndSales {
		if v > 0 {
			purchaseCount++
		}
	}
	if purchaseCount < 1 {
		return make(map[string]float64, 0), initialCash
	}
	return purchasesAndSales, cash
}
