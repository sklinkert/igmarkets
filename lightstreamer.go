package igmarkets

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type LightStreamerConnection struct {
	ig                       *IGMarkets
	sessionID                string
	serverHostname           string
	sessionBindTime          time.Time
	wsConn                   *websocket.Conn
	nextSubscriptionId       int
	heartbeatTicker          *time.Ticker
	ctx                      context.Context
	cancelFunc               func()
	lastError                error
	closeRequested           atomic.Bool
	marketSubscriptions      []*subscription[MarketTick]
	chartTickSubscriptions   []*subscription[ChartTick]
	chartCandleSubscriptions []*subscription[ChartCandle]
	tradeSubscriptions       []*subscription[TradeUpdate]
	priceTimestampsByEpic    map[string]struct{ fetchTime, priceTime time.Time }
	subscriptionReqChan      chan interface{}
}

type subscription[T MarketTick | ChartTick | ChartCandle | TradeUpdate] struct {
	channel         chan T
	subscriptionId  string
	requestID       string
	items           []string
	scale           string // Used for ChartCandle subscriptions
	lastStateByItem map[string]T
	errChan         chan error
}

type unsubscription[T MarketTick | ChartTick | ChartCandle | TradeUpdate] struct {
	channel <-chan T
	errChan chan error
}

type MarketTick struct {
	Epic          string
	RecvTime      time.Time
	RawUpdate     string      // The raw payload from the lightstreamer API
	Time          time.Time   `lightstreamer:"UPDATE_TIME,timeFromHHMMSS"`
	MarketDelay   bool        `lightstreamer:"MARKET_DELAY,boolFromInt"`
	Bid           float64     `lightstreamer:"BID"`
	Ask           float64     `lightstreamer:"OFFER"`
	High          float64     `lightstreamer:"HIGH"`
	Low           float64     `lightstreamer:"LOW"`
	MidOpen       float64     `lightstreamer:"MID_OPEN"`
	Change        float64     `lightstreamer:"CHANGE"`
	ChangePercent float64     `lightstreamer:"CHANGE_PCT"`
	MarketState   MarketState `lightstreamer:"MARKET_STATE,marketStateFromString"`
}

type ChartTick struct {
	Epic                     string
	RecvTime                 time.Time
	RawUpdate                string    // The raw payload from the lightstreamer API
	UpdateTime               time.Time `lightstreamer:"UTM,timeFromEpochMilliseconds"`
	Bid                      float64   `lightstreamer:"BID"`
	Ask                      float64   `lightstreamer:"OFR"`
	LastTradedPrice          float64   `lightstreamer:"LTP"`
	LastTradedVolume         float64   `lightstreamer:"LTV"`
	IncrementalTradingVolume float64   `lightstreamer:"TTV"`
	MidOpen                  float64   `lightstreamer:"DAY_OPEN_MID"`
	High                     float64   `lightstreamer:"DAY_HIGH"`
	Low                      float64   `lightstreamer:"DAY_LOW"`
	Change                   float64   `lightstreamer:"DAY_NET_CHG_MID"`
	ChangePercent            float64   `lightstreamer:"DAY_PERC_CHG_MID"`
}

type ChartCandle struct {
	Epic                     string
	RecvTime                 time.Time
	Scale                    string
	RawUpdate                string    // The raw payload from the lightstreamer API
	UpdateTime               time.Time `lightstreamer:"UTM,timeFromEpochMilliseconds"`
	LastTradedVolume         float64   `lightstreamer:"LTV"`
	IncrementalTradingVolume float64   `lightstreamer:"TTV"`
	DayOpenMid               float64   `lightstreamer:"DAY_OPEN_MID"`
	DayChangeMid             float64   `lightstreamer:"DAY_NET_CHG_MID"`
	DayChangePercentMid      float64   `lightstreamer:"DAY_PERC_CHG_MID"`
	DayHighMid               float64   `lightstreamer:"DAY_HIGH"`
	DayLowMid                float64   `lightstreamer:"DAY_LOW"`
	OfferOpen                float64   `lightstreamer:"OFR_OPEN"`
	OfferHigh                float64   `lightstreamer:"OFR_HIGH"`
	OfferLow                 float64   `lightstreamer:"OFR_LOW"`
	OfferClose               float64   `lightstreamer:"OFR_CLOSE"`
	BidOpen                  float64   `lightstreamer:"BID_OPEN"`
	BidHigh                  float64   `lightstreamer:"BID_HIGH"`
	BidLow                   float64   `lightstreamer:"BID_LOW"`
	BidClose                 float64   `lightstreamer:"BID_CLOSE"`
	LastTradedPriceOpen      float64   `lightstreamer:"LTP_OPEN"`
	LastTradedPriceHigh      float64   `lightstreamer:"LTP_HIGH"`
	LastTradedPriceLow       float64   `lightstreamer:"LTP_LOW"`
	LastTradedPriceClose     float64   `lightstreamer:"LTP_CLOSE"`
	Consolidated             bool      `lightstreamer:"CONS_END,boolFromInt"`
	ConsolidatedTickCount    int       `lightstreamer:"CONS_TICK_COUNT,int"`
}

type TradeUpdate struct {
	AccountID          string
	RecvTime           time.Time
	RawUpdate          string               // The raw payload from the lightstreamer API
	Confirms           *TradeUpdateConfirms `lightstreamer:"CONFIRMS,tradeUpdateConfirms"`
	OpenPositionUpdate *TradeUpdateOPU      `lightstreamer:"OPU,tradeUpdateOPU"`
	WorkingOrderUpdate *TradeUpdateWOU      `lightstreamer:"WOU,tradeUpdateWOU"`
}

type DealingWindow struct {
	Size   float64 `json:"size"`
	Expiry int64   `json:"expiry"`
}

type RepeatDealingWindow struct {
	Entries []DealingWindow `json:"entries"`
}

