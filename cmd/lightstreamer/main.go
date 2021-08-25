package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/igmarkets"
	"os"
	"time"
)

func main() {
	var (
		url         = igmarkets.DemoAPIURL
		apiKey      = os.Getenv("IG_API_KEY")
		accountId   = os.Getenv("IG_ACCOUNT")
		identifier  = os.Getenv("IG_IDENTIFIER")
		password    = os.Getenv("IG_PASSWORD")
		httpTimeout = time.Second * 10
	)

	var ctx = context.Background()

	var ig = igmarkets.New(url, apiKey, accountId, identifier, password, httpTimeout)
	if err := ig.Login(ctx); err != nil {
		log.WithError(err).Fatal("Login failed")
	}

	for {
		tickChan := make(chan igmarkets.LightStreamerTick)
		err := ig.OpenLightStreamerSubscription(ctx, []string{"CS.D.BITCOIN.CFD.IP"}, tickChan)
		if err != nil {
			log.WithError(err).Error("Starting lightstreamer subscription failed")
		}

		for tick := range tickChan {
			log.Infof("Tick: %+v", tick)
		}

		log.Info("Server closed stream, restarting...")
	}
}
