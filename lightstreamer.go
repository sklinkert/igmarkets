package igmarkets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LightStreamerTick struct {
	Epic string
	Time time.Time
	Bid  float64
	Ask  float64
}

// GetOTCWorkingOrders - Get all working orders
// epic: e.g. CS.D.BITCOIN.CFD.IP
// tickReceiver: receives all ticks from lightstreamer API
func (ig *IGMarkets) OpenLightStreamerSubscription(epics []string, tickReceiver chan LightStreamerTick) error {
	const contentType = "application/x-www-form-urlencoded"

	// Obtain CST and XST tokens first
	sessionVersion2, err := ig.LoginVersion2()
	if err != nil {
		return fmt.Errorf("ig.LoginVersion2() failed: %v", err)
	}

	tr := &http.Transport{
		MaxIdleConns:       1,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	c := &http.Client{Transport: tr}

	// Create Lightstreamer Session
	body := []byte("LS_polling=true&LS_polling_millis=0&LS_idle_millis=0&LS_op2=create&LS_password=CST-" +
		sessionVersion2.CSTToken + "|" + "XST-" + sessionVersion2.XSTToken + "&LS_user=" +
		sessionVersion2.CurrentAccountId + "&LS_cid=mgQkwtwdysogQz2BJ4Ji kOj2Bg")
	bodyBuf := bytes.NewBuffer(body)
	url := fmt.Sprintf("%s/lightstreamer/create_session.txt", sessionVersion2.LightstreamerEndpoint)
	resp, err := c.Post(url, contentType, bodyBuf)
	if err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("calling lightstreamer endpoint %s failed: %v http.StatusCode:%d Body: %q",
			url, err, resp.StatusCode, string(body))
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	sessionMsg := string(respBody[:])
	if !strings.HasPrefix(sessionMsg, "OK") {
		return fmt.Errorf("unexpected response from lightstreamer session endpoint %q: %q", url, sessionMsg)
	}
	sessionParts := strings.Split(sessionMsg, "\r\n")
	sessionID := sessionParts[1]
	sessionID = strings.ReplaceAll(sessionID, "SessionId:", "")

	// Adding subscription for epic
	var epicList string
	for i := range epics {
		epicList = epicList + "MARKET:" + epics[i] + "+"
	}
	body = []byte("LS_session=" + sessionID +
		"&LS_polling=true&LS_polling_millis=0&LS_idle_millis=0&LS_op=add&LS_Table=1&LS_id=" +
		epicList + "&LS_schema=UPDATE_TIME+BID+OFFER+MARKET_STATE&LS_mode=MERGE")
	bodyBuf = bytes.NewBuffer(body)
	url = fmt.Sprintf("%s/lightstreamer/control.txt", sessionVersion2.LightstreamerEndpoint)
	resp, err = c.Post(url, contentType, bodyBuf)
	if err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("calling lightstreamer endpoint %q failed: %v http.StatusCode:%d Body: %q",
			url, err, resp.StatusCode, string(body))
	}
	body, _ = ioutil.ReadAll(resp.Body)
	if !strings.HasPrefix(sessionMsg, "OK") {
		return fmt.Errorf("unexpected control.txt response: %q", body)
	}

	// Binding to subscription
	body = []byte("LS_session=" + sessionID + "&LS_polling=false&LS_polling_millis=0&LS_idle_millis=0")
	bodyBuf = bytes.NewBuffer(body)
	url = fmt.Sprintf("%s/lightstreamer/bind_session.txt", sessionVersion2.LightstreamerEndpoint)
	resp, err = c.Post(url, contentType, bodyBuf)
	if err != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("calling lightstreamer endpoint %q failed: %v http.StatusCode:%d Body: %q",
			url, err, resp.StatusCode, string(body))
	}
	go readLightStreamSubscription(epics, tickReceiver, resp)
	return nil
}

func readLightStreamSubscription(epics []string, tickReceiver chan LightStreamerTick, resp *http.Response) {
	var respBuf = make([]byte, 64)
	var priceTime string
	var priceBid, priceAsk float64

	defer close(tickReceiver)

	// map table index -> epic name
	var epicIndex = make(map[string]string)
	for i, epic := range epics {
		epicIndex[fmt.Sprintf("1,%d", i+1)] = epic
	}

	for {
		read, err := resp.Body.Read(respBuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("reading lightstreamer subscription failed: %v", err)
			break
		}

		priceMsg := string(respBuf[0:read])
		priceParts := strings.Split(priceMsg, "|")

		// Sever ends streaming
		if priceMsg == "LOOP\r\n\r\n" {
			fmt.Printf("ending\n")
			break
		}

		if len(priceParts) != 5 {
			//fmt.Printf("Malformed price message: %q\n", priceMsg)
			continue
		}

		if priceParts[1] != "" {
			priceTime = priceParts[1]
		}
		tableIndex := priceParts[0]
		priceBid, _ = strconv.ParseFloat(priceParts[2], 64)
		priceAsk, _ = strconv.ParseFloat(priceParts[3], 64)
		now := time.Now().UTC()
		parsedTime, _ := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%d-%d %s",
			now.Year(), now.Month(), now.Day(), priceTime), time.UTC)

		epic, found := epicIndex[tableIndex]
		if !found {
			epic = "unknown"
		}

		tick := LightStreamerTick{
			Epic: epic,
			Time: parsedTime,
			Bid:  priceBid,
			Ask:  priceAsk,
		}
		tickReceiver <- tick
	}
}

// LoginVersion2 - use old login version. contains required data for LightStreamer API
func (ig *IGMarkets) LoginVersion2() (*SessionVersion2, error) {
	bodyReq := new(bytes.Buffer)

	var authReq = authRequest{
		Identifier: ig.Identifier,
		Password:   ig.Password,
	}

	if err := json.NewEncoder(bodyReq).Encode(authReq); err != nil {
		return nil, fmt.Errorf("igmarkets: unable to encode JSON response: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", ig.APIURL, "gateway/deal/session"), bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}

	igResponseInterface, headers, err := ig.doRequestWithResponseHeaders(req, 2, SessionVersion2{})
	if err != nil {
		return nil, err
	}
	session, _ := igResponseInterface.(*SessionVersion2)
	if headers != nil {
		session.CSTToken = headers.Get("CST")
		session.XSTToken = headers.Get("X-SECURITY-TOKEN")
	}
	return session, nil
}