type TradeUpdateConfirms struct {
	Date                string              `json:"date"`
	LimitDistance       *float64            `json:"limitDistance"`
	Reason              string              `json:"reason"`
	LimitLevel          *float64            `json:"limitLevel"`
	Level               *float64            `json:"level"`
	DealID              string              `json:"dealId"`
	Channel             string              `json:"channel"`
	Epic                string              `json:"epic"`
	DealReference       string              `json:"dealReference"`
	DealStatus          string              `json:"dealStatus"`
	TrailingStop        bool                `json:"trailingStop"`
	Size                *float64            `json:"size"`
	StopLevel           *float64            `json:"stopLevel"`
	StopDistance        *float64            `json:"stopDistance"`
	ProfitCurrency      *string             `json:"profitCurrency"`
	Expiry              *string             `json:"expiry"`
	Profit              *float64            `json:"profit"`
	AffectedDeals       []AffectedDeal      `json:"affectedDeals"`
	RepeatDealingWindow RepeatDealingWindow `json:"repeatDealingWindow"`
	GuaranteedStop      bool                `json:"guaranteedStop"`
	Direction           string              `json:"direction"`
	Status              *string             `json:"status"`
}

type TradeUpdateOPU struct {
	DealReference  string  `json:"dealReference"`
	DealID         string  `json:"dealId"`
	Direction      string  `json:"direction"`
	Epic           string  `json:"epic"`
	Status         string  `json:"status"`
	DealStatus     string  `json:"dealStatus"`
	Level          float64 `json:"level"`
	Size           float64 `json:"size"`
	Timestamp      string  `json:"timestamp"`
	Channel        string  `json:"channel"`
	DealIDOrigin   string  `json:"dealIdOrigin"`
	Expiry         string  `json:"expiry"`
	StopLevel      float64 `json:"stopLevel"`
	LimitLevel     float64 `json:"limitLevel"`
	GuaranteedStop bool    `json:"guaranteedStop"`
}

type TradeUpdateWOU struct {
	// TODO - populate this
}

const LightstreamerProtocolVersion = "TLCP-2.4.0.lightstreamer.com"

const chartTickFields = "BID OFR DAY_NET_CHG_MID DAY_PERC_CHG_MID LTP LTV TTV UTM DAY_OPEN_MID DAY_HIGH DAY_LOW"

const marketTickFields = "BID OFFER CHANGE CHANGE_PCT UPDATE_TIME MARKET_DELAY MID_OPEN HIGH LOW MARKET_STATE"

const tradeFields = "CONFIRMS OPU WOU"

const chartCandleFields = "UTM LTV TTV DAY_OPEN_MID DAY_NET_CHG_MID DAY_PERC_CHG_MID DAY_HIGH DAY_LOW OFR_OPEN OFR_HIGH OFR_LOW OFR_CLOSE BID_OPEN BID_HIGH BID_LOW BID_CLOSE LTP_OPEN LTP_HIGH LTP_LOW LTP_CLOSE CONS_END CONS_TICK_COUNT"

type typeMap struct {
	typ             reflect.Type
	fields          string
	structFields    []int
	conversionFuncs []func(ls *LightStreamerConnection, epic string, field string) (reflect.Value, error)
}

var typeMaps = [...]typeMap{
	{
		typ:    reflect.TypeOf(ChartTick{}),
		fields: chartTickFields,
	},
	{
		typ:    reflect.TypeOf(MarketTick{}),
		fields: marketTickFields,
	},
	{
		typ:    reflect.TypeOf(TradeUpdate{}),
		fields: tradeFields,
	},
	{
		typ:    reflect.TypeOf(ChartCandle{}),
		fields: chartCandleFields,
	},
}

func init() {
	timeZone, err := time.LoadLocation("Europe/London")
	if err != nil {
		panic(err)
	}

	convFuncs := map[string]func(*LightStreamerConnection, string, string) (reflect.Value, error){
		"timeFromHHMMSS": func(ls *LightStreamerConnection, epic, s string) (reflect.Value, error) {
			now := time.Now().In(timeZone)
			parsedTime, err := time.ParseInLocation("2006-1-2 15:04:05", fmt.Sprintf("%d-%d-%d %s",
				now.Year(), now.Month(), now.Day(), s), timeZone)
			if err != nil {
				return reflect.Value{}, err
			}
			if parsedTime.Sub(now).Seconds() > 10 || parsedTime.Sub(now).Seconds() < -3600 {
				// Probably a time from when the market closed, check the date of the last reported price
				// TODO - check whether we fetched the last price recently, and reuse it, otherwise fetch it again
				times, ok := ls.priceTimestampsByEpic[epic]
				if !ok || now.Sub(times.fetchTime).Seconds() > 3600 {
					priceResp, err := ls.ig.GetPriceHistory(ls.ctx, epic, "SECOND", 1, time.Time{}, time.Time{})
					if err != nil || len(priceResp.Prices) != 1 {
						log.Printf("failed to fetch price history for epic '%s' - received ambiguous time %s but could not determine date: %v\n", epic, s, err)
						times.priceTime = time.Time{}
					} else {
						times.priceTime = priceResp.Prices[0].SnapshotTimeUTCParsed
						times.fetchTime = now
						ls.priceTimestampsByEpic[epic] = times
					}
				}
				parsedTime, err = time.ParseInLocation("2006-1-2 15:04:05", fmt.Sprintf("%04d-%d-%d %s",
					times.priceTime.Year(), times.priceTime.Month(), times.priceTime.Day(), s), timeZone)
				if err != nil {
					return reflect.Value{}, err
				}
			}

			return reflect.ValueOf(parsedTime), nil
		},
		"timeFromEpochMilliseconds": func(ls *LightStreamerConnection, epic, s string) (reflect.Value, error) {
			milli, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(time.UnixMilli(milli)), nil
		},
		"marketStateFromString": func(ls *LightStreamerConnection, epic, s string) (reflect.Value, error) {
			return reflect.ValueOf(MarketState(0).Parse(s)), nil
		},
		"tradeUpdateConfirms": func(ls *LightStreamerConnection, accountID, s string) (reflect.Value, error) {
			confirms := &TradeUpdateConfirms{}
			err := json.Unmarshal([]byte(s), confirms)
			return reflect.ValueOf(confirms), err
		},
		"tradeUpdateOPU": func(ls *LightStreamerConnection, accountID, s string) (reflect.Value, error) {
			opu := &TradeUpdateOPU{}
			err := json.Unmarshal([]byte(s), opu)
			return reflect.ValueOf(opu), err
		},
		"tradeUpdateWOU": func(ls *LightStreamerConnection, accountID, s string) (reflect.Value, error) {
			wou := &TradeUpdateWOU{}
			err := json.Unmarshal([]byte(s), wou)
			return reflect.ValueOf(wou), err
		},
		"boolFromInt": func(ls *LightStreamerConnection, epic, s string) (reflect.Value, error) {
			val, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			if val == 0 {
				return reflect.ValueOf(false), nil
			} else {
				return reflect.ValueOf(true), nil
			}
		},
		"float64": func(ls *LightStreamerConnection, epic, s string) (reflect.Value, error) {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(f), nil
		},
		"int": func(ls *LightStreamerConnection, epic, s string) (reflect.Value, error) {
			i, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(int(i)), nil
		},
	}

	for x := range typeMaps {
		t := typeMaps[x].typ
		fieldsList := strings.Split(typeMaps[x].fields, " ")
		structFields := make([]int, len(fieldsList))
		conversionFuncs := make([]func(*LightStreamerConnection, string, string) (reflect.Value, error), len(fieldsList))
		for j, field := range fieldsList {
			for i := 0; i < t.NumField(); i++ {
				tag := t.Field(i).Tag.Get("lightstreamer")
				if tag == "" {
					continue
				}
				tagFrags := strings.Split(tag, ",")
				if tagFrags[0] == field {
					structFields[j] = i
					var ok bool
					if len(tagFrags) > 1 {
						conversionFuncs[j], ok = convFuncs[tagFrags[1]]
						if !ok {
							panic("missing conversion function " + tagFrags[1])
						}
					} else if t.Field(i).Type == reflect.TypeOf(float64(0)) {
						conversionFuncs[j] = convFuncs["float64"]
					} else if t.Field(i).Type == reflect.TypeOf(int(0)) {
						conversionFuncs[j] = convFuncs["int"]
					} else {
						panic(fmt.Sprintf("unhandled field type %s for field %s", t.Field(i).Type, field))
					}
				}
			}
		}
		typeMaps[x].structFields = structFields
		typeMaps[x].conversionFuncs = conversionFuncs
	}
}

