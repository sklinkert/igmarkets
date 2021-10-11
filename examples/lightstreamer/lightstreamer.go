package main

import (
	"context"
	"github.com/lfritz/env"
	log "github.com/sirupsen/logrus"
	"github.com/sklinkert/igmarkets"
)

var conf struct {
	igAPIURL     string
	igIdentifier string
	igAPIKey     string
	igPassword   string
	igAccountID  string
	epics        []string
}

func main() {
	var e = env.New()
	e.OptionalList("EPICS", &conf.epics, ",", []string{"CS.D.EURUSD.MINI.IP", "CS.D.BITCOIN.CFD.IP"}, "Instruments to subscribe")
	e.OptionalString("IG_API_URL", &conf.igAPIURL, igmarkets.DemoAPIURL, "IG API URL")
	e.OptionalString("IG_IDENTIFIER", &conf.igIdentifier, "", "IG Identifier")
	e.OptionalString("IG_API_KEY", &conf.igAPIKey, "", "IG API key")
	e.OptionalString("IG_PASSWORD", &conf.igPassword, "", "IG password")
	e.OptionalString("IG_ACCOUNT", &conf.igAccountID, "", "IG account ID")
	if err := e.Load(); err != nil {
		log.WithError(err).Fatal("env loading failed")
	}

	var ctx = context.Background()

	for {
		igHandle := igmarkets.New(conf.igAPIURL, conf.igAPIKey, conf.igAccountID, conf.igIdentifier, conf.igPassword)
		if err := igHandle.Login(ctx); err != nil {
			log.WithError(err).Error("new fialed")
			return
		}

		tickChan := make(chan igmarkets.LightStreamerTick)
		err := igHandle.OpenLightStreamerSubscription(ctx, conf.epics, tickChan)
		if err != nil {
			log.WithError(err).Error("open stream fialed")
		}

		for tick := range tickChan {
			log.Infof("tick: %+v", tick)
		}

		log.Infof("Server closed stream, restarting...")
	}
}
