package targetAllocation_test

import (
	"reflect"
	"testing"

	"github.com/josephwest2/schwab-portfolio-manager/targetAllocation"
)
func TestLoadAllocations(t *testing.T) {
	tests := []struct {
		filepath string
		expected targetAllocation.TargetAllocations
		wantErr  bool
	}{
		{
			filepath: "testing/targetAllocation_targetAllocation_test1.yaml",
			expected: targetAllocation.TargetAllocations{
				"global": targetAllocation.TargetAllocation{
					"SWVXX": {
						FixedCashValue: 4000,
						Proportion:     0.0,
					},
					"DFAC": {
						FixedCashValue: 0.0,
						Proportion:     0.64,
					},
					"DFIC": {
						FixedCashValue: 0.0,
						Proportion:     0.27,
					},
					"DFEM": {
						FixedCashValue: 0,
						Proportion:     0.09,
					},
				},
			},
			wantErr: false,
		},
		{
			filepath: "testing/targetAllocation_targetAllocation_test2.yaml",
			expected: targetAllocation.TargetAllocations{
				"567": targetAllocation.TargetAllocation{
					"VTI": {
						FixedCashValue: 0.0,
						Proportion:     0.50,
					},
					"VSAIX": {
						FixedCashValue: 0.0,
						Proportion:     0.20,
					},
					"VXUS":{
						FixedCashValue: 0,
						Proportion:     0.20,
					},
					"VWO": {
						FixedCashValue: 0.0,
						Proportion:     0.10,
					},
				},
			},
			wantErr: false,
		},
		{
			// Allocation sums to 0.999 not 1
			filepath: "testing/targetAllocation_targetAllocation_test3.yaml",
			expected: nil,
			wantErr:  true,
		},
		{
			// Allocation sums to 1.001 not 1
			filepath: "testing/targetAllocation_targetAllocation_test4.yaml",
			expected: nil,
			wantErr:  true,
		},
		{
			filepath: "testing/targetAllocation_targetAllocation_test5.yaml",
			expected: targetAllocation.TargetAllocations{
				"global": map[targetAllocation.Ticker]targetAllocation.Allocation{
					"VTI": {
						FixedCashValue: 0.0,
						Proportion:     1.0,
					},
				},
				"123": map[targetAllocation.Ticker]targetAllocation.Allocation{
					"DFAC": {
						FixedCashValue: 0.0,
						Proportion:     0.64,
					},
					"DFIC": {
						FixedCashValue: 0.0,
						Proportion:     0.27,
					},
					"DFEM": {
						FixedCashValue: 0,
						Proportion:     0.09,
					},
					"SWVXX": {
						FixedCashValue: 2000,
						Proportion:     0.0,
					},
				},
				"456": map[targetAllocation.Ticker]targetAllocation.Allocation{
					"VTI": {
						FixedCashValue: 0.0,
						Proportion:     0.50,
					},
					"VSAIX": {
						FixedCashValue: 0.0,
						Proportion:     0.20,
					},
					"VXUS": {
						FixedCashValue: 0,
						Proportion:     0.20,
					},
					"VWO": {
						FixedCashValue: 0.0,
						Proportion:     0.10,
					},
				},
			},
			wantErr: false,
		},
	}
	for i, test := range tests {
		allocations, err := targetAllocation.LoadTargetAllocations(test.filepath)
		if test.wantErr && err == nil {
			t.Errorf("expected error on %v, got no error", test)
		}
		equal := reflect.DeepEqual(allocations, test.expected)
		if !equal {
			t.Errorf("expected %v, got %v, on test index %v, error: %v", test.expected, allocations, i, err.Error())
		}
	}
}