func (ig *IGMarkets) NewLightStreamerConnection(ctx context.Context) (*LightStreamerConnection, error) {
	lsConn := &LightStreamerConnection{
		ig:                    ig,
		nextSubscriptionId:    1,
		subscriptionReqChan:   make(chan interface{}),
		priceTimestampsByEpic: make(map[string]struct{ fetchTime, priceTime time.Time }),
	}

	ctx, lsConn.cancelFunc = context.WithCancel(ctx)
	lsConn.ctx = ctx

	sessionVersion2, err := ig.LoginVersion2(ctx)
	if err != nil {
		return nil, fmt.Errorf("error during login version 2: %w", err)
	}

	dialer := &websocket.Dialer{
		Subprotocols: []string{LightstreamerProtocolVersion},
	}

	endpoint := strings.Replace(strings.Replace(sessionVersion2.LightstreamerEndpoint, "https://", "wss://", 1), "http://", "ws://", 1)
	ws, resp, err := dialer.Dial(fmt.Sprintf("%s/lightstreamer", endpoint), http.Header{})
	if err != nil {
		if resp != nil {
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error when reading websockets response body: %w", err)
			}
			if len(respBody) > 0 {
				return nil, fmt.Errorf("error during websockets dial (response %d : %s): %w", resp.StatusCode, respBody, err)
			}
			return nil, fmt.Errorf("error during websockets dial (response %d): %w", resp.StatusCode, err)
		} else {
			return nil, fmt.Errorf("error during websockets dial: %w", err)
		}
	}
	lsConn.wsConn = ws

	// to ensure creation fails if server is unresponsive for too long
	err = ws.SetReadDeadline(time.Now().Add(time.Second * 60))
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Now().Add(time.Second*5))
		return nil, fmt.Errorf("error setting read deadline: %w", err)
	}
	err = ws.SetWriteDeadline(time.Now().Add(time.Second * 60))
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Now().Add(time.Second*5))
		return nil, fmt.Errorf("error setting write deadline: %w", err)
	}

	err = lsConn.validateWebsocketConnection()
	if err != nil {
		return nil, err
	}

	err = lsConn.createSession(sessionVersion2)
	if err != nil {
		return nil, err
	}

	err = lsConn.bindSession()
	if err != nil {
		return nil, err
	}

	err = ws.SetReadDeadline(time.Time{})
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Now().Add(time.Second*5))
		return nil, fmt.Errorf("error setting empty read deadline: %w", err)
	}
	err = ws.SetWriteDeadline(time.Time{})
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Now().Add(time.Second*5))
		return nil, fmt.Errorf("error setting empty write deadline: %w", err)
	}

	lsConn.heartbeatTicker = time.NewTicker(5 * time.Second)
	go lsConn.readLoop()

	go lsConn.writeLoop(ctx)

	return lsConn, nil
}

func (ls *LightStreamerConnection) validateWebsocketConnection() error {
	err := ls.wsConn.WriteMessage(websocket.TextMessage, []byte("wsok\r\n"))
	if err != nil {
		_ = ls.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Time{})
		return fmt.Errorf("error writing 'wsok': %w", err)
	}

	_, wsOkResponse, err := ls.wsConn.ReadMessage()
	if err != nil {
		_ = ls.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Time{})
		return fmt.Errorf("error reading 'wsok' response: %w", err)
	}

	if strings.TrimSpace(string(wsOkResponse)) != "WSOK" {
		_ = ls.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Time{})
		return fmt.Errorf("unexpected response from 'wsok': %s", wsOkResponse)
	}

	return nil
}

