package igmarkets

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ActivityResponse struct {
	Activities []Activity `json:"activities"`
}

type Activity struct {
	Channel     string `json:"channel"`     // The channel which triggered the activity. E.g. "WEB", "MOBILE", "DEALER"
	Date        string `json:"date"`        // The date the activity was triggered
	DealID      string `json:"dealId"`      // The deal ID of the activity
	Description string `json:"description"` // The description of the activity
	Details     struct {
		Actions []struct {
			ActionType     ActionType `json:"actionType"`
			AffectedDealId string     `json:"affectedDealId"`
		}
		Currency             string  `json:"currency"`
		DealReference        string  `json:"dealReference"`
		Direction            string  `json:"direction"`
		GoodTillDate         string  `json:"goodTillDate"`
		GuaranteedStop       bool    `json:"guaranteedStop"`
		Level                float64 `json:"level"`
		LimitDistance        float64 `json:"limitDistance"`
		LimitLevel           float64 `json:"limitLevel"`
		MarketName           string  `json:"marketName"`
		Size                 float64 `json:"size"`
		StopDistance         float64 `json:"stopDistance"`
		StopLevel            float64 `json:"stopLevel"`
		TrailingStep         float64 `json:"trailingStep"`
		TrailingStopDistance float64 `json:"trailingStopDistance"`
	}
	Epic     string       `json:"epic"`
	Period   string       `json:"period"` // The period of the activity item, e.g. "DFB" or "02-SEP-11". This will be the expiry time/date for sprint markets, e.g. "2015-10-13T12:42:05"
	Status   string       `json:"status"` // The status of the activity item, e.g. "ACCEPTED", "REJECTED", "DELETED"
	Type     ActivityType `json:"type"`
	Metadata struct {
		Paging struct {
			Next string `json:"next"`
			Size int    `json:"size"`
		}
	} `json:"metadata"`
}

type ActivityType string

const (
	EDIT_STOP_AND_LIMIT_ACTIVITY_TYPE ActivityType = "EDIT_STOP_AND_LIMIT"
	POSITION                          ActivityType = "POSITION"
	SYSTEM                            ActivityType = "SYSTEM"
	WORKING_ORDER                     ActivityType = "WORKING_ORDER"
)

type ActionType string

const (
	LIMIT_ORDER_AMENDED       ActionType = "LIMIT_ORDER_AMENDED"
	LIMIT_ORDER_DELETED       ActionType = "LIMIT_ORDER_DELETED"
	LIMIT_ORDER_FILLED        ActionType = "LIMIT_ORDER_FILLED"
	LIMIT_ORDER_OPENED        ActionType = "LIMIT_ORDER_OPENED"
	LIMIT_ORDER_ROLLED        ActionType = "LIMIT_ORDER_ROLLED"
	POSITION_CLOSED           ActionType = "POSITION_CLOSED"
	POSITION_DELETED          ActionType = "POSITION_DELETED"
	POSITION_OPENED           ActionType = "POSITION_OPENED"
	POSITION_PARTIALLY_CLOSED ActionType = "POSITION_PARTIALLY_CLOSED"
	POSITION_ROLLED           ActionType = "POSITION_ROLLED"
	STOP_LIMIT_AMENDED        ActionType = "STOP_LIMIT_AMENDED"
	STOP_ORDER_AMENDED        ActionType = "STOP_ORDER_AMENDED"
	STOP_ORDER_DELETED        ActionType = "STOP_ORDER_DELETED"
	STOP_ORDER_FILLED         ActionType = "STOP_ORDER_FILLED"
	STOP_ORDER_OPENED         ActionType = "STOP_ORDER_OPENED"
	STOP_ORDER_ROLLED         ActionType = "STOP_ORDER_ROLLED"
	UNKNOWN                   ActionType = "UNKNOWN"
	WORKING_ORDER_DELETED     ActionType = "WORKING_ORDER_DELETED"
)

// GetActivity - Returns the account activity history
func (ig *IGMarkets) GetActivity(ctx context.Context, from, to time.Time) (*ActivityResponse, error) {
	var parameters = []string{
		"detailed=true",
		fmt.Sprintf("from=%s", from.Format(timeFormat)),
		fmt.Sprintf("to=%s", to.Format(timeFormat)),
	}

	bodyReq := new(bytes.Buffer)

	url := fmt.Sprintf("%s/gateway/deal/history/activity?%s", ig.APIURL, strings.Join(parameters, "&"))
	req, err := http.NewRequest("GET", url, bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	endpointVersion := 3
	igResponseInterface, err := ig.doRequest(ctx, req, endpointVersion, ActivityResponse{})
	if err != nil {
		return nil, err
	}

	igResponse, _ := igResponseInterface.(*ActivityResponse)
	return igResponse, nil
}
