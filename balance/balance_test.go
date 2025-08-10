package balance

import (
	"reflect"
	"testing"
)

func TestLoadAllocations(t *testing.T) {
	tests := []struct {
		filepath string
		expected *DesiredAllocations
		wantErr  bool
	}{
		{
			filepath: "testing/desiredAllocations_test1.yaml",
			expected: &DesiredAllocations{
				Proportions: map[string]float64{
					"DFAC": 0.64,
					"DFIC": 0.27,
					"DFEM": 0.09,
				},
				FixedCashAmounts: map[string]float64{
					"SWVXX": 4000,
				},
			},
			wantErr: false,
		},
		{
			filepath: "testing/desiredAllocations_test2.yaml",
			expected: &DesiredAllocations{
				Proportions: map[string]float64{
					"VTI":   0.50,
					"VSIAX": 0.20,
					"VXUS":  0.20,
					"VWO":   0.10,
				},
				FixedCashAmounts: map[string]float64{
					"SWVXX": 4000,
				},
			},
			wantErr: false,
		},
		{
			// Allocation sums to 0.999 not 1
			filepath: "testing/desiredAllocations_test3.yaml",
			expected: nil,
			wantErr:  true,
		},
		{
			// Allocation sums to 1.001 not 1
			filepath: "testing/desiredAllocations_test4.yaml",
			expected: nil,
			wantErr:  true,
		},
	}
	for _, test := range tests {
		allocations, err := LoadDesiredAllocations(test.filepath)
		if test.wantErr && err == nil {
			t.Errorf("expected error on %v, got no error", test)
		}
		equal := reflect.DeepEqual(allocations, test.expected)
		if !equal {
			t.Errorf("expected %v, got %v", test.expected, allocations)
		}
	}
}

func TestBalancePurchase(t *testing.T) {
	alloc1, err := LoadDesiredAllocations("testing/desiredAllocations_test1.yaml")
	if err != nil {
		t.Fatal("cannot continue testing TestBalancePurchase: " + err.Error())
	}
	alloc2, err := LoadDesiredAllocations("testing/desiredAllocations_test2.yaml")
	if err != nil {
		t.Fatal("cannot continue testing TestBalancePurchase: " + err.Error())
	}

	tests := []struct {
		cash                  float64
		desiredAllocations    *DesiredAllocations
		holdings              map[string]float64
		prices                map[string]float64
		expectedPurchases     map[string]int64
		expectedCashRemaining float64
	}{
		{
			cash:               503.1,
			desiredAllocations: alloc1,
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
			expectedPurchases: map[string]int64{
				"DFAC":  10,
				"DFIC":  6,
				"DFEM":  8,
				"SWVXX": 2,
			},
			expectedCashRemaining: 1.1,
		},
		{
			cash:               999.99,
			desiredAllocations: alloc1,
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
			expectedPurchases: map[string]int64{
				"DFAC":  9,
				"SWVXX": 4,
			},
			expectedCashRemaining: 95.90,
		},
		{
			cash:               1501.5,
			desiredAllocations: alloc2,
			holdings: map[string]float64{
				"VTI":   10,
				"VSIAX": 10,
				"VXUS":  10,
				"VWO":   10,
				"SWVXX": 3500,
			},
			prices: map[string]float64{
				"VTI":   50,
				"VSIAX": 20,
				"VXUS":  20,
				"VWO":   10,
				"SWVXX": 1,
			},
			expectedPurchases: map[string]int64{
				"VTI":   10,
				"VSIAX": 10,
				"VXUS":  10,
				"VWO":   10,
				"SWVXX": 500,
			},
			expectedCashRemaining: 1.5,
		},
		{
			cash:               0.5,
			desiredAllocations: alloc2,
			holdings: map[string]float64{
				"VTI":   10,
				"VSIAX": 10,
				"VXUS":  10,
			},
			prices: map[string]float64{
				"VTI":   50,
				"VSIAX": 20,
				"VXUS":  20,
				"VWO":   10,
				"SWVXX": 1,
			},
			expectedPurchases:     map[string]int64{},
			expectedCashRemaining: 0.5,
		},
	}
	for _, test := range tests {
		purchases, cash := BalancePurchase(test.cash, test.holdings, test.prices, test.desiredAllocations)
		if !reflect.DeepEqual(purchases, test.expectedPurchases) {
			t.Errorf("expected purchases: %v, got %v", test.expectedPurchases, purchases)
		}
		if !AlmostEqual(cash, test.expectedCashRemaining, 1e-7) {
			t.Errorf("expected remaning cash: %v, got %v", test.expectedCashRemaining, cash)
		}
	}
}

