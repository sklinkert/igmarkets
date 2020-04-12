package igmarkets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
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

// WorkingOrders - Working orders
type WorkingOrders struct {
	WorkingOrders []OTCWorkingOrder `json:"workingOrders"`
}

// OTCWorkingOrder - Part of WorkingOrders
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

// MarketSearchResponse - Contains the response data for MarketSearch()
type MarketSearchResponse struct {
	Markets []MarketData `json:"markets"`
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

// DealingRules - Part of MarketsResponse
type DealingRules struct {
	MarketOrderPreference         string         `json:"marketOrderPreference"`
	TrailingStopsPreference       string         `json:"trailingStopsPreference"`
	MaxStopOrLimitDistance        UnitValueFloat `json:"maxStopOrLimitDistance"`
	MinControlledRiskStopDistance UnitValueFloat `json:"minControlledRiskStopDistance"`
	MinDealSize                   UnitValueFloat `json:"minDealSize"`
	MinNormalStopOrLimitDistance  UnitValueFloat `json:"minNormalStopOrLimitDistance"`
	MinStepDistance               UnitValueFloat `json:"minStepDistance"`
}

// Currency - Part of MarketsResponse
type Currency struct {
	BaseExchangeRate float64 `json:"baseExchangeRate"`
	Code             string  `json:"code"`
	ExchangeRate     float64 `json:"exchangeRate"`
	IsDefault        bool    `json:"isDefault"`
	Symbol           string  `json:"symbol"`
}

// UnitValueFloat - Part of MarketsResponse
type UnitValueFloat struct {
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
}

// Instrument - Part of MarketsResponse
type Instrument struct {
	ChartCode                string         `json:"chartCode"`
	ControlledRiskAllowed    bool           `json:"controlledRiskAllowed"`
	Country                  string         `json:"country"`
	Currencies               []Currency     `json:"currencies"`
	Epic                     string         `json:"epic"`
	Expiry                   string         `json:"expiry"`
	StreamingPricesAvailable bool           `json:"streamingPricesAvailable"`
	ForceOpenAllowed         bool           `json:"forceOpenAllowed"`
	Unit                     string         `json:"unit"`
	Type                     string         `json:"type"`
	MarketID                 string         `json:"marketID"`
	LotSize                  float64        `json:"lotSize"`
	MarginFactor             float64        `json:"marginFactor"`
	MarginFactorUnit         string         `json:"marginFactorUnit"`
	SlippageFactor           UnitValueFloat `json:"slippageFactor"`
	LimitedRiskPremium       UnitValueFloat `json:"limitedRiskPremium"`
	NewsCode                 string         `json:"newsCode"`
	ValueOfOnePip            string         `json:"valueOfOnePip"`
	OnePipMeans              string         `json:"onePipMeans"`
	ContractSize             string         `json:"contractSize"`
	SpecialInfo              []string       `json:"specialInfo"`
}

// Snapshot - Part of MarketsResponse
type Snapshot struct {
	MarketStatus              string  `json:"marketStatus"`
	NetChange                 float64 `json:"netChange"`
	PercentageChange          float64 `json:"percentageChange"`
	UpdateTime                string  `json:"updateTime"`
	DelayTime                 float64 `json:"delayTime"`
	Bid                       float64 `json:"bid"`
	Offer                     float64 `json:"offer"`
	High                      float64 `json:"high"`
	Low                       float64 `json:"low"`
	DecimalPlacesFactor       float64 `json:"decimalPlacesFactor"`
	ScalingFactor             float64 `json:"scalingFactor"`
	ControlledRiskExtraSpread float64 `json:"controlledRiskExtraSpread"`
}

// MarketsResponse - Marekt response for /markets/{epic}
type MarketsResponse struct {
	DealingRules DealingRules `json:"dealingRules"`
	Instrument   Instrument   `json:"instrument"`
	Snapshot     Snapshot     `json:"snapshot"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

// authResponse - IG auth response
type authResponse struct {
	AccountID             string     `json:"accountId"`
	ClientID              string     `json:"clientId"`
	LightstreamerEndpoint string     `json:"lightstreamerEndpoint"`
	OAuthToken            OAuthToken `json:"oauthToken"`
	TimezoneOffset        int        `json:"timezoneOffset"` // In seconds
}

// OAuthToken - part of the authResponse
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
	APIURL     string
	APIKey     string
	AccountID  string
	Identifier string
	Password   string
	OAuthToken OAuthToken
	httpClient *http.Client
	sync.RWMutex
}

// New - Create new instance of igmarkets
func New(apiURL, apiKey, accountID, identifier, password string, httpTimeout time.Duration) *IGMarkets {
	if apiURL != DemoAPIURL && apiURL != LiveAPIURL {
		log.Panic("Invalid endpoint URL", apiURL)
	}

	httpClient := &http.Client{
		Timeout: httpTimeout,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 5,
		},
	}

	return &IGMarkets{
		APIURL:     apiURL,
		APIKey:     apiKey,
		AccountID:  accountID,
		Identifier: identifier,
		Password:   password,
		httpClient: httpClient,
	}
}

// RefreshToken - Get new OAuthToken from API and set it to IGMarkets object
func (ig *IGMarkets) RefreshToken() error {
	bodyReq := new(bytes.Buffer)

	var authReq = refreshTokenRequest{
		RefreshToken: ig.OAuthToken.RefreshToken,
	}

	if err := json.NewEncoder(bodyReq).Encode(authReq); err != nil {
		return fmt.Errorf("igmarkets: unable to encode JSON response: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", ig.APIURL, "gateway/deal/session/refresh-token"), bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to send HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 1, OAuthToken{})
	if err != nil {
		return err
	}
	oauthToken, _ := igResponseInterface.(*OAuthToken)

	if oauthToken.AccessToken == "" {
		return fmt.Errorf("igmarkets: got response but access token is empty")
	}

	expiry, err := strconv.ParseInt(oauthToken.ExpiresIn, 10, 32)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to parse OAuthToken expiry field: %v", err)
	}

	// Refresh token before it will expire
	if expiry <= 10 {
		return fmt.Errorf("igmarkets: token expiry is too short for periodically renewals")
	}

	ig.Lock()
	ig.OAuthToken = *oauthToken
	ig.Unlock()

	return nil
}

// Login - Get new OAuthToken from API and set it to IGMarkets object
func (ig *IGMarkets) Login() error {
	bodyReq := new(bytes.Buffer)

	var authReq = authRequest{
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

	igResponseInterface, err := ig.doRequest(req, 3, authResponse{})
	if err != nil {
		return err
	}
	authResponse, _ := igResponseInterface.(*authResponse)

	if authResponse.OAuthToken.AccessToken == "" {
		return fmt.Errorf("igmarkets: got response but access token is empty")
	}

	expiry, err := strconv.ParseInt(authResponse.OAuthToken.ExpiresIn, 10, 32)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to parse OAuthToken expiry field: %v", err)
	}

	// Refresh token before it will expire
	if expiry <= 10 {
		return fmt.Errorf("igmarkets: token expiry is too short for periodically renewals")
	}

	ig.Lock()
	ig.OAuthToken = authResponse.OAuthToken
	ig.Unlock()

	return nil
}

// GetPrice - Return the minute prices for the last 10 minutes for the given epic.
func (ig *IGMarkets) GetPrice(epic string) (*PriceResponse, error) {
	return ig.GetPriceHistory(epic, ResolutionSecond, 1, time.Time{}, time.Time{})
}

// GetTransactions - Return all transaction
func (ig *IGMarkets) GetTransactions(transactionType string, from time.Time) (*HistoryTransactionResponse, error) {
	bodyReq := new(bytes.Buffer)
	fromStr := from.Format("2006-01-02T15:04:05")

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gateway/deal/history/transactions?from=%s&type=%s&pageSize=0",
		ig.APIURL, fromStr, transactionType), bodyReq)
	if err != nil {
		return &HistoryTransactionResponse{}, fmt.Errorf("igmarkets: unable to get transactions: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 2, HistoryTransactionResponse{})
	igResponse, _ := igResponseInterface.(*HistoryTransactionResponse)

	return igResponse, err
}

// GetPriceHistory - Return the minute prices for the last 10 minutes for the given epic.
func (ig *IGMarkets) GetPriceHistory(epic, resolution string, max int, from, to time.Time) (*PriceResponse, error) {
	bodyReq := new(bytes.Buffer)

	limitStr := ""
	if !to.IsZero() && !from.IsZero() {
		fromStr := from.Format("2006-01-02T15:04:05")
		toStr := to.Format("2006-01-02T15:04:05")
		limitStr = fmt.Sprintf("&from=%s&to=%s", fromStr, toStr)
	} else if max > 0 {
		limitStr = fmt.Sprintf("&max=%d", max)
	}

	page := "&max=1&pageSize=100"
	
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gateway/deal/prices/%s?resolution=%s",
		ig.APIURL, epic, resolution)+limitStr+page, bodyReq)
	if err != nil {
		return &PriceResponse{}, fmt.Errorf("igmarkets: unable to get price: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 3, PriceResponse{})
	igResponse, _ := igResponseInterface.(*PriceResponse)

	return igResponse, err
}

// PlaceOTCOrder - Place an OTC order
func (ig *IGMarkets) PlaceOTCOrder(order OTCOrderRequest) (string, error) {
	bodyReq, err := json.Marshal(&order)
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}
	req, err := http.NewRequest("POST", ig.APIURL+"/gateway/deal/positions/otc", bytes.NewReader(bodyReq))
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 2, DealReference{})
	igResponse, _ := igResponseInterface.(*DealReference)

	return igResponse.DealReference, err
}

// UpdateOTCOrder - Update an exisiting OTC order
func (ig *IGMarkets) UpdateOTCOrder(dealID string, order OTCUpdateOrderRequest) (string, error) {
	bodyReq, err := json.Marshal(&order)
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}
	req, err := http.NewRequest("PUT", ig.APIURL+"/gateway/deal/positions/otc/"+dealID, bytes.NewReader(bodyReq))
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 2, DealReference{})
	igResponse, _ := igResponseInterface.(*DealReference)

	return igResponse.DealReference, err
}

// CloseOTCPosition - Close an OTC position
func (ig *IGMarkets) CloseOTCPosition(close OTCPositionCloseRequest) (string, error) {
	bodyReq, err := json.Marshal(&close)
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot marshal: %v", err)
	}
	req, err := http.NewRequest("POST", ig.APIURL+"/gateway/deal/positions/otc", bytes.NewReader(bodyReq))
	if err != nil {
		return "", fmt.Errorf("igmarkets: cannot create HTTP request: %v", err)
	}

	req.Header.Set("_method", "DELETE")

	igResponseInterface, err := ig.doRequest(req, 1, DealReference{})
	igResponse, _ := igResponseInterface.(*DealReference)

	return igResponse.DealReference, err
}

// GetDealConfirmation - Check if the given order was closed/filled
func (ig *IGMarkets) GetDealConfirmation(dealRef string) (*OTCDealConfirmation, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", ig.APIURL+"/gateway/deal/confirms/"+dealRef, bodyReq)
	if err != nil {
		return &OTCDealConfirmation{}, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 1, OTCDealConfirmation{})
	igResponse, _ := igResponseInterface.(*OTCDealConfirmation)

	return igResponse, err
}

// GetPositions - Get all open positions
func (ig *IGMarkets) GetPositions() (*PositionsResponse, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", ig.APIURL+"/gateway/deal/positions/", bodyReq)
	if err != nil {
		return &PositionsResponse{}, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 2, PositionsResponse{})
	igResponse, _ := igResponseInterface.(*PositionsResponse)

	return igResponse, err
}

// DeletePositionsOTC - Closes one or more OTC positions
func (ig *IGMarkets) DeletePositionsOTC() error {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("DELETE", ig.APIURL+"/gateway/deal/positions/otc", bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	_, err = ig.doRequest(req, 1, nil)
	return err
}

// PlaceOTCWorkingOrder - Place an OTC workingorder
func (ig *IGMarkets) PlaceOTCWorkingOrder(order OTCWorkingOrderRequest) (string, error) {
	bodyReq, err := json.Marshal(&order)
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to marshal JSON: %v", err)
	}
	req, err := http.NewRequest("POST", ig.APIURL+"/gateway/deal/workingorders/otc", bytes.NewReader(bodyReq))
	if err != nil {
		return "", fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 2, DealReference{})
	igResponse, _ := igResponseInterface.(*DealReference)

	return igResponse.DealReference, err
}

// GetOTCWorkingOrders - Get all working orders
func (ig *IGMarkets) GetOTCWorkingOrders() (*WorkingOrders, error) {
	bodyReq := new(bytes.Buffer)
	req, err := http.NewRequest("GET", ig.APIURL+"/gateway/deal/workingorders/", bodyReq)
	if err != nil {
		return &WorkingOrders{}, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 2, WorkingOrders{})
	igResponse, _ := igResponseInterface.(*WorkingOrders)

	return igResponse, err
}

// DeleteOTCWorkingOrder - Delete workingorder
func (ig *IGMarkets) DeleteOTCWorkingOrder(dealRef string) error {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("DELETE", ig.APIURL+"/gateway/deal/workingorders/otc/"+dealRef, bodyReq)
	if err != nil {
		return fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	_, err = ig.doRequest(req, 2, nil)

	return err
}

// GetMarkets - Return markets information for given epic
func (ig *IGMarkets) GetMarkets(epic string) (*MarketsResponse, error) {
	bodyReq := new(bytes.Buffer)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gateway/deal/markets/%s",
		ig.APIURL, epic), bodyReq)
	if err != nil {
		return &MarketsResponse{}, fmt.Errorf("igmarkets: unable to get markets data: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 3, MarketsResponse{})
	igResponse, _ := igResponseInterface.(*MarketsResponse)

	return igResponse, err
}

// MarketSearch - Search for ISIN or share names to get the epic.
func (ig *IGMarkets) MarketSearch(term string) (*MarketSearchResponse, error) {
	bodyReq := new(bytes.Buffer)

	// E.g. https://demo-api.ig.com/gateway/deal/markets?searchTerm=DE0005008007
	url := fmt.Sprintf("%s/gateway/deal/markets?searchTerm=%s", ig.APIURL, term)
	req, err := http.NewRequest("GET", url, bodyReq)
	if err != nil {
		return &MarketSearchResponse{}, fmt.Errorf("igmarkets: unable to get markets data: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 1, MarketSearchResponse{})
	igResponse, _ := igResponseInterface.(*MarketSearchResponse)

	return igResponse, err
}

func (ig *IGMarkets) doRequest(req *http.Request, endpointVersion int, igResponse interface{}) (interface{}, error) {
	ig.RLock()
	req.Header.Set("X-IG-API-KEY", ig.APIKey)
	req.Header.Set("Authorization", "Bearer "+ig.OAuthToken.AccessToken)
	req.Header.Set("IG-ACCOUNT-ID", ig.AccountID)
	ig.RUnlock()

	req.Header.Set("Accept", "application/json; charset=UTF-8")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("VERSION", fmt.Sprintf("%d", endpointVersion))

	resp, err := ig.httpClient.Do(req)
	if err != nil {
		return igResponse, fmt.Errorf("igmarkets: unable to get markets data: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return igResponse, fmt.Errorf("igmarkets: unable to get body of transactions markets data: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return igResponse, fmt.Errorf("igmarkets: unexpected HTTP status code: %d", resp.StatusCode)
	}

	if igResponse != nil {
		objType := reflect.TypeOf(igResponse)
		obj := reflect.New(objType).Interface()
		if obj != nil {
			if err := json.Unmarshal(body, &obj); err != nil {
				return obj, fmt.Errorf("igmarkets: unable to unmarshal JSON response: %v", err)
			}

			return obj, nil
		}
	}

	return igResponse, nil
}