func (ls *LightStreamerConnection) createSession(sessionVersion2 *SessionVersion2) error {
	ws := ls.wsConn
	createVals := url.Values{}
	createVals.Set("LS_reduce_head", "true")
	createVals.Set("LS_cid", "mgQkwtwdysogQz2BJ4Ji kOj2Bg") // Lightstreamer Client ID https://sdk.lightstreamer.com/ls-generic-client/2.4.0/TLCP%20Specifications.pdf
	createVals.Set("LS_user", sessionVersion2.CurrentAccountId)
	createVals.Set("LS_password", "CST-"+sessionVersion2.CSTToken+"|"+"XST-"+sessionVersion2.XSTToken)

	createMsg := []byte("create_session\r\n" + createVals.Encode())

	err := ws.WriteMessage(websocket.TextMessage, createMsg)
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Time{})
		return fmt.Errorf("error writing 'create_session' : %w", err)
	}

	err = ws.SetReadDeadline(time.Now().Add(time.Second * 30))
	_, createResponse, err := ws.ReadMessage()
	if err != nil {
		_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Time{})
		return fmt.Errorf("error reading 'create_session' response: %w", err)
	}

	// createResponse looks something like:
	// CONOK,S98d4635f6979fe17Ma7eT3149399,50000,5000,apd245f.marketdatasystems.com
	// SERVNAME,Lightstreamer HTTPS NSE Listener
	// CLIENTIP,172.28.3.205

	createResponseLines := strings.Split(string(createResponse), "\r\n")
	for _, line := range createResponseLines {
		csvResp := strings.Split(line, ",")
		switch csvResp[0] {
		case "SERVNAME", "CLIENTIP":
			// Nothing to do
		case "CONOK":
			if len(csvResp) != 5 {
				_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Time{})
				return fmt.Errorf("expected CONOK reponse to have 5 fields, got: %s", createResponse)
			}
			sessionID := csvResp[1]
			serverHostname := csvResp[4]
			ls.sessionID = sessionID
			ls.serverHostname = serverHostname
		case "":
			// nothing to do
		default:
			_ = ws.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Time{})
			return fmt.Errorf("unexpected response from create_session: %s", createResponse)
		}
	}

	return nil
}

func (ls *LightStreamerConnection) bindSession() error {
	bindVals := url.Values{}
	bindVals.Set("LS_session", ls.sessionID)
	bindMsg := []byte("bind_session\r\n" + bindVals.Encode())

	err := ls.wsConn.WriteMessage(websocket.TextMessage, bindMsg)
	if err != nil {
		_ = ls.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""), time.Time{})
		return fmt.Errorf("error writing 'bind_session': %w", err)
	}
	ls.sessionBindTime = time.Now() // Used to keep track of processing latency via SYNC messages which report time in seconds since session bind
	return nil
}

func (ls *LightStreamerConnection) writeLoop(ctx context.Context) {
	pprof.SetGoroutineLabels(pprof.WithLabels(context.Background(), pprof.Labels("func", "LightStreamerConnection.writeLoop")))
	for {
		select {
		case <-ls.heartbeatTicker.C:
			hbVals := url.Values{}
			hbVals.Set("LS_session", ls.sessionID)
			hbMsg := []byte("heartbeat\r\n" + hbVals.Encode())
			err := ls.wsConn.WriteMessage(websocket.TextMessage, hbMsg)
			if err != nil {
				if !ls.closeRequested.Load() {
					ls.fatalError(err)
				}
			}
		case <-ctx.Done():
			return
		case subReqI := <-ls.subscriptionReqChan:
			switch subReq := subReqI.(type) {
			case subscription[ChartTick]:
				ls.nextSubscriptionId++
				requestID := rand.Int63()
				var items = make([]string, len(subReq.items))
				for i, epic := range subReq.items {
					items[i] = fmt.Sprintf("CHART:%s:TICK", epic)
				}
				subReq.requestID = fmt.Sprintf("%d", requestID)
				subReq.subscriptionId = fmt.Sprintf("%d", ls.nextSubscriptionId)
				subReq.lastStateByItem = make(map[string]ChartTick)
				ctrlVals := url.Values{}
				ctrlVals.Set("LS_op", "add")
				ctrlVals.Set("LS_mode", "DISTINCT")
				ctrlVals.Set("LS_snapshot", "true")
				ctrlVals.Set("LS_subId", fmt.Sprintf("%s", subReq.subscriptionId))
				ctrlVals.Set("LS_group", strings.Join(items, " "))
				ctrlVals.Set("LS_schema", chartTickFields)
				ctrlVals.Set("LS_reqId", fmt.Sprintf("%s", subReq.requestID))
				msg := []byte("control\r\n" + ctrlVals.Encode())
				err := ls.wsConn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					subReq.errChan <- err
					if !ls.closeRequested.Load() {
						ls.fatalError(err)
					}
					return
				}

				ls.chartTickSubscriptions = append(ls.chartTickSubscriptions, &subReq)
				// Read loop should handle REQOK/SUBOK/ERROR
			case subscription[MarketTick]:
				ls.nextSubscriptionId++
				requestID := rand.Int63()
				var items = make([]string, len(subReq.items))
				for i, epic := range subReq.items {
					items[i] = fmt.Sprintf("MARKET:%s", epic)
				}
				subReq.requestID = fmt.Sprintf("%d", requestID)
				subReq.subscriptionId = fmt.Sprintf("%d", ls.nextSubscriptionId)
				subReq.lastStateByItem = make(map[string]MarketTick)
				ctrlVals := url.Values{}
				ctrlVals.Set("LS_op", "add")
				ctrlVals.Set("LS_mode", "MERGE")
				ctrlVals.Set("LS_snapshot", "true")
				ctrlVals.Set("LS_subId", fmt.Sprintf("%s", subReq.subscriptionId))
				ctrlVals.Set("LS_group", strings.Join(items, " "))
				ctrlVals.Set("LS_schema", marketTickFields)
				ctrlVals.Set("LS_reqId", fmt.Sprintf("%s", subReq.requestID))
				msg := []byte("control\r\n" + ctrlVals.Encode())

				err := ls.wsConn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					subReq.errChan <- err
					ls.lastError = err
					if !ls.closeRequested.Load() {
						ls.fatalError(err)
					}
					return
				}
				ls.marketSubscriptions = append(ls.marketSubscriptions, &subReq)
				// Read loop should handle REQOK/SUBOK/ERROR
			case subscription[TradeUpdate]:
				ls.nextSubscriptionId++
				requestID := rand.Int63()
				var items = make([]string, len(subReq.items))
				for i, accountID := range subReq.items {
					items[i] = fmt.Sprintf("TRADE:%s", accountID)
				}
				subReq.requestID = fmt.Sprintf("%d", requestID)
				subReq.subscriptionId = fmt.Sprintf("%d", ls.nextSubscriptionId)
				subReq.lastStateByItem = make(map[string]TradeUpdate)
				ctrlVals := url.Values{}
				ctrlVals.Set("LS_op", "add")
				ctrlVals.Set("LS_mode", "DISTINCT")
				ctrlVals.Set("LS_snapshot", "true")
				ctrlVals.Set("LS_subId", fmt.Sprintf("%s", subReq.subscriptionId))
				ctrlVals.Set("LS_group", strings.Join(items, " "))
				ctrlVals.Set("LS_schema", tradeFields)
				ctrlVals.Set("LS_reqId", fmt.Sprintf("%s", subReq.requestID))
				msg := []byte("control\r\n" + ctrlVals.Encode())
				err := ls.wsConn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					subReq.errChan <- err
					if !ls.closeRequested.Load() {
						ls.fatalError(err)
					}
					return
				}

				ls.tradeSubscriptions = append(ls.tradeSubscriptions, &subReq)
				// Read loop should handle REQOK/SUBOK/ERROR
			case subscription[ChartCandle]:
				ls.nextSubscriptionId++
				requestID := rand.Int63()
				var items = make([]string, len(subReq.items))
				for i, epic := range subReq.items {
					items[i] = fmt.Sprintf("CHART:%s:%s", epic, subReq.scale)
				}
				subReq.requestID = fmt.Sprintf("%d", requestID)
				subReq.subscriptionId = fmt.Sprintf("%d", ls.nextSubscriptionId)
				subReq.lastStateByItem = make(map[string]ChartCandle)
				ctrlVals := url.Values{}
				ctrlVals.Set("LS_op", "add")
				ctrlVals.Set("LS_mode", "MERGE")
				ctrlVals.Set("LS_snapshot", "true")
				ctrlVals.Set("LS_subId", fmt.Sprintf("%s", subReq.subscriptionId))
				ctrlVals.Set("LS_group", strings.Join(items, " "))
				ctrlVals.Set("LS_schema", chartCandleFields)
				ctrlVals.Set("LS_reqId", fmt.Sprintf("%s", subReq.requestID))
				msg := []byte("control\r\n" + ctrlVals.Encode())
				err := ls.wsConn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					subReq.errChan <- err
					if !ls.closeRequested.Load() {
						ls.fatalError(err)
					}
					return
				}

				ls.chartCandleSubscriptions = append(ls.chartCandleSubscriptions, &subReq)
				// Read loop should handle REQOK/SUBOK/ERROR
			case unsubscription[MarketTick]:
				sendUnsubscribeRequest(ls, subReq, ls.marketSubscriptions)
				// reader will receive UNSUB
			case unsubscription[ChartTick]:
				sendUnsubscribeRequest(ls, subReq, ls.chartTickSubscriptions)
				// reader will receive UNSUB
			case unsubscription[TradeUpdate]:
				sendUnsubscribeRequest(ls, subReq, ls.tradeSubscriptions)
				// reader will receive UNSUB
			case unsubscription[ChartCandle]:
				sendUnsubscribeRequest(ls, subReq, ls.chartCandleSubscriptions)
				// reader will receive UNSUB
			default:
				panic(fmt.Sprintf("unknown subscription request type %T", subReqI))
			}
		}
	}
}

