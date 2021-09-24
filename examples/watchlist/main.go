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

	watchlistID, err := ig.CreateWatchlist(ctx, "example watchlist", []string{})
	checkErr(err)
	fmt.Printf("Watchlist created: %q\n", watchlistID)

	err = ig.AddToWatchlist(ctx, watchlistID, "CS.D.EURJPY.CFD.IP")
	checkErr(err)
	fmt.Println("Epic added")

	watchlist, err := ig.GetWatchlist(ctx, watchlistID)
	checkErr(err)
	fmt.Printf("Got watchlist: %v\n", watchlist)

	watchlists, err := ig.GetAllWatchlists(ctx)
	checkErr(err)
	for _, list := range *watchlists {
		fmt.Printf("Found watchlist: %v\n", list)
	}

	ig.DeleteFromWatchlist(ctx, watchlistID, "CS.D.EURJPY.CFD.IP")
	checkErr(err)
	fmt.Println("Epic deleted")

	ig.DeleteWatchlist(ctx, watchlistID)
	checkErr(err)
	fmt.Println("Watchlist deleted")
}