func TestRebalanceWithSelling(t *testing.T) {
	alloc1, err := LoadDesiredAllocations("testing/desiredAllocations_test1.yaml")
	if err != nil {
		t.Fatal("cannot continue testing TestRebalanceWithSelling: " + err.Error())
	}
	alloc2, err := LoadDesiredAllocations("testing/desiredAllocations_test2.yaml")
	if err != nil {
		t.Fatal("cannot continue testing TestRebalanceWithSelling: " + err.Error())
	}

	tests := []struct {
		cash                      float64
		desiredAllocations        *DesiredAllocations
		holdings                  map[string]float64
		prices                    map[string]float64
		expectedPurchasesAndSales map[string]int64
		expectedCashRemaining     float64
	}{
		{
			cash:               0.32,
			desiredAllocations: alloc1,
			holdings: map[string]float64{
				"DFAC": 66,
				"DFIC": 22,
				"DFEM": 12,
				"SWVXX": 4000,
			},
			prices: map[string]float64{
				"DFAC": 1,
				"DFIC": 1,
				"DFEM": 1,
				"SWVXX": 1,
			},
			expectedPurchasesAndSales: map[string]int64{
				"DFAC": -2,
				"DFIC": 5,
				"DFEM": -3,
			},
			expectedCashRemaining: 0.32,
		},
		{
			cash:               0.99,
			desiredAllocations: alloc1,
			holdings: map[string]float64{
				"DFAC": 170,
				"DFIC": 5,
				"DFEM": 25,
				"SWVXX": 4000,
			},
			prices: map[string]float64{
				"DFAC": 1,
				"DFIC": 1,
				"DFEM": 1,
				"SWVXX": 1,
			},
			expectedPurchasesAndSales: map[string]int64{
				"DFAC": -42,
				"DFIC": 49,
				"DFEM": -7,
			},
			expectedCashRemaining: 0.99,
		},
		{
			cash:               202.12,
			desiredAllocations: alloc2,
			holdings: map[string]float64{
				"VTI":   10,
				"VSIAX": 10,
				"VXUS":  10,
				"VWO":   10,
				"SWVXX": 4000,
			},
			prices: map[string]float64{
				"VTI":   10,
				"VSIAX": 10,
				"VXUS":  10,
				"VWO":   10,
				"SWVXX": 1,
			},
			expectedPurchasesAndSales: map[string]int64{
				"VTI":   20,
				"VSIAX": 2,
				"VXUS":  2,
				"VWO":   -4,
			},
			expectedCashRemaining: 2.12,
		},
	}

	for _, test := range tests {
		purchasesAndSales, cash := RebalanceWithSelling(test.cash, test.holdings, test.prices, test.desiredAllocations)
		if !reflect.DeepEqual(purchasesAndSales, test.expectedPurchasesAndSales) {
			t.Errorf("expected purchases and sales: %v, got %v", test.expectedPurchasesAndSales, purchasesAndSales)
		}
		if !AlmostEqual(cash, test.expectedCashRemaining, 1e-7) {
			t.Errorf("expected remaning cash: %v, got %v", test.expectedCashRemaining, cash)
		}
	}
}
