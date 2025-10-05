package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/josephwest2/schwab-portfolio-manager/auth"
	"github.com/josephwest2/schwab-portfolio-manager/balance"
	"github.com/josephwest2/schwab-portfolio-manager/targetAllocation"
	"github.com/josephwest2/schwab-portfolio-manager/types/schwab/marketData"
	"github.com/josephwest2/schwab-portfolio-manager/types/schwab/trader"
	"golang.org/x/oauth2"
)

const SchwabTraderApiAddress = "https://api.schwabapi.com/trader/v1/"
const SchwabMarketDataApiAddress = "https://api.schwabapi.com/marketdata/v1/"

type Account struct {
	SecuritiesAccount trader.SecuritiesAccount
	AccountHashValue  string
}

type App struct {
	client    *http.Client
	tokenChan chan *oauth2.Token
	stateChan chan string
	accounts  []Account
	next      AppHandler
}

type AppHandler func(*App) AppHandler

func NewApp() *App {
	return &App{
		tokenChan: make(chan *oauth2.Token),
		next:      PrintAccountsHandler,
	}
}

func (a *App) Run() {
	go auth.InitAuthCallbackServer(a.tokenChan)
	a.client = auth.InitClient(a.tokenChan, a.stateChan)

	for a.accounts == nil {
		accounts, err := a.GetAccounts()
		if err != nil {
			if err == auth.ErrUnauthorized {
				a.client = auth.Authenticate(a.tokenChan)
			} else {
				log.Fatal(err)
			}
		}
		a.accounts = accounts
	}

	for a.next != nil {
		a.next = a.next(a)
	}

}

func MainOptionsHandler(a *App) AppHandler {
	fmt.Println("1. Print accounts")
	fmt.Println("2. Invest cash")
	fmt.Println("3. Rebalance accounts")
	fmt.Println("4. Exit")

	for {
		var input int
		_, err := fmt.Scan(&input)
		if err != nil {
			fmt.Println("invalid input", err)
			continue
		}

		switch input {
		case 1:
			return PrintAccountsHandler
		case 2:
			return InvestCashSelectAccountHandler
		case 3:
			return RebalanceAccountsSelectAccountHandler
		case 4:
			return nil
		default:
			fmt.Println("invalid input")
		}
	}
}

func RebalanceAccountsSelectAccountHandler(a *App) AppHandler {
	PrintAccounts(a.accounts)
	fmt.Println("\nSelect account to rebalance, q to cancel")

	for {
		var input int
		_, err := fmt.Scan(&input)
		if err != nil {
			return MainOptionsHandler
		}

		if input >= 0 && input < len(a.accounts) {
			return RebalanceAccountHandlerFunc(a, &a.accounts[input])
		}
		fmt.Println("invalid input")
	}
}

func PrintAccounts(accounts []Account) {
	for i, acc := range accounts {
		fmt.Fprintf(os.Stdout, "\n#%v ********%v\n", i, acc.SecuritiesAccount.AccountNumber[len(acc.SecuritiesAccount.AccountNumber)-3:])
		fmt.Fprintf(os.Stdout, "Account value: $%v\n", acc.SecuritiesAccount.InitialBalances.AccountValue)
		fmt.Fprintf(os.Stdout, "Cash: $%.2f\n\n", acc.SecuritiesAccount.InitialBalances.CashBalance)
	}
}

func InvestCashSelectAccountHandler(a *App) AppHandler {
	PrintAccounts(a.accounts)

	fmt.Println("\nSelect account to allocate cash to, q to cancel")

	for {
		var input int
		_, err := fmt.Scan(&input)
		if err != nil {
			return MainOptionsHandler
		}

		if input >= 0 && input < len(a.accounts) {
			return InvestCashHandlerFunc(a, &a.accounts[input])
		}
		fmt.Println("invalid input")
	}
}

// gives holdings that are tracked by desired allocations
func GetTrackedHoldings(positions []trader.Position, targetAllocation targetAllocation.TargetAllocation) map[string]float64 {
	trackedHoldings := make(map[string]float64)
	for _, pos := range positions {
		if targetAllocation[pos.Instrument.Symbol].Proportion == 0 && targetAllocation[pos.Instrument.Symbol].FixedCashValue == 0 {
			continue
		}
		trackedHoldings[pos.Instrument.Symbol] = pos.LongQuantity
	}
	for ticker := range targetAllocation {
		if _, ok := trackedHoldings[ticker]; !ok {
			trackedHoldings[ticker] = 0
		}
	}
	return trackedHoldings
}

func PrintCurrentPositions(positions []trader.Position, accountValue float64, targetAllocation targetAllocation.TargetAllocation) {
	fmt.Printf("Curent positions:\n")
	for _, pos := range positions {
		proportion := pos.MarketValue / accountValue
		fmt.Fprintf(os.Stdout, "%v: %v shares, $%.2f, %.2f%%\n", pos.Instrument.Symbol, pos.LongQuantity, pos.MarketValue, proportion*100)
		if targetAllocation[pos.Instrument.Symbol].Proportion == 0 && targetAllocation[pos.Instrument.Symbol].FixedCashValue == 0 {
			fmt.Fprintf(os.Stdout, "No desired allocation for %v, skipping inclusion in further calculations\n", pos.Instrument.Symbol)
		}
		fmt.Println()
	}
}

