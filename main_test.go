package main

import (
	"math"
	"testing"
)

func TestBalanceAllocation(t *testing.T) {
	cash := 500.0
	assets := []Asset{
		{"DFAC", 30, 30},
		{"DFIC", 20, 20},
		{"DFEM", 10, 10},
	}
	proportions := map[string]float64{
		"DFAC": 0.64,
		"DFIC": 0.27,
		"DFEM": 0.09,
	}
	expected := map[string]int64{
		"DFAC": 10,
		"DFIC": 6,
		"DFEM": 8,
	}
	purchases, cash, assets := balanceAllocation(cash, assets, proportions)
	if len(purchases) != len(expected) {
		t.Fatalf("Expected %v, got %v", expected, purchases)
	}
	for k, v := range purchases {
		if v != expected[k] {
			t.Fatalf("Expected %v, got %v", expected, purchases)
		}
	}
	expectedTotals := map[string]float64{
		"DFAC": 40,
		"DFIC": 26,
		"DFEM": 18,
	}
	for _, v := range assets {
		if v.Amount != expectedTotals[v.Ticker] {
			t.Fatalf("Expected %v, got %v", expectedTotals, assets)
		}
	}
}

func TestSortByPurchasePriority(t *testing.T) {
	assets := []Asset{
		{"DFAC", 20, 30},
		{"DFIC", 20, 20},
		{"DFEM", 10, 10},
	}
	proportions := map[string]float64{
		"DFAC": 0.64,
		"DFIC": 0.27,
		"DFEM": 0.09,
	}
	expected := []string{"DFAC", "DFEM", "DFIC"}
	sortByPurchasePriority(assets, proportions)
	for i, v := range assets {
		if v.Ticker != expected[i] {
			t.Fatalf("Expected %v, got %v", expected, assets)
		}
	}

	assets = []Asset{
		{"DFAC", 30, 35},
		{"DFIC", 20, 10},
		{"DFEM", 20, 10},
	}
	proportions = map[string]float64{
		"DFAC": 0.64,
		"DFIC": 0.27,
		"DFEM": 0.09,
	}
	expected = []string{"DFIC", "DFEM", "DFAC"}
	sortByPurchasePriority(assets, proportions)
	for i, v := range assets {
		if v.Ticker != expected[i] {
			t.Fatalf("Expected %v, got %v", expected, assets)
		}
	}
}

func TestAllocationTotal(t *testing.T) {
	total := 0.0
	for _, v := range tickerAllocations {
		total += v
	}
	if total != 1.0 {
		t.Errorf("Allocation total should be 1.0, was %v", total)
	}
}

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestRebalanceWithSelling(t *testing.T) {
	assets := []Asset{
		{"DFAC", 30, 9.8},
		{"DFIC", 30, 10.2},
		{"DFEM", 30, 10.1},
	}
	proportions := map[string]float64{
		"DFAC": 0.64,
		"DFIC": 0.27,
		"DFEM": 0.09,
	}
	cash := 5.0
	oldSumValue := sumAssetValues(assets) + cash
	purchasesAndSales, cash, assets := rebalanceWithSelling(cash, assets, proportions)
	newSumValue := sumAssetValues(assets) + cash
	if !almostEqual(oldSumValue, newSumValue, 1e-7) {
		t.Errorf("Expected oldSumValue and newSumValue to be equal, was %v and %v", oldSumValue, newSumValue)
	}
	expectedAssets := map[string]float64{
		"DFAC": 59,
		"DFIC": 24,
		"DFEM": 8,
	}
	for _, v := range assets {
		if v.Amount != expectedAssets[v.Ticker] {
			t.Errorf("Expected %v, got %v", expectedAssets, assets)
		}
	}
	expectedPurchasesAndSales := map[string]int64{
		"DFAC": 29,
		"DFIC": -6,
		"DFEM": -22,
	}
	for k, v := range purchasesAndSales {
		if v != expectedPurchasesAndSales[k] {
			t.Errorf("Expected %v, got %v", expectedPurchasesAndSales, purchasesAndSales)
		}
	}
}
