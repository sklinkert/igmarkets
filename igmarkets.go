package igmarkets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// OTCPositionCloseRequest - request struct for closing positions
type OTCPositionCloseRequest struct {
	DealID      string  `json:"dealId,omitempty"`
	Direction   string  `json:"direction"` // "BUY" or "SELL"
	Epic        string  `json:"epic,omitempty"`
	Expiry      string  `json:"expiry,omitempty"`
	Level       string  `json:"level,omitempty"`
	OrderType   string  `json:"orderType"`
	QuoteID     string  `json:"quoteId,omitempty"`
	Size        float64 `json:"size"`                  // Deal size
	TimeInForce string  `json:"timeInForce,omitempty"` // "EXECUTE_AND_ELIMINATE" or "FILL_OR_KILL"
}

// OTCOrderRequest - request struct for placing orders
type OTCOrderRequest struct {
	Epic                  string  `json:"epic"`
	Level                 string  `json:"level,omitempty"`
	ForceOpen             bool    `json:"forceOpen"`
	OrderType             string  `json:"orderType"`
	CurrencyCode          string  `json:"currencyCode"`
	Direction             string  `json:"direction"` // "BUY" or "SELL"
	Expiry                string  `json:"expiry"`
	Size                  float64 `json:"size"` // Deal size
	StopDistance          string  `json:"stopDistance,omitempty"`
	StopLevel             string  `json:"stopLevel,omitempty"`
	LimitDistance         string  `json:"limitDistance,omitempty"`
	LimitLevel            string  `json:"limitLevel,omitempty"`
	QuoteID               string  `json:"quoteId,omitempty"`
	TimeInForce           string  `json:"timeInForce,omitempty"` // "EXECUTE_AND_ELIMINATE" or "FILL_OR_KILL"
	TrailingStop          bool    `json:"trailingStop"`
	TrailingStopIncrement string  `json:"trailingStopIncrement,omitempty"`
	GuaranteedStop        bool    `json:"guaranteedStop"`
	DealReference         string  `json:"dealReference,omitempty"`
}

// OTCUpdateOrderRequest - request struct for updating orders
type OTCUpdateOrderRequest struct {
	StopLevel             float64 `json:"stopLevel"`
	LimitLevel            float64 `json:"limitLevel"`
	TrailingStop          bool    `json:"trailingStop"`
	TrailingStopIncrement string  `json:"trailingStopIncrement,omitempty"`
}

// OTCWorkingOrderRequest - request struct for placing workingorders
type OTCWorkingOrderRequest struct {
	CurrencyCode   string  `json:"currencyCode"`
	DealReference  string  `json:"dealReference,omitempty"`
	Direction      string  `json:"direction"` // "BUY" or "SELL"
	Epic           string  `json:"epic"`
	Expiry         string  `json:"expiry"`
	ForceOpen      bool    `json:"forceOpen"`
	GoodTillDate   string  `json:"goodTillDate,omitempty"`
	GuaranteedStop bool    `json:"guaranteedStop"`
	Level          float64 `json:"level"`
	LimitDistance  string  `json:"limitDistance,omitempty"`
	LimitLevel     string  `json:"limitLevel,omitempty"`
	Size           float64 `json:"size"` // Deal size
	StopDistance   string  `json:"stopDistance,omitempty"`
	StopLevel      string  `json:"stopLevel,omitempty"`
	TimeInForce    string  `json:"timeInForce,omitempty"` // "GOOD_TILL_CANCELLED", "GOOD_TILL_DATE"
	Type           string  `json:"type"`
}

// OTCWorkingOrder - Working order
type OTCWorkingOrder struct {
	MarketData       MarketData       `json:"marketData"`
	WorkingOrderData WorkingOrderData `json:"workingOrderData"`
}