func InvestCashHandlerFunc(a *App, account *Account) AppHandler {
	return func(a *App) AppHandler {

		targetAllocations, err := targetAllocation.LoadTargetAllocations(targetAllocation.TargetAllocationFile)
		if err != nil {
			fmt.Println("failed to load targetAllocations", err)
			return MainOptionsHandler
		}

		targetAllocation := targetAllocations[account.SecuritiesAccount.AccountNumber[len(account.SecuritiesAccount.AccountNumber)-3:]]

		trackedHoldings := GetTrackedHoldings(account.SecuritiesAccount.Positions, targetAllocation)
		PrintCurrentPositions(account.SecuritiesAccount.Positions, account.SecuritiesAccount.InitialBalances.AccountValue, targetAllocation)

		tickers := slices.Collect(maps.Keys(trackedHoldings))
		for ticker := range targetAllocations {
			if _, ok := trackedHoldings[ticker]; !ok {
				tickers = append(tickers, ticker)
			}
		}
		trackedPrices := GetAssetPrices(a, tickers)

		purchases, cash := balance.BalancePurchase(account.SecuritiesAccount.InitialBalances.CashBalance, trackedHoldings, trackedPrices, targetAllocation)

		if len(purchases) == 0 {
			fmt.Println("Not enough cash to make any purchases")
			return MainOptionsHandler
		}
		fmt.Println("Optimal purchases:")
		for k, v := range purchases {
			fmt.Fprintf(os.Stdout, "%v: %v shares\n", k, v)
		}
		fmt.Fprintf(os.Stdout, "Resulting cash: $%.2f\n\n", cash)

		fmt.Println("type \"proceed\" to place the orders, anything else to cancel")

		for {
			var input string
			_, err := fmt.Scan(&input)
			if err != nil {
				fmt.Println("invalid input", err)
				continue
			}
			if input == "proceed" {
				return PlaceBuyOrderHandlerFunc(a, account, purchases)
			} else {
				return MainOptionsHandler
			}
		}
	}
}

func GetAssetPrices(a *App, tickers []string) map[string]float64 {
	addr := fmt.Sprintf(SchwabMarketDataApiAddress+"quotes?symbols=%s", strings.Join(tickers, "%2C")) + "&fields=quote&indicative=false"
	resp, err := a.client.Get(addr)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal(resp.Status)
	}

	quoteResponse := make(marketData.QuoteResponse)
	err = json.NewDecoder(resp.Body).Decode(&quoteResponse)
	if err != nil {
		log.Fatal(err)
	}

	prices := make(map[string]float64)
	for _, data := range quoteResponse {
		prices[data.Symbol] = data.Quote.LastPrice
	}
	return prices
}

func PlaceTriggerOrderHandlerFunc(a *App, account *Account, orders map[string]float64) AppHandler {
	return func(a *App) AppHandler {
		fmt.Println("placing trigger order")
		order := trader.Order{
			OrderType:          "MARKET",
			Session:            "NORMAL",
			Duration:           "DAY",
			OrderStrategyType:  "TRIGGER",
			OrderLegCollection: make([]trader.OrderLeg, 0),
			ChildOrderStrategies: []trader.Order{
				{
					OrderType:          "MARKET",
					Session:            "NORMAL",
					Duration:           "DAY",
					OrderStrategyType:  "SINGLE",
					OrderLegCollection: make([]trader.OrderLeg, 0),
				},
			},
		}
		for ticker, count := range orders {
			if count < 0 {
				order.OrderLegCollection = append(order.OrderLegCollection, trader.OrderLeg{
					Instruction: "SELL",
					Quantity:    -count,
					Instrument: trader.Instrument{
						Symbol:    ticker,
						AssetType: "EQUITY",
					},
				})
			} else if count > 0 {
				order.ChildOrderStrategies[0].OrderLegCollection = append(order.ChildOrderStrategies[0].OrderLegCollection, trader.OrderLeg{
					Instruction: "BUY",
					Quantity:    count,
					Instrument: trader.Instrument{
						Symbol:    ticker,
						AssetType: "EQUITY",
					},
				})
			}
		}

		orderData, err := json.Marshal(order)
		fmt.Println("serialized order", string(orderData))
		if err != nil {
			log.Fatal(err)
		}
		resp, err := a.client.Post(
			SchwabTraderApiAddress+fmt.Sprintf("accounts/%v/orders", account.AccountHashValue),
			"application/json",
			bytes.NewBuffer(orderData),
		)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != 201 {

			log.Fatal("Failed to place order", string(respBody))
		}
		fmt.Println("\n\n Order Placed \n\n", string(respBody))
		return MainOptionsHandler
	}

}

