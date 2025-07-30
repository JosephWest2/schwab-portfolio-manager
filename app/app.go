package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/josephwest2/schwab-portfolio-manager/balance"
	"github.com/josephwest2/schwab-portfolio-manager/encryption"
	"github.com/josephwest2/schwab-portfolio-manager/schwabTypes"
	"github.com/josephwest2/schwab-portfolio-manager/server"
	"golang.org/x/oauth2"
)

var errUnauthorized = errors.New("unauthorized")

type HookedTokenSource struct {
	src oauth2.TokenSource
	old *oauth2.Token
	mu  sync.Mutex
}

func NewHookedTokenSource(token *oauth2.Token) *HookedTokenSource {
	return &HookedTokenSource{
		src: server.OauthConfig.TokenSource(context.Background(), token),
	}
}

func (ts *HookedTokenSource) Token() (*oauth2.Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	token, err := ts.src.Token()
	if err != nil {
		return nil, err
	}

	if ts.old == nil || ts.old != token {
		fmt.Println("Token changed")
		WriteTokenToFile(token)
		ts.old = token
	}

	return token, nil
}

// serialize, encrypt, and write token to file
func WriteTokenToFile(token *oauth2.Token) {
	var tokenData []byte

	tokenData, err := json.Marshal(token)
	if err != nil {
		log.Fatal(err)
	}
	encryption.EncryptToFile(tokenData, encryption.EncryptedTokenFilename)
}

func ReadTokenFromFile() (*oauth2.Token, error) {
	tokenData, err := encryption.DecryptFromFile(encryption.EncryptedTokenFilename)
	if err != nil {
		return nil, err
	}

	var token *oauth2.Token
	err = json.Unmarshal(tokenData, &token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func Authenticate(tokenChan chan *oauth2.Token) *http.Client {
	authCodeUrl := server.OauthConfig.AuthCodeURL("", oauth2.AccessTypeOnline)
	fmt.Fprintf(os.Stdout, "\nAuthenticate here:\n\n%v\n\n", authCodeUrl)

	token := <-tokenChan

	hookedTokenSource := NewHookedTokenSource(token)
	return oauth2.NewClient(context.Background(), hookedTokenSource)
}

func CreateClientFromTokenFile() (*http.Client, error) {
	token, err := ReadTokenFromFile()
	if err != nil {
		return nil, err
	}

	hookedTokenSource := NewHookedTokenSource(token)
	return oauth2.NewClient(context.Background(), hookedTokenSource), nil
}

func InitClient(tokenChan chan *oauth2.Token) *http.Client {
	client, err := CreateClientFromTokenFile()
	if err != nil {
		fmt.Println(err)
		client = Authenticate(tokenChan)
	}
	return client
}

type Account struct {
	SecuritiesAccount schwabTypes.SecuritiesAccount
	AccountHashValue  string
}

type App struct {
	client    *http.Client
	tokenChan chan *oauth2.Token
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
	go server.InitAuthCallbackServer(a.tokenChan)
	a.client = InitClient(a.tokenChan)

	for a.accounts == nil {
		accounts, err := a.GetAccounts()
		if err != nil {
			if err == errUnauthorized {
				a.client = Authenticate(a.tokenChan)
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
		fmt.Fprintf(os.Stdout, "\n#%v\n\n", i+1)
		fmt.Fprintf(os.Stdout, "********%v\n", acc.SecuritiesAccount.AccountNumber[len(acc.SecuritiesAccount.AccountNumber)-3:])
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

		holdings := make(map[string]float64)
		prices := make(map[string]float64)

		fmt.Printf("Curent positions:\n")
		for _, pos := range positions {
			proportion := pos.LongQuantity / account.SecuritiesAccount.InitialBalances.AccountValue
			fmt.Fprintf(os.Stdout, "%v: %v shares, $%v, %v%%\n", pos.Instrument.Symbol, pos.LongQuantity, pos.MarketValue, proportion*100)
			if desiredAllocations[pos.Instrument.Symbol] == 0 {
				fmt.Fprintf(os.Stdout, "No desired allocation for %v\n", pos.Instrument.Symbol)
			}
			holdings[pos.Instrument.Symbol] = pos.LongQuantity
			prices[pos.Instrument.Symbol] = pos.CurrentDayCost

			fmt.Println()
		}

		fmt.Println("Available cash: $", cash)
		purchases, cash := balance.BalancePurchase(cash, holdings, prices, desiredAllocations)

		fmt.Println("Optimal purchases:")
		for k, v := range purchases {
			fmt.Fprintf(os.Stdout, "%v: %v shares", k, v)
		}
		fmt.Fprintf(os.Stdout, "Resulting cash: $%v\n\n", cash)

		fmt.Println("type \"proceed\" to execute the trades, anything else to cancel")

		for {
			var input string
			_, err := fmt.Scan(&input)
			if err != nil {
				fmt.Println("invalid input", err)
				continue
			}

			if input == "proceed" {
				fmt.Println("This would have executed trades :)")
				return MainOptions
			}
		}
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
	resp, err := a.client.Get("https://api.schwabapi.com/trader/v1/accounts/accountNumbers")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, errUnauthorized
	}

	var accounts schwabTypes.AccountNumbersResponse
	err = json.NewDecoder(resp.Body).Decode(&accounts)
	if err != nil {
		log.Fatal(err)
	}

	var res []Account
	for _, acc := range accounts {
		resp, err := a.client.Get("https://api.schwabapi.com/trader/v1/accounts/" + acc.HashValue)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, errUnauthorized
		}

		var securitiesAccount schwabTypes.AccountResponse
		err = json.NewDecoder(resp.Body).Decode(&securitiesAccount)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, Account{securitiesAccount.SecuritiesAccount, acc.HashValue})
	}

	return res, nil
}