// MarketData - Subset of OTCWorkingOrder
type MarketData struct {
	Bid                      float64 `json:"bid"`
	DelayTime                int     `json:"delayTime"`
	Epic                     string  `json:"epic"`
	ExchangeID               string  `json:"exchangeId"`
	Expiry                   string  `json:"expiry"`
	High                     float64 `json:"high"`
	InstrumentName           string  `json:"instrumentName"`
	InstrumentType           string  `json:"instrumentType"`
	LotSize                  float64 `json:"lotSize"`
	Low                      float64 `json:"low"`
	MarektStatus             string  `json:"marketStatus"`
	NetChange                float64 `json:"netChange"`
	Offer                    float64 `json:"offer"`
	PercentageChange         float64 `json:"percentageChange"`
	ScalingFactor            int     `json:"scalingFactor"`
	StreamingPricesAvailable bool    `json:"streamingPricesAvailable"`
	UpdateTime               string  `json:"updateTime"`
	UpdateTimeUTC            string  `json:"updateTimeUTC"`
}

// WorkingOrderData - Subset of OTCWorkingOrder
type WorkingOrderData struct {
	CreatedDate     string  `json:"createdDate"`
	CreatedDateUTC  string  `json:"createdDateUTC"`
	CurrencyCode    string  `json:"currencyCode"`
	DealID          string  `json:"dealId"`
	Direction       string  `json:"direction"` // "BUY" or "SELL"
	DMA             bool    `json:"dma"`
	Epic            string  `json:"epic"`
	GoodTillDate    string  `json:"goodTillDate"`
	GoodTillDateISO string  `json:"goodTillDateISO"`
	GuaranteedStop  bool    `json:"guaranteedStop"`
	LimitDistance   float64 `json:"limitDistance"`
	OrderLevel      float64 `json:"orderLevel"`
	OrderSize       float64 `json:"orderSize"` // Deal size
	OrderType       string  `json:"orderType"`
	StopDistance    float64 `json:"stopDistance"`
	TimeInForce     string  `json:"timeInForce,omitempty"` // "EXECUTE_AND_ELIMINATE" or "FILL_OR_KILL"
}

// PositionsResponse - Response from positions endpoint
type PositionsResponse struct {
	Positions []Position `json:"positions"`
}

// Position - part of PositionsResponse
type Position struct {
	MarketData MarketData `json:"market"`
	Position   struct {
		ContractSize         float64 `json:"contractSize"`
		ControlledRisk       bool    `json:"controlledRisk"`
		CreatedDate          string  `json:"createdDate"`
		CreatedDateUTC       string  `json:"createdDateUTC"`
		Currencry            string  `json:"currency"`
		DealID               string  `json:"dealId"`
		DealReference        string  `json:"dealReference"`
		Direction            string  `json:"direction"`
		Level                float64 `json:"level"`
		LimitLevel           float64 `json:"limitLevel"`
		Size                 float64 `json:"size"`
		StopLevel            float64 `json:"stopLevel"`
		TrailingStep         float64 `json:"trailingStep"`
		TrailingStopDistance float64 `json:"trailingStopDistance"`
	} `json:"position"`
}

// HistoryTransactionResponse - Response for  transactions endpoint
type HistoryTransactionResponse struct {
	Transactions []Transaction `json:"transactions"`
	MetaData     struct {
		PageData struct {
			PageNumber int `json:"pageNumber"`
			PageSize   int `json:"pageSize"`
			TotalPages int `json:"totalPages"`
		} `json:"pageData"`
		Size int `json:"size"`
	} `json:"metaData"`
}

// Transaction - Part of HistoryTransactionResponse
type Transaction struct {
	CashTransaction bool   `json:"cashTransaction"`
	CloseLevel      string `json:"closeLevel"`
	Currency        string `json:"currency"`
	Date            string `json:"date"`
	DateUTC         string `json:"dateUtc"`
	InstrumentName  string `json:"instrumentName"`
	OpenDateUtc     string `json:"openDateUtc"`
	OpenLevel       string `json:"openLevel"`
	Period          string `json:"period"`
	ProfitAndLoss   string `json:"profitAndLoss"`
	Reference       string `json:"reference"`
	Size            string `json:"size"`
	TransactionType string `json:"transactionType"`
}

