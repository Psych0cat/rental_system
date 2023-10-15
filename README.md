

## Quick Start
`docker compose up`

### API 
##### `GET  /api/v1/auto/type/:type` - get available auto by type. `standard` and `special` by default 
##### `POST /api/v1/auto/bind` - rent an auto. Body example: `{"auto_id": "MINI-COOPER-SE", "days": 9}`
##### `GET  /api/v1/auto/release/:autoId` - return an auto, get checkuot in response
##### `GET  /api/v1/auto/commission/:auto_id` - get current commission and insurance for the auto

### EXAMPLE

The request on Special auto was made on 15.10.23 and canceled in the same day.
`curl --location 'http://appaddress:8080/api/v1/auto/bind' \
--header 'Content-Type: application/json' \
--data '{"auto_id": "John-Deere-1050K", "days": 10}'`

returns 200 ok

`curl --location --request GET 'http://localhost:8080/api/v1/auto/commission/John-Deere-1050K'`

returns 200 {
`"commission": 440,"insurance": 0 }`

15.10.23 is Sunday, we count the full last day of the contract + agreement commission
No insurance for the type

1 * 200 + 200 * 0,2 + 200 = 440

`curl --location 'http://localhost:8080/api/v1/auto/release/John-Deere-1050K'`

returns 200 `{"checkout": 2320 }`

Minimum days for this type is 10, they will be paid in full. 
We have 7 business days + 3 weekends with commission 0.20 + agreement cost.

10 * 200 + 3 * 20 / 100 + 200 = 2320

Please see more examples of calculations in the service testing file.

### Testing
`go test ./...`
All the business logic is covered in the Service repository
Test are supposed to run on a live DB, they will create their own auto types and clean after, not to interfere with manual tests for example.
### Customization
Commissions and thresholds can be set via corresponding DB tables. By default, all is set up according to the requirements.
### Limitations
Commissions calculation is flexible enough, though some corner cases, not mentioned in the technical task, might not be implemented. For example, weekday commission  + penalty commission without weekend commission.`
