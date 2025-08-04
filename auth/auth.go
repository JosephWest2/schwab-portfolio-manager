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
	"golang.org/x/oauth2"
)

var port = os.Getenv("SCHWAB_OAUTH_SERVER_PORT")

var schwabEndpoint oauth2.Endpoint = oauth2.Endpoint{
	AuthURL:   "https://api.schwabapi.com/v1/oauth/authorize",
	TokenURL:  "https://api.schwabapi.com/v1/oauth/token",
	AuthStyle: oauth2.AuthStyleInHeader,
}

var OauthConfig *oauth2.Config = &oauth2.Config{
	RedirectURL:  fmt.Sprintf("https://127.0.0.1:%s/oauth2/callback", port),
	ClientID:     os.Getenv("SCHWAB_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("SCHWAB_OAUTH_CLIENT_SECRET"),
	Endpoint:     schwabEndpoint,
}

// server to handle the callback after authentication
func InitAuthCallbackServer(tokenChan chan *oauth2.Token) {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	mux.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No code in url", http.StatusBadRequest)
			return
		}
		fmt.Println("Auth code received: " + code)
		token, err := OauthConfig.Exchange(context.Background(), code)
		if err != nil {
			log.Fatal("Failed to get token: " + err.Error())
		}
		tokenChan <- token

		w.Write([]byte("Authentication successful, you can close this window and return to the application."))
	})

	err := server.ListenAndServeTLS("127.0.0.1.pem", "127.0.0.1-key.pem")
	if err != nil {
		log.Fatal(err)
	}
}

var ErrUnauthorized = errors.New("unauthorized")

type HookedTokenSource struct {
	src oauth2.TokenSource
	old *oauth2.Token
	mu  sync.Mutex
}

func NewHookedTokenSource(token *oauth2.Token) *HookedTokenSource {
	return &HookedTokenSource{
		src: OauthConfig.TokenSource(context.Background(), token),
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
	authCodeUrl := OauthConfig.AuthCodeURL("", oauth2.AccessTypeOnline)
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

func InitClient(tokenChan chan *oauth2.Token, stateChan chan string) *http.Client {
	client, err := CreateClientFromTokenFile()
	if err != nil {
		fmt.Println(err)
		client = Authenticate(tokenChan)
	}
	return client
}
