preflight : 
	go test --coverprofile cover.out

build:
	go generate

test: 
	sh tests.sh

coverage:
	go tool cover -html=cover.out

all : preflight build test
cover : preflight build test coverage
