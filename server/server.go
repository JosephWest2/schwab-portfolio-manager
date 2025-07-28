package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
	})

	fmt.Println("Auth callback server starting")
	err := http.ListenAndServeTLS(":"+port, "127.0.0.1.pem", "127.0.0.1-key.pem", mux)
	if err != nil {
		log.Fatal(err)
	}
}