func PlaceBuyOrderHandlerFunc(a *App, account *Account, orders map[string]float64) AppHandler {
	return func(a *App) AppHandler {
		fmt.Println("placing buy order")
		order := trader.Order{
			OrderType:          "MARKET",
			Session:            "NORMAL",
			Cancelable:         true,
			Duration:           "DAY",
			OrderStrategyType:  "SINGLE",
			OrderLegCollection: make([]trader.OrderLeg, 0),
		}
		for ticker, count := range orders {
			if count < 1 {
				continue
			}
			order.OrderLegCollection = append(order.OrderLegCollection, trader.OrderLeg{
				Instruction: "BUY",
				Quantity:    count,
				Instrument: trader.Instrument{
					Symbol:    ticker,
					AssetType: "EQUITY",
				},
			})
		}
		orderData, err := json.Marshal(order)
		fmt.Println("serialized order", string(orderData))
		if err != nil {
			log.Fatal(err)
		}
		resp, err := a.client.Post(
			SchwabTraderApiAddress+fmt.Sprintf("accounts/%v/orders", account.AccountHashValue),
			"application/json",
			bytes.NewBuffer(orderData),
		)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != 201 {
			log.Fatal("Failed to place order", string(respBody))
		}

		fmt.Println("\n\n Order Placed\n\n", string(respBody))
		return MainOptionsHandler
	}
}

func RebalanceAccountHandlerFunc(a *App, account *Account) AppHandler {
	return func(a *App) AppHandler {

		targetAllocations, err := targetAllocation.LoadTargetAllocations(targetAllocation.TargetAllocationFile)
		if err != nil {
			fmt.Println("failed to load targetAllocations", err)
			return MainOptionsHandler
		}

		targetAllocation := targetAllocations[account.SecuritiesAccount.AccountNumber[len(account.SecuritiesAccount.AccountNumber)-3:]]

		trackedHoldings := GetTrackedHoldings(account.SecuritiesAccount.Positions, targetAllocation)
		PrintCurrentPositions(account.SecuritiesAccount.Positions, account.SecuritiesAccount.InitialBalances.AccountValue, targetAllocation)

		tickers := slices.Collect(maps.Keys(trackedHoldings))
		for ticker := range targetAllocation {
			if _, ok := trackedHoldings[ticker]; !ok {
				tickers = append(tickers, ticker)
			}
		}
		trackedPrices := GetAssetPrices(a, tickers)

		orders, cash := balance.RebalanceWithSelling(account.SecuritiesAccount.InitialBalances.CashBalance, trackedHoldings, trackedPrices, targetAllocation)
		purchases := make(map[string]float64)
		sales := make(map[string]float64)
		for k, v := range orders {
			if v > 0 {
				purchases[k] = v
			} else if v < 0 {
				sales[k] = v
			}
		}
		if len(purchases) == 0 {
			fmt.Println("Portfolio is already optimally balanced")
			return MainOptionsHandler
		}
		fmt.Println("Optimal sales:")
		for k, v := range sales {
			fmt.Fprintf(os.Stdout, "%v: %v shares\n", k, v)
		}
		fmt.Println("Optimal purchases:")
		for k, v := range purchases {
			fmt.Fprintf(os.Stdout, "%v: %v shares\n", k, v)
		}
		fmt.Fprintf(os.Stdout, "Resulting cash: $%.2f\n\n", cash)

		fmt.Println("type \"proceed\" to place the orders, anything else to cancel")

		for {
			var input string
			_, err := fmt.Scan(&input)
			if err != nil {
				fmt.Println("invalid input", err)
				continue
			}
			if input == "proceed" {
				if len(sales) == 0 {
					return PlaceBuyOrderHandlerFunc(a, account, purchases)
				}
				return PlaceTriggerOrderHandlerFunc(a, account, orders)
			} else {
				return MainOptionsHandler
			}
		}
	}
}

func PrintAccountsHandler(a *App) AppHandler {
	PrintAccounts(a.accounts)
	return MainOptionsHandler
}

func (a *App) GetAccounts() ([]Account, error) {
	resp, err := a.client.Get(SchwabTraderApiAddress + "accounts/accountNumbers")
	if err != nil {
		ue := err.(*url.Error)
		if _, ok := ue.Err.(*oauth2.RetrieveError); ok {
			return nil, auth.ErrUnauthorized
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, auth.ErrUnauthorized
	}

	var accounts trader.AccountNumbersResponse
	err = json.NewDecoder(resp.Body).Decode(&accounts)
	if err != nil {
		log.Fatal(err)
	}

	var res []Account
	for _, acc := range accounts {
		resp, err := a.client.Get(SchwabTraderApiAddress + "accounts/" + acc.HashValue + "?fields=positions")
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, auth.ErrUnauthorized
		}

		var securitiesAccount trader.AccountResponse
		err = json.NewDecoder(resp.Body).Decode(&securitiesAccount)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, Account{securitiesAccount.SecuritiesAccount, acc.HashValue})
	}

	return res, nil
}