func (ls *LightStreamerConnection) readLoop() {
	pprof.SetGoroutineLabels(pprof.WithLabels(context.Background(), pprof.Labels("func", "LightStreamerConnection.readLoop")))
	for !ls.closeRequested.Load() {
		_, msg, err := ls.wsConn.ReadMessage()
		if err != nil {
			if !ls.closeRequested.Load() {
				ls.fatalError(err)
			}
			return
		}

		lines := strings.Split(string(msg), "\r\n")
		for _, line := range lines {
			csvResp := strings.SplitN(line, ",", 4)
			switch csvResp[0] {
			case "CONOK", "SERVNAME", "CLIENTIP", "PROBE", "CONF", "CONS", "EOS", "REQOK", "":
				// Nothing to do here
			case "REQERR":
				ls.handleReqErr(csvResp)
			case "SUBOK":
				ls.handleSubOk(csvResp)
			case "UNSUB":
				ls.handleUnsubscribe(csvResp)
			case "END":
				ls.handleEnd(csvResp)
			case "SYNC":
				ls.handleSync(csvResp)
			case "LOOP":
				if !ls.closeRequested.Load() {
					ls.fatalError(fmt.Errorf("server sent LOOP message"))
				}
			case "U":
				ls.handleUpdate(csvResp)
			default:
				log.Printf("WARNING - unhandled lightstreamer message: %s\n", line)
			}
		}
	}
}

func (ls *LightStreamerConnection) handleUpdate(args []string) {
	if len(args) != 4 {
		log.Printf("expected 4 fields in update message, got %d: %s\n", len(args), args)
		return
	}
	subID := args[1]

	itemIdIndex, err := strconv.Atoi(args[2])
	if err != nil {
		log.Printf("error parsing item index: %v\n", err)
		return
	}
	itemIdIndex--
	for _, sub := range ls.marketSubscriptions {
		if sub.subscriptionId == subID {
			if itemIdIndex < 0 || itemIdIndex >= len(sub.items) {
				log.Printf("epic index %d out of range\n", itemIdIndex)
				return
			}
			epic := sub.items[itemIdIndex]
			lastState := sub.lastStateByItem[epic]
			marketTick := lastState
			marketTick.Epic = epic
			marketTick.RecvTime = time.Now()
			marketTick.RawUpdate = args[3]
			err = parseUpdate(ls, args[3], "Epic", reflect.ValueOf(&marketTick))
			if err != nil {
				log.Printf("failed to parse update %s: %s\n", args[3], err)
				return
			}
			sub.channel <- marketTick
			sub.lastStateByItem[epic] = marketTick
			return
		}
	}

	for _, sub := range ls.chartTickSubscriptions {
		if sub.subscriptionId == subID {
			if itemIdIndex < 0 || itemIdIndex >= len(sub.items) {
				log.Printf("epic index %d out of range\n", itemIdIndex)
				return
			}
			epic := sub.items[itemIdIndex]
			lastState := sub.lastStateByItem[epic]
			chartTick := lastState
			chartTick.Epic = epic
			chartTick.RecvTime = time.Now()
			chartTick.RawUpdate = args[3]
			err = parseUpdate(ls, args[3], "Epic", reflect.ValueOf(&chartTick))
			if err != nil {
				log.Printf("failed to parse update %s: %s\n", args[3], err)
				return
			}
			sub.channel <- chartTick
			sub.lastStateByItem[epic] = chartTick
			return
		}
	}

	for _, sub := range ls.chartCandleSubscriptions {
		if sub.subscriptionId == subID {
			if itemIdIndex < 0 || itemIdIndex >= len(sub.items) {
				log.Printf("epic index %d out of range\n", itemIdIndex)
				return
			}
			epic := sub.items[itemIdIndex]
			lastState := sub.lastStateByItem[epic]
			chartCandle := lastState
			chartCandle.Epic = epic
			chartCandle.RecvTime = time.Now()
			chartCandle.RawUpdate = args[3]
			chartCandle.Scale = sub.scale
			err = parseUpdate(ls, args[3], "Epic", reflect.ValueOf(&chartCandle))
			if err != nil {
				log.Printf("failed to parse update %s: %s\n", args[3], err)
				return
			}
			sub.channel <- chartCandle
			sub.lastStateByItem[epic] = chartCandle
			return
		}
	}

	for _, sub := range ls.tradeSubscriptions {
		if sub.subscriptionId == subID {
			if itemIdIndex < 0 || itemIdIndex >= len(sub.items) {
				log.Printf("account ID index %d out of range\n", itemIdIndex)
				return
			}
			accountID := sub.items[itemIdIndex]
			lastState := sub.lastStateByItem[accountID]
			tradeUpdate := lastState
			tradeUpdate.AccountID = accountID
			tradeUpdate.RecvTime = time.Now()
			tradeUpdate.RawUpdate = args[3]
			err = parseUpdate(ls, args[3], "AccountID", reflect.ValueOf(&tradeUpdate))
			if err != nil {
				log.Printf("failed to parse update %s: %s\n", args[3], err)
				return
			}
			sub.channel <- tradeUpdate
			sub.lastStateByItem[accountID] = tradeUpdate
			return
		}
	}
}

