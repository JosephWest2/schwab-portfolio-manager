package auth

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
	"github.com/josephwest2/schwab-portfolio-manager/server"
	"golang.org/x/oauth2"
)

var ErrUnauthorized = errors.New("unauthorized")

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