// AffectedDeal - part of order confirmation
type AffectedDeal struct {
	DealID   string `json:"dealId"`
	Constant string `json:"constant"` // "FULLY_CLOSED"
}

// DealReference - deal reference struct for responses
type DealReference struct {
	DealReference string `json:"dealReference"`
}

// OTCDealConfirmation - Deal confirmation
type OTCDealConfirmation struct {
	Epic                  string         `json:"epic"`
	AffectedDeals         []AffectedDeal `json:"affectedDeals"`
	Level                 float64        `json:"level"`
	ForceOpen             bool           `json:"forceOpen"`
	DealStatus            string         `json:"dealStatus"`
	Reason                string         `json:"reason"`
	Status                string         `json:"status"`
	OrderType             string         `json:"orderType"`
	Profit                float64        `json:"profit"`
	ProfitCurrency        string         `json:"profitCurrency"`
	CurrencyCode          string         `json:"currencyCode"`
	Direction             string         `json:"direction"` // "BUY" or "SELL"
	Expiry                string         `json:"expiry,omitempty"`
	Size                  float64        `json:"size"` // Deal size
	StopDistance          float64        `json:"stopDistance"`
	StopLevel             float64        `json:"stopLevel"`
	LimitDistance         string         `json:"limitDistance,omitempty"`
	LimitLevel            float64        `json:"limitLevel"`
	QuoteID               string         `json:"quoteId,omitempty"`
	TimeInForce           string         `json:"timeInForce,omitempty"` // "EXECUTE_AND_ELIMINATE" or "FILL_OR_KILL"
	TrailingStop          bool           `json:"trailingStop"`
	TrailingStopIncrement float64        `json:"trailingIncrement"`
	GuaranteedStop        bool           `json:"guaranteedStop"`
	DealReference         string         `json:"dealReference,omitempty"`
}

// AuthRequest - Encapsualates the real auth request object
type AuthRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// AuthResponse - Auth response
type AuthResponse struct {
	AccountID             string     `json:"accountId"`
	ClientID              string     `json:"clientId"`
	LightstreamerEndpoint string     `json:"lightstreamerEndpoint"`
	OAuthToken            OAuthToken `json:"oauthToken"`
	TimezoneOffset        int        `json:"timezoneOffset"` // In seconds
}

// OAuthToken - part of the AuthResponse
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// PriceResponse - Response for price query
type PriceResponse struct {
	Prices []struct {
		SnapshotTime     string `json:"snapshotTime"`
		SnapshotTimeUTC  string `json:"snapshotTimeUTC"`
		OpenPrice        Price  `json:"openPrice"`
		LowPrice         Price  `json:"lowPrice"`
		HighPrice        Price  `json:"highPrice"`
		ClosePrice       Price  `json:"closePrice"`
		LastTradedVolume int    `json:"lastTradedVolume"`
	}
	InstrumentType string   `json:"instrumentType"`
	MetaData       struct{} `json:"-"`
}

// Price - Subset of PriceResponse
type Price struct {
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
	LastTraded float64 `json:"lastTraded"` // Last traded price
}

const (
	// ResolutionSecond - 1 second price snapshot
	ResolutionSecond = "SECOND"
	// ResolutionMinute - 1 minute price snapshot
	ResolutionMinute = "MINUTE"
	// ResolutionHour - 1 hour price snapshot
	ResolutionHour = "HOUR"
	// ResolutionTwoHour - 2 hour price snapshot
	ResolutionTwoHour = "HOUR_2"
	// ResolutionThreeHour - 3 hour price snapshot
	ResolutionThreeHour = "HOUR_3"
	// ResolutionFourHour - 4 hour price snapshot
	ResolutionFourHour = "HOUR_4"
	// ResolutionDay - 1 day price snapshot
	ResolutionDay = "DAY"
	// ResolutionWeek - 1 week price snapshot
	ResolutionWeek = "WEEK"
	// ResolutionMonth - 1 month price snapshot
	ResolutionMonth = "MONTH"

	// DemoAPIURL - Demo API URL
	DemoAPIURL = "https://demo-api.ig.com"
	// LiveAPIURL - Live API URL - Real trading!
	LiveAPIURL = "https://api.ig.com"
)

