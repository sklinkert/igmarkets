# igmarkets
Unofficial IG Markets API for Golang

This is an **unofficial** API for [IG Markets Trading API](https://labs.ig.com/rest-trading-api-reference). The StreamingAPI is not part of this project.

**Disclaimer**: This library is not associated with IG Markets Limited. If you use this library, you should contact them to make sure they are okay with how you intend to use it.

Reference: https://labs.ig.com/rest-trading-api-reference

## Currently supported endpoints

### Session

- POST /session

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


### History

- GET /history/activity
- GET /history/transactions

## Example

...tbc...

## TODOs

- Write basic tests
- Implement /session/refresh-token
