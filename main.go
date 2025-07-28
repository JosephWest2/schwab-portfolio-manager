package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/josephwest2/schwab-portfolio-manager/server"
	"golang.org/x/oauth2"
)

func main() {
	tokenChan := make(chan *oauth2.Token)
	go server.InitCallbackServer(tokenChan)

	authCodeUrl := server.OauthConfig.AuthCodeURL("", oauth2.AccessTypeOnline)
	fmt.Println("Authenticate here:\n" + authCodeUrl)

	token := <-tokenChan
	fmt.Println("Token received in main")
	client := server.OauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.schwabapi.com/trader/v1/accounts")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}
