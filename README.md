# word-of-wisdom-go

## TODOs

* Timeouts
* Better justification of the hashing algo
* Command token to constants
* Increase complexity based on request rate
* Solve in parallel

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

### High level design of the system is outlined below
<img src="./doc/wow-high-level.svg">

Each time the client will be requesting a next WoW, the server will generate the challenge, send it back to the client. Once response is received - the server will verify the challenge and in case of success - return a valid WoW.

Every new request will be recorded by request rate monitor (using client IP address). The challenge generator will use the request rate monitoring data to define a complexity for the challenge to solve. The complexity will increase exponentially.

We want to be nice to our clients, so good actors will not be required solving challenge.

### Base Challenges Algo

The algorithm should allow varying complexity and should be friendly to end users (e.g should be possible to run on a end user hardware that can be weak). It should also be relatively cheap to verify server side.

The hash based (Hashcash) SHA-256 algorithm is a good candidate for this. The complexity of solving the challenge will be controlled by varying leading zeros in the hash output. The number of leading zeros will be varying and will increase based on the request rate of the target IP address.

## Project structure

* [cmd/server](./cmd/server) is a main entrypoint to start API server. Dependencies wire-up is happening here.
* [pkg/api/tcp](./pkg/api/tcp) - includes components required to process raw TCP requests
* `pkg/app` - is assumed to include application layer code (e.g business logic). 
* `pkg/services` - lower level components are supposed to be here (e.g database access layer e.t.c).

## Project Setup

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
go test -v ./service/pkg/api/http/v1controllers/ --run TestHealthCheckController

# Run same test multiple times
# This is useful for tests that are flaky
go test -v -count=5 ./service/pkg/api/http/v1controllers/ --run TestHealthCheckController

# Run and watch
gow test -v ./service/pkg/api/http/v1controllers/ --run TestHealthCheckController
```
### Run local server:

```bash
# Regular mode
go run ./cmd/service/

# Watch mode (double ^C to stop)
gow run ./cmd/service/
```