func parseUpdate(ls *LightStreamerConnection, update string, itemIDField string, output reflect.Value) error {
	output = output.Elem()
	epic := output.FieldByName(itemIDField).String()
	var convFuncs []func(*LightStreamerConnection, string, string) (reflect.Value, error)
	var structFields []int
	for _, typemap := range typeMaps {
		if typemap.typ == output.Type() {
			convFuncs = typemap.conversionFuncs
			structFields = typemap.structFields
			break
		}
	}
	if structFields == nil {
		panic(fmt.Sprintf("unexpected type %s", output.Type()))
	}

	fields := strings.Split(update, "|")
	var offset int
	for i, field := range fields {
		if i+offset >= len(structFields) {
			return fmt.Errorf("unexpected extra field '%s' in update '%s'", field, update)
		}

		switch field {
		case "#", "$", "":
			// null, empty, empty respectively
		default:
			if strings.HasPrefix(field, "^") {
				fieldCount, err := strconv.ParseInt(field[1:], 10, 64)
				if err != nil {
					return fmt.Errorf("failed to parse field count %s in update '%s': %w", field, update, err)
				}
				offset += int(fieldCount - 1)
			} else {
				val, err := convFuncs[i+offset](ls, epic, field)
				if err != nil {
					return err
				}
				output.Field(structFields[i+offset]).Set(val)
			}
		}
	}
	return nil
}

func (ls *LightStreamerConnection) handleReqErr(args []string) {
	if len(args) != 4 {
		log.Printf("expected 4 fields in REQERR message, got: %s\n", args)
		return
	}
	handleReqErr(args, ls.marketSubscriptions)
	handleReqErr(args, ls.chartTickSubscriptions)
	handleReqErr(args, ls.tradeSubscriptions)
	handleReqErr(args, ls.chartCandleSubscriptions)
}

func handleReqErr[T MarketTick | ChartTick | ChartCandle | TradeUpdate](args []string, subs []*subscription[T]) {
	reqID := args[1]
	for _, sub := range subs {
		if sub.requestID == reqID {
			select {
			case sub.errChan <- fmt.Errorf("request error: code: %s, message: %s", args[2], args[3]):
			}
			break
		}
	}
}

func sendUnsubscribeRequest[T MarketTick | ChartTick | ChartCandle | TradeUpdate](ls *LightStreamerConnection, unSubReq unsubscription[T], subs []*subscription[T]) {
	found := false
	for _, sub := range subs {
		if sub.channel == unSubReq.channel {
			found = true
			ctrlVals := url.Values{}
			sub.errChan = unSubReq.errChan
			ctrlVals.Set("LS_op", "delete")
			ctrlVals.Set("LS_subId", fmt.Sprintf("%s", sub.subscriptionId))
			ctrlVals.Set("LS_reqId", fmt.Sprintf("%d", rand.Int63()))
			msg := []byte("control\r\n" + ctrlVals.Encode())

			err := ls.wsConn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				unSubReq.errChan <- err
				ls.lastError = err
				if !ls.closeRequested.Load() {
					ls.fatalError(err)
				}
			}
			break
		}
	}
	if !found {
		unSubReq.errChan <- fmt.Errorf("market subscription not found")
	}
}

func (ls *LightStreamerConnection) handleSubOk(args []string) {
	if len(args) != 4 {
		log.Printf("expected 4 fields in SUBOK message, got: %s\n", args)
		return
	}

	subID := args[1]
	for _, sub := range ls.marketSubscriptions {
		if sub.subscriptionId == subID {
			select {
			case sub.errChan <- nil:
			}
			break
		}
	}
	for _, sub := range ls.chartTickSubscriptions {
		if sub.subscriptionId == subID {
			select {
			case sub.errChan <- nil:
			}
			break
		}
	}
	for _, sub := range ls.tradeSubscriptions {
		if sub.subscriptionId == subID {
			select {
			case sub.errChan <- nil:
			}
			break
		}
	}
	for _, sub := range ls.chartCandleSubscriptions {
		if sub.subscriptionId == subID {
			select {
			case sub.errChan <- nil:
			}
			break
		}
	}
}

func (ls *LightStreamerConnection) handleUnsubscribe(args []string) {
	if len(args) != 2 {
		log.Printf("expected 2 fields in UNSUB message, got: %s\n", args)
		return
	}

	subID := args[1]
	handleUnsubscribe(subID, &ls.marketSubscriptions)
	handleUnsubscribe(subID, &ls.chartTickSubscriptions)
	handleUnsubscribe(subID, &ls.tradeSubscriptions)
	handleUnsubscribe(subID, &ls.chartCandleSubscriptions)
}

