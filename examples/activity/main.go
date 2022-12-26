package main

import (
	"context"
	"fmt"
	"github.com/lfritz/env"
	log "github.com/sirupsen/logrus"
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

	from := time.Now().AddDate(0, 0, -30) // 30 days ago
	to := time.Now()

	activity, err := ig.GetActivity(ctx, from, to)
	checkErr(err)

	for _, activity := range activity.Activities {
		fmt.Printf("Activity: dealID: %s\n", activity.DealID)
		for _, action := range activity.Details.Actions {
			fmt.Printf("\t\tAction: ActionType=%s DealId=%s\n", action.ActionType, action.AffectedDealId)
		}
	}
}
