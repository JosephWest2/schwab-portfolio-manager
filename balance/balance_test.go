package balance

import (
	"reflect"
	"testing"

	"github.com/josephwest2/schwab-portfolio-manager/targetAllocation"
	"github.com/josephwest2/schwab-portfolio-manager/util"
)

func TestBalancePurchase(t *testing.T) {
	alloc1, err := targetAllocation.LoadTargetAllocations("testing/targetAllocation_balance_test1.yaml")
	if err != nil {
		t.Fatal("cannot continue testing TestBalancePurchase: " + err.Error())
	}
	alloc2, err := targetAllocation.LoadTargetAllocations("testing/targetAllocation_balance_test2.yaml")
	if err != nil {
		t.Fatal("cannot continue testing TestBalancePurchase: " + err.Error())
	}

	tests := []struct {
		cash                  float64
		targetAllocation      targetAllocation.TargetAllocation
		holdings              map[string]float64
		prices                map[string]float64
		expectedPurchases     map[string]float64
		expectedCashRemaining float64
	}{
		{
			cash:             503.1,
			targetAllocation: alloc1["global"],
			holdings: map[string]float64{
				"DFAC":  30,
				"DFIC":  20,
				"DFEM":  10,
				"SWVXX": 3998,
			},
			prices: map[string]float64{
				"DFAC":  30,
				"DFIC":  20,
				"DFEM":  10,
				"SWVXX": 1,
			},
			expectedPurchases: map[string]float64{
				"DFAC":  10,
				"DFIC":  6,
				"DFEM":  8,
				"SWVXX": 2,
			},
			expectedCashRemaining: 1.1,
		},
		{
			cash:             999.99,
			targetAllocation: alloc1["global"],
			holdings: map[string]float64{
				"DFAC":  55,
				"DFIC":  27,
				"DFEM":  9,
				"SWVXX": 3996,
			},
			prices: map[string]float64{
				"DFAC":  100.01,
				"DFIC":  100.01,
				"DFEM":  100.01,
				"SWVXX": 1,
			},
			expectedPurchases: map[string]float64{
				"DFAC":  9,
				"SWVXX": 4,
			},
			expectedCashRemaining: 95.90,
		},
		{
			cash:             1501.5,
			targetAllocation: alloc2["567"],
			holdings: map[string]float64{
				"VTI":   10,
				"VSAIX": 10,
				"VXUS":  10,
				"VWO":   10,
				"SWVXX": 3500,
			},
			prices: map[string]float64{
				"VTI":   50,
				"VSAIX": 20,
				"VXUS":  20,
				"VWO":   10,
				"SWVXX": 1,
			},
			expectedPurchases: map[string]float64{
				"VTI":   10,
				"VSAIX": 10,
				"VXUS":  10,
				"VWO":   10,
				"SWVXX": 500,
			},
			expectedCashRemaining: 1.5,
		},
		{
			cash:             0.5,
			targetAllocation: alloc2["567"],
			holdings: map[string]float64{
				"VTI":   10,
				"VSAIX": 10,
				"VXUS":  10,
			},
			prices: map[string]float64{
				"VTI":   50,
				"VSAIX": 20,
				"VXUS":  20,
				"VWO":   10,
				"SWVXX": 1,
			},
			expectedPurchases:     map[string]float64{},
			expectedCashRemaining: 0.5,
		},
	}
	for i, test := range tests {
		purchases, cash := BalancePurchase(test.cash, test.holdings, test.prices, test.targetAllocation)
		if !reflect.DeepEqual(purchases, test.expectedPurchases) {
			t.Errorf("expected purchases: %v, got %v, test index: %v", test.expectedPurchases, purchases, i)
		}
		if !util.AlmostEqual(cash, test.expectedCashRemaining, 1e-7) {
			t.Errorf("expected purchases: %v, got %v, test index: %v", test.expectedPurchases, purchases, i)
		}
	}
}

