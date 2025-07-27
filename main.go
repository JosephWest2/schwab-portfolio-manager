package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"slices"

	"github.com/goccy/go-yaml"
	"golang.org/x/oauth2"
	"gonum.org/v1/gonum/optimize"
)

func main() {
	token := make(chan *oauth2.Token)
	go initServer(token)

	authCodeUrl := oauthConfig.AuthCodeURL("", oauth2.AccessTypeOnline)
	fmt.Println("Authenticate here:\n" + authCodeUrl)

	t := <-token
	fmt.Println("Token received in main")
	client := oauthConfig.Client(context.Background(), t)
	resp, err := client.Get("https://api.schwabapi.com/trader/v1/accounts")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}

var schwabEndpoint oauth2.Endpoint = oauth2.Endpoint{
	AuthURL:   "https://api.schwabapi.com/v1/oauth/authorize",
	TokenURL:  "https://api.schwabapi.com/v1/oauth/token",
	AuthStyle: oauth2.AuthStyleInHeader,
}

var oauthConfig *oauth2.Config = &oauth2.Config{
	RedirectURL:  "https://localhost:34970/oauth2/callback",
	ClientID:     os.Getenv("SCHWAB_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("SCHWAB_OUATH_CLIENT_SECRET"),
	Endpoint:     schwabEndpoint,
}

func initServer(token chan *oauth2.Token) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code in url", http.StatusBadRequest)
			return
		}
		fmt.Println("Auth code received: " + code)
		t, err := oauthConfig.Exchange(context.Background(), code)
		if err != nil {
			log.Fatal("Failed to get token: " + err.Error())
		}
		token <- t
	})
	err := http.ListenAndServe(":34970", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func loadDesiredAllocations(filepath string) (map[string]float64, error) {

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
	if !almostEqual(sum, 1.0, 1e-7) {
		return nil, errors.New("allocations do not sum to 1.0")
	}

	return result, err
}

type Holding struct {
	Ticker string
	Count  int64
}

// sort by deviation from expected proportion
func purchasePriorityFunc(totalHoldingsValue float64, prices map[string]float64, desiredAllocations map[string]float64) func(a, b Holding) int {
	return func(a, b Holding) int {
		va := float64(a.Count) * prices[a.Ticker]
		vb := float64(b.Count) * prices[b.Ticker]
		da := va/totalHoldingsValue - desiredAllocations[a.Ticker]
		db := vb/totalHoldingsValue - desiredAllocations[b.Ticker]
		return cmp.Compare(da, db)
	}
}

// returns purchases to be made and remaining cash
func balancePurchase(cash float64, holdings map[string]int64, prices map[string]float64, desiredAllocations map[string]float64) (map[string]int64, float64) {
	holdingsSlice := make([]Holding, 0, len(holdings))
	for k, v := range holdings {
		holdingsSlice = append(holdingsSlice, Holding{k, v})
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
		slices.SortFunc(holdingsSlice, purchasePriorityFunc(totalHoldingsValue, prices, desiredAllocations))
		// buy the asset with the lowest deviation that can be afforded
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

// order of prices and desired Allocations should be alligned with the order of holdings
func objectiveFuncDeviation(prices []float64, desiredAllocations []float64) func([]float64) float64 {
	return func(holdings []float64) float64 {
		totalHoldingsValue := 0.0
		for i, h := range holdings {
			totalHoldingsValue += h * prices[i]
		}
		deviation := 0.0
		for i, h := range holdings {
			deviation += math.Pow(h*prices[i]/totalHoldingsValue-desiredAllocations[i], 2)
		}
		return deviation
	}
}

// returns purchases and sales to be made and remaining cash
func rebalanceWithSelling(cash float64, holdings map[string]int64, prices map[string]float64, desiredAllocations map[string]float64) (map[string]int64, float64) {
	if len(holdings) != len(prices) || len(holdings) != len(desiredAllocations) {
		log.Fatal("holdings, prices, and desired allocations do not have the same length")
	}

	pricesSlice := make([]float64, 0, len(prices))
	desiredAllocationsSlice := make([]float64, 0, len(desiredAllocations))
	holdingsSlice := make([]float64, 0, len(holdings))
	tickersSlice := make([]string, 0, len(holdings))

	for ticker, holding := range holdings {
		price := prices[ticker]
		desiredAllocation := desiredAllocations[ticker]

		pricesSlice = append(pricesSlice, price)
		desiredAllocationsSlice = append(desiredAllocationsSlice, desiredAllocation)
		holdingsSlice = append(holdingsSlice, float64(holding))
		tickersSlice = append(tickersSlice, ticker)
	}
	problem := optimize.Problem{
		Func: objectiveFuncDeviation(pricesSlice, desiredAllocationsSlice),
	}
	result, err := optimize.Minimize(problem, holdingsSlice, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
	purchasesAndSales := make(map[string]int64, 0)
	for i, ticker := range tickersSlice {
		newHolding := int64(result.X[i])
		purchasesAndSales[ticker] = newHolding - holdings[ticker]
		cash -= float64(purchasesAndSales[ticker]) * prices[ticker]
		holdings[ticker] = newHolding
	}
	purchases, cash := balancePurchase(cash, holdings, prices, desiredAllocations)
	for k, v := range purchases {
		purchasesAndSales[k] += v
	}
	return purchasesAndSales, cash
}
