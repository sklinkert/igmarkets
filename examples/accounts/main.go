package main

import (
	"context"
	"fmt"
	"github.com/lfritz/env"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/igmarkets"
	"os"
)

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var conf struct {
	igAPIURL     string
	igIdentifier string
	igAPIKey     string
	igPassword   string
	igAccountID  string
	instrument   string
}

func main() {
	var e = env.New()
	e.OptionalString("INSTRUMENT", &conf.instrument, "CS.D.EURUSD.MINI.IP", "instrument to trade")
	e.OptionalString("IG_API_URL", &conf.igAPIURL, igmarkets.DemoAPIURL, "IG API URL")
	e.OptionalString("IG_IDENTIFIER", &conf.igIdentifier, "", "IG Identifier")
	e.OptionalString("IG_API_KEY", &conf.igAPIKey, "", "IG API key")
	e.OptionalString("IG_PASSWORD", &conf.igPassword, "", "IG password")
	e.OptionalString("IG_ACCOUNT", &conf.igAccountID, "", "IG account ID")
	if err := e.Load(); err != nil {
		log.WithError(err).Fatal("env loading failed")
	}

	var ctx = context.Background()
	ig := igmarkets.New(conf.igAPIURL, conf.igAPIKey, conf.igAccountID, conf.igIdentifier, conf.igPassword)
	err := ig.Login(ctx)
	checkErr(err)

	accounts, err := ig.GetAccounts(ctx)
	checkErr(err)

	for _, account := range accounts.Accounts {
		fmt.Printf("Account: %q\n", account.AccountId)
		fmt.Printf("Type: %q\n", account.AccountType)
		fmt.Printf("Balance: %f\n", account.Balance.Balance)
		fmt.Printf("Available: %f\n", account.Balance.Available)
		fmt.Printf("ProfitLoss: %f\n", account.Balance.ProfitLoss)
		fmt.Printf("Deposit: %f\n", account.Balance.Deposit)
		fmt.Printf("Status: %q\n", account.Status)
	}

	accountPreferences, err := ig.GetAccountPreferences(ctx)
	checkErr(err)
	fmt.Printf("TralingStopLossEnabled: %t\n", accountPreferences.TrailingStopsEnabled)
}