func TestRebalanceWithSelling(t *testing.T) {
	alloc1, err := targetAllocation.LoadTargetAllocations("testing/targetAllocation_balance_test1.yaml")
	if err != nil {
		t.Fatal("cannot continue testing TestRebalanceWithSelling: " + err.Error())
	}
	alloc2, err := targetAllocation.LoadTargetAllocations("testing/targetAllocation_balance_test2.yaml")
	if err != nil {
		t.Fatal("cannot continue testing TestRebalanceWithSelling: " + err.Error())
	}
	alloc5, err := targetAllocation.LoadTargetAllocations("testing/targetAllocation_balance_test5.yaml")
	if err != nil {
		t.Fatal("cannot continue testing TestRebalanceWithSelling: " + err.Error())
	}

	tests := []struct {
		cash                      float64
		targetAllocation          targetAllocation.TargetAllocation
		holdings                  map[string]float64
		prices                    map[string]float64
		expectedPurchasesAndSales map[string]float64
		expectedCashRemaining     float64
	}{
		{
			cash:             0.32,
			targetAllocation: alloc1["global"],
			holdings: map[string]float64{
				"DFAC":  66,
				"DFIC":  22,
				"DFEM":  12,
				"SWVXX": 4000,
			},
			prices: map[string]float64{
				"DFAC":  1,
				"DFIC":  1,
				"DFEM":  1,
				"SWVXX": 1,
			},
			expectedPurchasesAndSales: map[string]float64{
				"DFAC": -2,
				"DFIC": 5,
				"DFEM": -3,
			},
			expectedCashRemaining: 0.32,
		},
		{
			cash:             0.55,
			targetAllocation: alloc5["123"],
			holdings: map[string]float64{
				"DFAC": 1000,
				"DFIC": 1000,
				"DFEM": 1000,
				"SWVXX": 2000,
			},
			prices: map[string]float64{
				"DFAC": 64.001,
				"DFIC": 27.001,
				"DFEM": 9.001,
				"SWVXX": 2000,
			},
			expectedPurchasesAndSales: map[string]float64{},
			expectedCashRemaining:     0.55,
		},
		{
			cash:             0.32,
			targetAllocation: alloc1["global"],
			holdings: map[string]float64{
				"DFAC":  4066,
				"DFIC":  22,
				"DFEM":  12,
				"SWVXX": 0,
			},
			prices: map[string]float64{
				"DFAC":  1,
				"DFIC":  1,
				"DFEM":  1,
				"SWVXX": 1,
			},
			expectedPurchasesAndSales: map[string]float64{
				"DFAC":  -4002,
				"DFIC":  5,
				"DFEM":  -3,
				"SWVXX": 4000,
			},
			expectedCashRemaining: 0.32,
		},
		{
			cash:             0.99,
			targetAllocation: alloc1["global"],
			holdings: map[string]float64{
				"DFAC":  170,
				"DFIC":  5,
				"DFEM":  25,
				"SWVXX": 4000,
			},
			prices: map[string]float64{
				"DFAC":  1,
				"DFIC":  1,
				"DFEM":  1,
				"SWVXX": 1,
			},
			expectedPurchasesAndSales: map[string]float64{
				"DFAC": -42,
				"DFIC": 49,
				"DFEM": -7,
			},
			expectedCashRemaining: 0.99,
		},
		{
			cash:             202.12,
			targetAllocation: alloc2["567"],
			holdings: map[string]float64{
				"VTI":   10,
				"VSAIX": 10,
				"VXUS":  10,
				"VWO":   10,
				"SWVXX": 4000,
			},
			prices: map[string]float64{
				"VTI":   10,
				"VSAIX": 10,
				"VXUS":  10,
				"VWO":   10,
				"SWVXX": 1,
			},
			expectedPurchasesAndSales: map[string]float64{
				"VTI":   20,
				"VSAIX": 2,
				"VXUS":  2,
				"VWO":   -4,
			},
			expectedCashRemaining: 2.12,
		},
		{
			cash:             0.10,
			targetAllocation: alloc1["global"],
			holdings: map[string]float64{
				"DFAC":  64.1,
				"DFIC":  27.05,
				"DFEM":  9.08,
				"SWVXX": 4000,
			},
			prices: map[string]float64{
				"DFAC":  1,
				"DFIC":  1,
				"DFEM":  1,
				"SWVXX": 1,
			},
			expectedPurchasesAndSales: map[string]float64{},
			expectedCashRemaining:     0.10,
		},
	}

	for i, test := range tests {
		purchasesAndSales, cash := RebalanceWithSelling(test.cash, test.holdings, test.prices, test.targetAllocation)
		if !reflect.DeepEqual(purchasesAndSales, test.expectedPurchasesAndSales) {
			t.Errorf("expected purchases and sales: %v, got %v, on test index %v", test.expectedPurchasesAndSales, purchasesAndSales, i)
		}
		if !util.AlmostEqual(cash, test.expectedCashRemaining, 1e-7) {
			t.Errorf("expected purchases and sales: %v, got %v, on test index %v", test.expectedPurchasesAndSales, purchasesAndSales, i)
		}
	}
}
