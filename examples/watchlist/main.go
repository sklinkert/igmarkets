package main

import (
	"fmt"
	"github.com/sklinkert/igmarkets"
	"os"
	"time"
)

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	apiKey := ""
	accountID := ""
	igIdent := ""
	igPassword := ""

	ig := igmarkets.New(igmarkets.DemoAPIURL, apiKey, accountID, igIdent, igPassword, time.Duration(5*time.Second))
	err := ig.Login()
	checkErr(err)

	watchlistID, err := ig.CreateWatchlist("example watchlist", []string{})
	checkErr(err)
	fmt.Printf("Watchlist created: %q\n", watchlistID)

	err = ig.AddToWatchlist(watchlistID, "CS.D.EURJPY.CFD.IP")
	checkErr(err)
	fmt.Println("Epic added")

	watchlist, err := ig.GetWatchlist(watchlistID)
	checkErr(err)
	fmt.Printf("Got watchlist: %v\n", watchlist)

	watchlists, err := ig.GetAllWatchlists()
	checkErr(err)
	for _, list := range *watchlists {
		fmt.Printf("Found watchlist: %v\n", list)
	}

	ig.DeleteFromWatchlist(watchlistID, "CS.D.EURJPY.CFD.IP")
	checkErr(err)
	fmt.Println("Epic deleted")

	ig.DeleteWatchlist(watchlistID)
	checkErr(err)
	fmt.Println("Watchlist deleted")
}