// IGMarkets - Object with all information we need to access IG REST API
type IGMarkets struct {
	APIURL                string
	APIKey                string
	AccountID             string
	Identifier            string
	Password              string
	AutomaticTokenRefresh bool
	OAuthToken            OAuthToken
	Timeout               time.Duration // HTTP Timeout
	Lock                  sync.RWMutex
}

// New - Create new instance of igmarkets
func New(apiURL, apiKey, accountID, identifier, password string, automaticTokenRefresh bool, httpTimeout time.Duration) *IGMarkets {
	return &IGMarkets{
		APIURL:                apiURL,
		APIKey:                apiKey,
		AccountID:             accountID,
		AutomaticTokenRefresh: automaticTokenRefresh,
		Identifier:            identifier,
		Password:              password,
		Timeout:               httpTimeout,
	}
}

// Login - Get new OAuthToken from API and set it to IGMarkets object
func (ig *IGMarkets) Login() error {
	bodyReq := new(bytes.Buffer)

	var authReq = AuthRequest{
		Identifier: ig.Identifier,
		Password:   ig.Password,
	}

	if err := json.NewEncoder(bodyReq).Encode(authReq); err != nil {
		return fmt.Errorf("igmarkets: unable to encode JSON response: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", ig.APIURL, "gateway/deal/session"), bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "3")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		if ig.AutomaticTokenRefresh {
			go func() {
				time.Sleep(time.Duration(5) * time.Second)
				ig.Login()
			}()
		}
		return fmt.Errorf("igmarkets: unexpected error while sending HTTP request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if ig.AutomaticTokenRefresh {
			go func() {
				time.Sleep(time.Duration(5) * time.Second)
				ig.Login()
			}()
		}
		return fmt.Errorf("igmarkets: unexpected error while reading body from HTTP request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		if ig.AutomaticTokenRefresh {
			go func() {
				time.Sleep(time.Duration(5) * time.Second)
				ig.Login()
			}()
		}
		return fmt.Errorf("igmarkets: unexpected HTTP status code %d", resp.StatusCode)
	}

	authResponse := AuthResponse{}
	err = json.Unmarshal(body, &authResponse)
	if err != nil {
		if ig.AutomaticTokenRefresh {
			go func() {
				time.Sleep(time.Duration(5) * time.Second)
				ig.Login()
			}()
		}
		return fmt.Errorf("igmarkets: unable to unmarshal json response: %v", err)
	}

	if authResponse.OAuthToken.AccessToken == "" {
		if ig.AutomaticTokenRefresh {
			go func() {
				time.Sleep(time.Duration(5) * time.Second)
				ig.Login()
			}()
		}
		return fmt.Errorf("igmarkets: got response but access token is empty")
	}

	expiry, err := strconv.ParseInt(authResponse.OAuthToken.ExpiresIn, 10, 32)
	if err != nil {
		if ig.AutomaticTokenRefresh {
			go func() {
				time.Sleep(time.Duration(5) * time.Second)
				ig.Login()
			}()
		}
		return fmt.Errorf("igmarkets: unable to parse OAuthToken expiry field: %v", err)
	}

	// Refresh token before it will expire
	if expiry <= 10 {
		if ig.AutomaticTokenRefresh {
			go func() {
				time.Sleep(time.Duration(5) * time.Second)
				ig.Login()
			}()
		}
		return fmt.Errorf("igmarkets: token expiry is too short for periodically renewals")
	}

	if ig.AutomaticTokenRefresh {
		go func() {
			time.Sleep(time.Duration((expiry - 7)) * time.Second)
			ig.Login()
		}()
	}

	ig.Lock.Lock()
	ig.OAuthToken = authResponse.OAuthToken
	ig.Lock.Unlock()

	return nil
}

