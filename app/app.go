package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/josephwest2/schwab-portfolio-manager/auth"
	"github.com/josephwest2/schwab-portfolio-manager/balance"
	marketData "github.com/josephwest2/schwab-portfolio-manager/schwab/marketData"
	"github.com/josephwest2/schwab-portfolio-manager/schwab/trader"
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
		next:      PrintAccounts,
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

func MainOptions(a *App) AppHandler {
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
			return PrintAccounts
		case 2:
			return InvestCashSelectAccount
		case 3:
			return RebalanceAccounts
		case 4:
			return nil
		default:
			fmt.Println("invalid input")
		}
	}
}

func InvestCashSelectAccount(a *App) AppHandler {
	for i, acc := range a.accounts {
		fmt.Fprintf(os.Stdout, "\n#%v ********%v\n", i+1, acc.SecuritiesAccount.AccountNumber[len(acc.SecuritiesAccount.AccountNumber)-3:])
		fmt.Fprintf(os.Stdout, "Account value: $%v\n", acc.SecuritiesAccount.InitialBalances.AccountValue)
		fmt.Fprintf(os.Stdout, "Cash: $%v\n", acc.SecuritiesAccount.InitialBalances.CashBalance)
	}

	fmt.Println("\nSelect account to allocate cash to, 0 to cancel")

	for {
		var input int
		_, err := fmt.Scan(&input)
		if err != nil {
			fmt.Println("invalid input", err)
			continue
		}

		if input > 0 && input <= len(a.accounts) {
			return InvestCash(a, &a.accounts[input-1])
		}

		if input == 0 {
			return MainOptions
		}
	}
}

func InvestCash(a *App, account *Account) AppHandler {
	return func(a *App) AppHandler {
		cash := account.SecuritiesAccount.InitialBalances.CashBalance
		positions := account.SecuritiesAccount.Positions

		desiredAllocations, err := balance.LoadDesiredAllocations(balance.DesiredAllocationsFile)
		if err != nil {
			fmt.Println("failed to load desiredAllocations", err)
			return MainOptions
		}

		trackedHoldings := make(map[string]float64)

		fmt.Printf("Curent positions:\n")
		for _, pos := range positions {
			proportion := pos.MarketValue / account.SecuritiesAccount.InitialBalances.AccountValue
			fmt.Fprintf(os.Stdout, "%v: %v shares, $%.2f, %.2f%%\n", pos.Instrument.Symbol, pos.LongQuantity, pos.MarketValue, proportion*100)
			if desiredAllocations.Proportions[pos.Instrument.Symbol] == 0 && desiredAllocations.FixedCashAmounts[pos.Instrument.Symbol] == 0 {
				fmt.Fprintf(os.Stdout, "No desired allocation for %v\n", pos.Instrument.Symbol)
			} else {
				trackedHoldings[pos.Instrument.Symbol] = pos.LongQuantity
			}
			fmt.Println()
		}

		tickers := make([]string, 0, len(trackedHoldings))
		for k := range trackedHoldings {
			tickers = append(tickers, k)
		}
		for k := range desiredAllocations.Proportions {
			tickers = append(tickers, k)
		}
		for k := range desiredAllocations.FixedCashAmounts {
			tickers = append(tickers, k)
		}
		trackedPrices := GetAssetPrices(a, tickers)

		purchases, cash := balance.BalancePurchase(cash, trackedHoldings, trackedPrices, desiredAllocations)

		if len(purchases) == 0 {
			fmt.Println("Not enough cash to make any purchases")
			return MainOptions
		}
		fmt.Println("Optimal purchases:")
		for k, v := range purchases {
			fmt.Fprintf(os.Stdout, "%v: %v shares", k, v)
		}
		fmt.Fprintf(os.Stdout, "Resulting cash: $%v\n\n", cash)

		fmt.Println("type \"proceed\" to place the orders, anything else to cancel")

		for {
			var input string
			_, err := fmt.Scan(&input)
			if err != nil {
				fmt.Println("invalid input", err)
				continue
			}
			if input == "proceed" {
				return PlacePurchaseOrders(a, account, purchases)
			} else {
				return MainOptions
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
	fmt.Println(quoteResponse)

	prices := make(map[string]float64)
	for _, data := range quoteResponse {
		prices[data.Symbol] = data.Quote.AskPrice
	}
	return prices
}

func PlacePurchaseOrders(a *App, account *Account, orders map[string]int64) AppHandler {
	return func(a *App) AppHandler {
		for ticker, count := range orders {
			if count < 1 {
				continue
			}
			order := trader.Order{
				OrderType:         "MARKET",
				Session:           "NORMAL",
				Duration:          "DAY",
				OrderStrategyType: "SINGLE",
				OrderLegCollection: []trader.OrderLeg{
					{
						Instruction: "BUY",
						Quantity:    float64(count),
						Instrument: trader.Instrument{
							Symbol:    ticker,
							AssetType: "EQUITY",
						},
					},
				},
			}
			orderData, err := json.Marshal(order)
			fmt.Println("serialized order", string(orderData))
			return MainOptions
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
			if resp.StatusCode != 200 {
				log.Fatal("Failed to place order", resp.Body)
			}
		}

		return MainOptions
	}
}

func RebalanceAccounts(a *App) AppHandler {
	return MainOptions
}

func PrintAccounts(a *App) AppHandler {
	for _, acc := range a.accounts {
		fmt.Fprintf(os.Stdout, "******%v value: $%v\n", acc.SecuritiesAccount.AccountNumber[len(acc.SecuritiesAccount.AccountNumber)-3:], acc.SecuritiesAccount.InitialBalances.AccountValue)
	}

	return MainOptions
}

func (a *App) GetAccounts() ([]Account, error) {
	resp, err := a.client.Get(SchwabTraderApiAddress + "accounts/accountNumbers")
	if err != nil {
		log.Fatal(err)
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
