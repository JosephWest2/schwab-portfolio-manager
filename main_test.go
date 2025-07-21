package main

import (
	"fmt"
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
	purchasesAndSales, cash, assets := rebalanceWithSelling(5, assets, proportions)
	fmt.Println("assets", assets)
	fmt.Println("cash", cash)
	fmt.Println(purchasesAndSales)
	total := sumAssetValues(assets)
	for _, v := range assets {
		fmt.Println(v.Ticker, v.Amount*v.Price/total)
	}
}
