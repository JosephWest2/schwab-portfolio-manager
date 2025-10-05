package targetAllocation

import (
	"errors"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/josephwest2/schwab-portfolio-manager/util"
)

var TargetAllocationFile = "targetAllocation.yaml"

// last 3 digits of account or 'global' for cross account allocation
type AccountIdentifier = string

type Ticker = string

type Allocation struct {
	Proportion     float64 `yaml:"proportion"`
	FixedCashValue float64 `yaml:"fixedCashValue"`
}

type TargetAllocation = map[Ticker]Allocation

type TargetAllocations map[AccountIdentifier]TargetAllocation

func (da *TargetAllocations) Tickers(account AccountIdentifier) []Ticker {
	tickers := make([]Ticker, 0, len((*da)[account]))
	for ticker := range (*da)[account] {
		tickers = append(tickers, ticker)
	}
	return tickers
}

func LoadTargetAllocations(filepath string) (TargetAllocations, error) {

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, errors.New("failed to read allocation file: " + err.Error())
	}

	result := make(TargetAllocations)

	err = yaml.Unmarshal(data, &result)
	if err != nil {
		return nil, errors.New("failed to parse allocation file: " + err.Error())
	}

	for _, accountAllocation := range result {
		sum := 0.0
		for _, tickerAllocData := range accountAllocation {
			sum += tickerAllocData.Proportion
		}
		if !util.AlmostEqual(sum, 1.0, 1e-7) {
			return nil, errors.New("allocation proportions do not sum to 1.0")
		}
	}

	return result, err
}
