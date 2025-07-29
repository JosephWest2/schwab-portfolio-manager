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
}

func NewApp() *App {
	return &App{
		tokenChan: make(chan *oauth2.Token),
	}
}

func (a *App) Run() {
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