func handleUnsubscribe[T MarketTick | ChartTick | ChartCandle | TradeUpdate](subID string, subs *[]*subscription[T]) {
	var foundIndex int
	var found bool
	for i, sub := range *subs {
		if sub.subscriptionId == subID {
			found = true
			foundIndex = i
			select {
			case sub.errChan <- nil:
				close(sub.channel)
			}
			break
		}
	}
	if found {
		*subs = append((*subs)[:foundIndex], (*subs)[foundIndex+1:]...)
	}
}

func (ls *LightStreamerConnection) handleSync(args []string) {
	const maxClientLagSeconds = 2.0
	if len(args) != 2 {
		log.Printf("expected 2 fields in SYNC message, got: %s\n", args)
		return
	}
	seconds, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		log.Printf("failed to parse seconds from SYNC message (%s): %s\n", args, err)
		return
	}
	clientDuration := time.Now().Sub(ls.sessionBindTime)

	if clientDuration.Seconds()-seconds > maxClientLagSeconds {
		log.Printf("WARNING - lightstreamer client detected drift of %f seconds\n", clientDuration.Seconds()-seconds)
	}
}

func (ls *LightStreamerConnection) handleEnd(args []string) {
	if len(args) != 3 {
		log.Printf("expected 2 fields in END message, got: %s\n", args)
		return
	}
	if args[1] != "40" {
		log.Printf("unexpected cause code %s in END message: %s\n", args[1], args[2])
	}
}

func (ls *LightStreamerConnection) fatalError(err error) {
	ls.lastError = err

	log.Printf("lightstreamer encountered connection error - recreating subscriptions: %s\n", err)
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = 0
	expBackoff.MaxInterval = 10 * time.Second
	backOff := backoff.WithContext(expBackoff, ls.ctx)
	err = backoff.RetryNotify(func() error {
		lsNew, err := ls.ig.NewLightStreamerConnection(ls.ctx)
		if err != nil {
			return err
		}
		// Reuse the old subscription requests so the channels are re-used
		for _, ms := range ls.marketSubscriptions {
			_, err = subscribe(lsNew, lsNew.ctx, *ms)
			if err != nil {
				_ = lsNew.Close()
				return err
			}
		}

		for _, cs := range ls.chartTickSubscriptions {
			_, err = subscribe(lsNew, lsNew.ctx, *cs)
			if err != nil {
				_ = lsNew.Close()
				return err
			}
		}
		for _, ts := range ls.tradeSubscriptions {
			_, err = subscribe(lsNew, lsNew.ctx, *ts)
			if err != nil {
				_ = lsNew.Close()
				return err
			}
		}
		for _, cs := range ls.chartCandleSubscriptions {
			_, err = subscribe(lsNew, lsNew.ctx, *cs)
			if err != nil {
				_ = lsNew.Close()
				return err
			}
		}
		ls.closeRequested.Store(true)
		ls.cancelFunc()
		_ = ls.wsConn.Close()

		*ls = *lsNew
		return nil
	}, backOff, func(err error, duration time.Duration) {
		log.Printf("lightstreamer encountered connection error (retrying in %s): %s\n", duration, err)
	})
	if err != nil {
		log.Printf("lightstreamer encountered fatal error")
	}
}

func (ls *LightStreamerConnection) SubscribeTradeUpdates(ctx context.Context, bufferSize int, accounts ...string) (<-chan TradeUpdate, error) {
	if ls.closeRequested.Load() {
		return nil, fmt.Errorf("cannot subscribe using a closed lightstreamer connection")
	}

	return subscribe(ls, ctx, makeNewSubscription[TradeUpdate](accounts, bufferSize))
}

func (ls *LightStreamerConnection) UnsubscribeTradeUpdates(tickChan <-chan TradeUpdate) error {
	return unsubscribe(ls, tickChan)
}

func (ls *LightStreamerConnection) SubscribeChartTicks(ctx context.Context, bufferSize int, epics ...string) (<-chan ChartTick, error) {
	if ls.closeRequested.Load() {
		return nil, fmt.Errorf("cannot subscribe using a closed lightstreamer connection")
	}
	return subscribe(ls, ctx, makeNewSubscription[ChartTick](epics, bufferSize))
}

func (ls *LightStreamerConnection) UnsubscribeChartTicks(tickChan <-chan ChartTick) error {
	return unsubscribe(ls, tickChan)
}

func (ls *LightStreamerConnection) SubscribeMarkets(ctx context.Context, bufferSize int, epics ...string) (<-chan MarketTick, error) {
	if ls.closeRequested.Load() {
		return nil, fmt.Errorf("cannot subscribe using a closed lightstreamer connection")
	}

	return subscribe(ls, ctx, makeNewSubscription[MarketTick](epics, bufferSize))
}

func (ls *LightStreamerConnection) UnsubscribeMarkets(tickChan <-chan MarketTick) error {
	return unsubscribe(ls, tickChan)
}

func (ls *LightStreamerConnection) SubscribeChartCandles(ctx context.Context, bufferSize int, scale string, epics ...string) (<-chan ChartCandle, error) {
	if ls.closeRequested.Load() {
		return nil, fmt.Errorf("cannot subscribe using a closed lightstreamer connection")
	}

	subReq := makeNewSubscription[ChartCandle](epics, bufferSize)
	subReq.scale = scale
	return subscribe(ls, ctx, subReq)
}

func (ls *LightStreamerConnection) UnsubscribeChartCandles(candleChan <-chan ChartCandle) error {
	return unsubscribe(ls, candleChan)
}

func makeNewSubscription[T MarketTick | ChartTick | ChartCandle | TradeUpdate](items []string, bufferSize int) subscription[T] {
	subReq := subscription[T]{
		channel:         make(chan T, bufferSize),
		items:           items,
		errChan:         make(chan error, 1),
		lastStateByItem: make(map[string]T),
	}
	return subReq
}

