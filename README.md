# igmarkets - Unofficial IG Markets Trading API for Golang

This is an **unofficial** API for [IG Markets Trading REST API](https://labs.ig.com/rest-trading-api-reference). The StreamingAPI is not part of this project.

**Disclaimer**: This library is not associated with IG Markets Limited or any of its affiliates or subsidiaries. If you use this library, you should contact them to make sure they are okay with how you intend to use it. Use this lib at your own risk.

Reference: https://labs.ig.com/rest-trading-api-reference

## Currently supported endpoints

### Session

- POST /session

### Markets

- GET /markets/{epic}

### Positions

- POST /positions/otc
- PUT /positions/otc/{dealId}
- GET /positions
- DELETE /positions
- GET /confirms/{dealReference}

### Workingorders
- GET /workingorders
- POST /workingorders/otc
- DELETE /workingorders/otc/{dealId}

### Prices

- GET /prices/{epic}/{resolution}/{startDate}/{endDate}

### Watchlists
- GET /watchlists/{watchlistid}

### History

- GET /history/activity
- GET /history/transactions

## Example

```go
package main

import (
        "fmt"
        "github.com/sklinkert/igmarkets"
        "time"
)

var ig *igmarkets.IGMarkets

func main() {
        httpTimeout := time.Duration(5 * time.Second)

        ig = igmarkets.New(igmarkets.DemoAPIURL, "APIKEY", "ACCOUNTID", "USERNAME/IDENTIFIER", "PASSWORD", httpTimeout)
        if err := ig.Login(); err != nil {
                fmt.Println("Unable to login into IG account", err)
        }

        // Get current open ask, open bid, close ask, close bid, high ask, high bid, low ask, and low bid
        prices, _ := ig.GetPrice("CS.D.EURUSD.CFD.IP")

        fmt.Println(prices)

        // Place a new order
        order := igmarkets.OTCOrderRequest{
                Epic:           "CS.D.EURUSD.CFD.IP",
                OrderType:      "MARKET",
                CurrencyCode:   "USD",
                Direction:      "BUY",
                Size:           1.0,
                Expiry:         "-",
                StopDistance:   "10", // Pips
                LimitDistance:  "5",  // Pips
                GuaranteedStop: true,
                ForceOpen:      true,
        }
        dealRef, err := ig.PlaceOTCOrder(order)
        if err != nil {
                fmt.Println("Unable to place order:", err)
                return
        }
        fmt.Println("New order placed with dealRef", dealRef)

        // Check order status
        confirmation, err := ig.GetDealConfirmation(dealRef)
        if err != nil {
                fmt.Println("Cannot get deal confirmation for:", dealRef, err)
                return
        }

        fmt.Println("Order dealRef", dealRef)
        fmt.Println("DealStatus", confirmation.DealStatus) // "ACCEPTED"
        fmt.Println("Profit", confirmation.Profit, confirmation.ProfitCurrency)
        fmt.Println("Status", confirmation.Status) // "OPEN"
        fmt.Println("Reason", confirmation.Reason)
        fmt.Println("Level", confirmation.Level) // Buy price
}
```

## TODOs

- Write basic tests

Feel free to send PRs.