// GetPrice - Return the minute prices for the last 10 minutes for the given epic.
func (ig *IGMarkets) GetPrice(epic string) (PriceResponse, error) {
	return ig.GetPriceHistory(epic, ResolutionSecond, 1, time.Time{}, time.Time{})
}

// GetTransactions - Return all transaction
func (ig *IGMarkets) GetTransactions(transactionType string, from time.Time) (HistoryTransactionResponse, error) {
	bodyReq := new(bytes.Buffer)
	transactionResp := HistoryTransactionResponse{}
	fromStr := from.Format("2006-01-02T15:04:05")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gateway/deal/history/transactions?from=%s&type=%s",
		DemoAPIURL, fromStr, transactionType), bodyReq)
	if err != nil {
		return transactionResp, fmt.Errorf("igmarkets: unable to get transactions: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "2")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return transactionResp, fmt.Errorf("igmarkets: unable to get transactions: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return transactionResp,
			fmt.Errorf("igmarkets: unable to get body of transactions request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return transactionResp,
			fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}
	if err := json.Unmarshal(body, &transactionResp); err != nil {
		return transactionResp, fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
	}

	return transactionResp, nil
}

// GetPriceHistory - Return the minute prices for the last 10 minutes for the given epic.
func (ig *IGMarkets) GetPriceHistory(epic, resolution string, max int, from, to time.Time) (PriceResponse, error) {
	bodyReq := new(bytes.Buffer)
	priceResp := PriceResponse{}

	limitStr := ""
	if !to.IsZero() && !from.IsZero() {
		fromStr := from.Format("2006-01-02T15:04:05")
		toStr := to.Format("2006-01-02T15:04:05")
		limitStr = fmt.Sprintf("&from=%s&to=%s", fromStr, toStr)
	} else if max > 0 {
		limitStr = fmt.Sprintf("&max=%d", max)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gateway/deal/prices/%s?resolution=%s",
		DemoAPIURL, epic, resolution)+limitStr, bodyReq)
	if err != nil {
		return priceResp, fmt.Errorf("igmarkets: unable to get price: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "3")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return priceResp, fmt.Errorf("igmarkets: unable to get price: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return priceResp, fmt.Errorf("igmarkets: unable to read HTTP body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return priceResp,
			fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}
	if err := json.Unmarshal(body, &priceResp); err != nil {
		return priceResp, fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
	}

	return priceResp, nil
}

// PlaceOTCOrder - Place an OTC order
func (ig *IGMarkets) PlaceOTCOrder(order OTCOrderRequest) (string, error) {
	bodyReq, err := json.Marshal(&order)
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}
	req, err := http.NewRequest("POST", DemoAPIURL+"/gateway/deal/positions/otc", bytes.NewReader(bodyReq))
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot create HTTP request: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "2")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to read HTTP body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}
	ref := DealReference{}
	if err := json.Unmarshal(body, &ref); err != nil {
		return "", fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
	}

	return ref.DealReference, nil
}

// UpdateOTCOrder - Update an exisiting OTC order
func (ig *IGMarkets) UpdateOTCOrder(dealID string, order OTCUpdateOrderRequest) (string, error) {
	bodyReq, err := json.Marshal(&order)
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}
	req, err := http.NewRequest("PUT", DemoAPIURL+"/gateway/deal/positions/otc/"+dealID, bytes.NewReader(bodyReq))
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot create HTTP request: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "2")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to read HTTP body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}
	ref := DealReference{}
	if err := json.Unmarshal(body, &ref); err != nil {
		return "", fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
	}

	return ref.DealReference, nil
}