func subscribe[T MarketTick | ChartTick | ChartCandle | TradeUpdate](ls *LightStreamerConnection, ctx context.Context, subReq subscription[T]) (<-chan T, error) {
	select {
	case ls.subscriptionReqChan <- subReq:
		select {
		case err := <-subReq.errChan:
			if err != nil {
				return nil, err
			}
			return subReq.channel, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func unsubscribe[T MarketTick | ChartTick | ChartCandle | TradeUpdate](ls *LightStreamerConnection, tickChan <-chan T) error {
	if ls.closeRequested.Load() {
		return fmt.Errorf("cannot unsubscribe using a closed lightstreamer connection")
	}
	unsubReq := unsubscription[T]{
		channel: tickChan,
		errChan: make(chan error),
	}
	ls.subscriptionReqChan <- unsubReq
	return <-unsubReq.errChan
}

func (ls *LightStreamerConnection) Close() error {
	ls.closeRequested.Store(true)
	ls.cancelFunc()

	defer func() {
		for _, sub := range ls.marketSubscriptions {
			close(sub.channel)
		}
		for _, sub := range ls.chartTickSubscriptions {
			close(sub.channel)
		}
		for _, sub := range ls.chartCandleSubscriptions {
			close(sub.channel)
		}
		ls.chartTickSubscriptions = nil
		ls.marketSubscriptions = nil
		ls.chartCandleSubscriptions = nil
	}()

	err := ls.wsConn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, ""), time.Time{})

	if err != nil {
		return err
	}
	err = ls.wsConn.Close()

	if err != nil {
		return err
	}

	return nil
}

const lightStreamerContentType = "application/x-www-form-urlencoded"

// connectToLightStream - Create new lightstreamer session
func (ig *IGMarkets) lightStreamerConnect(client *http.Client, sessionVersion2 *SessionVersion2) (sessionID, sessionMsg string, err error) {
	createVals := url.Values{}
	createVals.Set("LS_op", "create")
	createVals.Set("LS_cid", "mgQkwtwdysogQz2BJ4Ji kOj2Bg")
	createVals.Set("LS_adapter_set", "DEFAULT")
	createVals.Set("LS_user", sessionVersion2.CurrentAccountId)
	createVals.Set("LS_password", "CST-"+sessionVersion2.CSTToken+"|XST-"+sessionVersion2.XSTToken)
	createVals.Set("LS_polling", "true")
	createVals.Set("LS_polling_millis", "0")
	createVals.Set("LS_idle_millis", "0")
	bodyBuf := strings.NewReader(createVals.Encode())
	url := fmt.Sprintf("%s/lightstreamer/create_session.txt", sessionVersion2.LightstreamerEndpoint)
	resp, err := client.Post(url, lightStreamerContentType, bodyBuf)
	if err != nil {
		if resp != nil {
			body, err2 := io.ReadAll(resp.Body)
			if err2 != nil {
				return "", "", fmt.Errorf("calling lightstreamer endpoint %s failed: %v; reading HTTP body also failed: %v",
					url, err, err2)
			}
			return "", "", fmt.Errorf("calling lightstreamer endpoint %s failed: %v http.StatusCode:%d Body: %q",
				url, err, resp.StatusCode, string(body))
		}
		return "", "", fmt.Errorf("calling lightstreamer endpoint %q failed: %v", url, err)
	}
	respBody, _ := io.ReadAll(resp.Body)
	sessionMsg = string(respBody[:])
	if !strings.HasPrefix(sessionMsg, "OK") {
		return "", "", fmt.Errorf("unexpected response from lightstreamer session endpoint %q: %q", url, sessionMsg)
	}
	sessionParts := strings.Split(sessionMsg, "\r\n")
	sessionID = sessionParts[1]
	sessionID = strings.ReplaceAll(sessionID, "SessionId:", "")
	return sessionID, sessionMsg, nil
}

func (ig *IGMarkets) lightstreamerReadSubscription(epics []string, tickReceiver chan MarketTick, resp *http.Response) {
	const epicNameUnknown = "unknown"
	var lastTicks = make(map[string]MarketTick, len(epics)) // epic -> tick

	defer close(tickReceiver)

	// map table index -> epic name
	var epicIndex = make(map[string]string, len(epics))
	for i, epic := range epics {
		epicIndex[fmt.Sprintf("1,%d", i+1)] = epic
	}

	splitFunc := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := bytes.Index(data, []byte("\r\n")); i >= 0 {
			return i + 2, data[0:i], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return
	}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(splitFunc)
	scanner.Buffer(make([]byte, 16384), 16384)

	var err error
	for {
		var priceMsg string
		if scanner.Scan() {
			priceMsg = scanner.Text()
		} else {
			err := scanner.Err()
			if err == nil {
				break
			} else {
				fmt.Printf("reading lightstreamer subscription failed: %v", err)
			}
		}
		priceParts := strings.Split(priceMsg, "|")

		// Server ends streaming
		if priceMsg == "LOOP" {
			fmt.Printf("ending\n")
			break
		}

		if len(priceParts) < 5 {
			fmt.Printf("Malformed price message: %q\n", priceMsg)
			continue
		}

		var parsedTime time.Time
		if priceParts[1] != "" {
			priceTime := priceParts[1]
			now := time.Now().In(ig.TimeZoneLightStreamer)
			parsedTime, err = time.ParseInLocation("2006-1-2 15:04:05", fmt.Sprintf("%d-%d-%d %s",
				now.Year(), now.Month(), now.Day(), priceTime), ig.TimeZoneLightStreamer)
			if err != nil {
				fmt.Printf("parsing time failed: %v time=%q\n", err, priceTime)
				continue
			}
		}
		tableIndex := priceParts[0]
		priceBid, _ := strconv.ParseFloat(priceParts[2], 64)
		priceAsk, _ := strconv.ParseFloat(priceParts[3], 64)

		epic, found := epicIndex[tableIndex]
		if !found {
			fmt.Printf("unknown epic %q\n", tableIndex)
			epic = epicNameUnknown
		}

		if epic != epicNameUnknown {
			var lastTick, found = lastTicks[epic]
			if found {
				if priceAsk == 0 {
					priceAsk = lastTick.Ask
				}
				if priceBid == 0 {
					priceBid = lastTick.Bid
				}
				if parsedTime.IsZero() {
					parsedTime = lastTick.Time
				}
			}
		}

		tick := MarketTick{
			Epic: epic,
			Time: parsedTime,
			Bid:  priceBid,
			Ask:  priceAsk,
		}
		tickReceiver <- tick
		lastTicks[epic] = tick
	}
}
