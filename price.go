package igmarkets

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"
)

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

// GetPriceHistory - Return the minute prices for the last 10 minutes for the given epic.
func (ig *IGMarkets) GetPriceHistory(ctx context.Context, epic, resolution string, max int, from, to time.Time) (*PriceResponse, error) {
	bodyReq := new(bytes.Buffer)

	limitStr := ""
	if !to.IsZero() && !from.IsZero() {
		fromStr := from.Format("2006-01-02T15:04:05")
		toStr := to.Format("2006-01-02T15:04:05")
		limitStr = fmt.Sprintf("&from=%s&to=%s", fromStr, toStr)
	} else if max > 0 {
		limitStr = fmt.Sprintf("&max=%d", max)
	}

	page := "&max=100&pageSize=100"

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/gateway/deal/prices/%s?resolution=%s",
		ig.APIURL, epic, resolution)+limitStr+page, bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to get price: %v", err)
	}

	igResponseInterface, err := ig.doRequest(ctx, req, 3, PriceResponse{})
	if err != nil {
		return nil, err
	}
	igResponse, _ := igResponseInterface.(*PriceResponse)

	return igResponse, err
}

// GetPrice - Return the minute prices for the last 10 minutes for the given epic.
func (ig *IGMarkets) GetPrice(ctx context.Context, epic string) (*PriceResponse, error) {
	return ig.GetPriceHistory(ctx, epic, ResolutionSecond, 1, time.Time{}, time.Time{})
}
