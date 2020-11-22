package igmarkets

import (
	"bytes"
	"fmt"
	"net/http"
)

// GetOTCWorkingOrders - Get all working orders
func (ig *IGMarkets) Bla() (*WorkingOrders, error) {
	bodyReq := new(bytes.Buffer)
	req, err := http.NewRequest("GET", ig.APIURL+"/gateway/deal/workingorders/", bodyReq)
	if err != nil {
		return nil, fmt.Errorf("igmarkets: unable to create HTTP request: %v", err)
	}

	igResponseInterface, err := ig.doRequest(req, 2, WorkingOrders{})
	if err != nil {
		return nil, err
	}
	igResponse, _ := igResponseInterface.(*WorkingOrders)

	return igResponse, err
}