// CloseOTCPosition - Close an OTC position
func (ig *IGMarkets) CloseOTCPosition(close OTCPositionCloseRequest) (string, error) {
	bodyReq, err := json.Marshal(&close)
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}
	req, err := http.NewRequest("POST", DemoAPIURL+"/gateway/deal/positions/otc", bytes.NewReader(bodyReq))
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot create HTTP request: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "1")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	req.Header.Set("_method", "DELETE")
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to read HTTP body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}
	ref := DealReference{}
	if err := json.Unmarshal(body, &ref); err != nil {
		return "", fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
	}

	return ref.DealReference, nil
}

// GetDealConfirmation - Check if the given order was closed/filled
func (ig *IGMarkets) GetDealConfirmation(dealRef string) (OTCDealConfirmation, error) {
	dealConfirmation := OTCDealConfirmation{}
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", DemoAPIURL+"/gateway/deal/confirms/"+dealRef, bodyReq)
	if err != nil {
		return dealConfirmation, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "1")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return dealConfirmation, fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return dealConfirmation, fmt.Errorf("igmarkets: unable to read HTTP body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return dealConfirmation, fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}
	if err := json.Unmarshal(body, &dealConfirmation); err != nil {
		return dealConfirmation, fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
	}

	return dealConfirmation, nil
}

// GetPositions - Get all open positions
func (ig *IGMarkets) GetPositions() (PositionsResponse, error) {
	positions := PositionsResponse{}
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", DemoAPIURL+"/gateway/deal/positions/", bodyReq)
	if err != nil {
		return positions, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "2")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return positions, fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return positions, fmt.Errorf("igmarkets: unable to read HTTP body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return positions, fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}
	err = json.Unmarshal(body, &positions)
	if err != nil {
		return positions, fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
	}

	return positions, nil
}

// DeleteOTCOrder - Delete order
func (ig *IGMarkets) DeleteOTCOrder(dealRef string) error {
	dealConfirmation := OTCDealConfirmation{}
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("DELETE", DemoAPIURL+"/gateway/deal/positions/otc", bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "1")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to read HTTP body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}
	err = json.Unmarshal(body, &dealConfirmation)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
	}

	return nil
}

// PlaceOTCWorkingOrder - Place an OTC workingorder
func (ig *IGMarkets) PlaceOTCWorkingOrder(order OTCWorkingOrderRequest) (dealRef string, err error) {
	bodyReq, err := json.Marshal(&order)
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to marshal JSON: %v", err)
	}
	req, err := http.NewRequest("POST", DemoAPIURL+"/gateway/deal/workingorders/otc", bytes.NewReader(bodyReq))
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "2")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to read HTTP body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}
	ref := DealReference{}
	if err := json.Unmarshal(body, &ref); err != nil {
		return "", fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
	}

	return ref.DealReference, nil
}

// GetOTCWorkingOrders - Get all working orders
func (ig *IGMarkets) GetOTCWorkingOrders() (orders []OTCWorkingOrder, err error) {
	bodyReq := new(bytes.Buffer)
	req, err := http.NewRequest("GET", DemoAPIURL+"/gateway/deal/workingorders/", bodyReq)
	if err != nil {
		return orders, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "2")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return orders, fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return orders, fmt.Errorf("igmarkets: unable to read HTTP body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return orders, fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}
	type WorkingOrders struct {
		WorkingOrders []OTCWorkingOrder `json:"workingOrders"`
	}
	wo := WorkingOrders{}
	if err := json.Unmarshal(body, &wo); err != nil {
		return orders, fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
	}

	return wo.WorkingOrders, nil
}

// DeleteOTCWorkingOrder - Delete workingorder
func (ig *IGMarkets) DeleteOTCWorkingOrder(dealRef string) error {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("DELETE", DemoAPIURL+"/gateway/deal/workingorders/otc/"+dealRef, bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	ig.Lock.RLock()
	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", "2")
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.Lock.RUnlock()

	client := &http.Client{
		Timeout: ig.Timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}

	return nil
}
