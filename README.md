# word-of-wisdom-go

[![Test](https://github.com/gemyago/word-of-wisdom-go/actions/workflows/run-tests.yml/badge.svg)](https://github.com/gemyago/word-of-wisdom-go/actions/workflows/run-tests.yml)
[![Coverage](https://raw.githubusercontent.com/gemyago/word-of-wisdom-go/test-artifacts/coverage/golang-coverage.svg)](https://htmlpreview.github.io/?https://raw.githubusercontent.com/gemyago/word-of-wisdom-go/test-artifacts/coverage/golang-coverage.html)

## Running Locally Via Docker

It is assumed that GNU make and docker are available

```sh
# Build local images
make docker-images

# Start tcp server
docker run -p 44221:44221 --rm -it localhost:6000/word-of-wisdom-server:latest server tcp

# Run client
docker run --net=host --rm -it localhost:6000/word-of-wisdom-client:latest client get-wow
```

Both client & server support additional flags, run with -h to see the details.

## TODOs

* Increase complexity based on request rate
* Solve in parallel
* Better justification of the hashing algo
* Command token to constants

## Task Definition

Design and implement “Word of Wisdom” tcp server.
 • TCP server should be protected from DDOS attacks with the Proof of Work (https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.
 • The choice of the POW algorithm should be explained.
 • After Proof Of Work verification, server should send one of the quotes from “word of wisdom” book or any other collection of the quotes.
 • Docker file should be provided both for the server and for the client that solves the POW challenge

## System Design

Abbreviations:
* WoW - Word of Wisdom
* PoW - Proof of Work

### Functional Requirements
* Client should be able to request WoW phrase

### Non Functional
* The system should have a protection mechanism from DDOS attacks using PoW with the challenge-response protocol
* The system should be friendly to good actors (clients)
  * Good actors should get no challenge at all or simple challenge to solve
* Bad actors should start getting increasingly more complex challenge to solve

Nice to have, but out of scope:
- Monitoring mechanism to see a rate of bad challenge responses
- Monitoring mechanism to see if we are under attack (e.g DDOS started)
- CI/CD, Deployment pipeline e.t.c

### High level conceptual design of the system is outlined below
<img src="./doc/wow-high-level.svg">

Each time the client will be requesting a next WoW, the server may generate the challenge, send it back to the client. Once the response is received - the server will verify the challenge and in case of success - return a valid WoW.

Every new request will be recorded by request rate monitor (using client IP address). The challenge generator will use the request rate monitoring data to define a complexity for the challenge to solve. The complexity will increase exponentially.

We want to be nice to our clients, so good actors will not be required solving challenge.

### Base Challenges Algo

The algorithm should allow varying complexity and should be friendly to end users (e.g should be possible to run on a end user hardware that can be weak). It should also be relatively cheap to verify server side.

The hash based (Hashcash) SHA-256 algorithm is a good candidate for this. The complexity of solving the challenge will be controlled by varying leading zeros in the hash output. The number of leading zeros will be varying and will increase based on the request rate of the target IP address and will also depend on overall server load.

## Project structure

### Packages overview:
* [cmd/server](./cmd/server) Server cli
* [cmd/client](./cmd/client) Client cli
* [pkg/api/tcp](./pkg/api/tcp) - TCP server and protocol processing
* [pkg/app](./pkg/app) - Application business logic related components
* [pkg/services](./pkg/services) - Lower level services and other supporting components

### Modules used

* [dig](github.com/uber-go/dig) - DI toolkit
* [cobra](github.com/spf13/cobra) - CLI interactions
* [viper](github.com/spf13/viper) - Configuration management
* [testify](github.com/stretchr/testify) - Assertion toolkit
* [mockery](github.com/vektra/mockery) - Mocks generator
* [golangci-lint](https://golangci-lint.run/) - Linter
* slog for logging

## Dev env setup

Please have the following tools installed: 
* [direnv](https://github.com/direnv/direnv) 
* [gobrew](https://github.com/kevincobain2000/gobrew#install-or-update)

Install/Update dependencies: 
```sh
# Install
go mod download
make tools

# Update:
go get -u ./... && go mod tidy
```

### Lint and Tests

Run all lint and tests:
```bash
make lint
make test
```

Run specific tests:
```bash
# Run once
go test -v ./pkg/app/challenges/ --run TestChallenges

# Run same test multiple times
# This is useful to catch flaky tests
go test -v -count=5 ./pkg/app/challenges/ --run TestChallenges

# Run and watch
gow test -v ./pkg/app/challenges/ --run TestChallenges
```
### Run local server:

```bash
# Start TCP server
# use gow to run it in watch mode
go run ./cmd/server/ tcp

# Run client
go run ./cmd/client/ get-wow
```

### Tips on testing

To test challenge solution behavior and performance use `solve-challenge` command of the client:
```sh
go run ./cmd/client/ solve-challenge --challenge some-challenge-string -c 1
```

To test the TCP server directly use nc to send raw commands:
```sh
# Initiate nc client session
nc -v localhost 44221
```

Testing flow:
* Start the TCP server
* Initiate nc client session
* Type `GET_WOW`, should see `WOW: <phrase>` immediately if withing the global limit or per client limit.
* Retry `GET_WOW`, should see `CHALLENGE_REQUIRED` with the challenge and the complexity.
  * If responding with `CHALLENGE_RESULT: invalid-solution`, should see `ERR: CHALLENGE_VERIFICATION_FAILED`
  * If responding with valid solution, should get the `WOW: <phrase>`
  * Use `solve-challenge` command above to solve the challenge and use the solution
  * Note: max session duration is 10s (see default.json). You may want to increase it for local testing purposes.