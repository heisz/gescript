# 
#  Makefile to simplify go command sequences.
# 
#  Copyright (C) 2005-2026 J.M. Heisz.  All Rights Reserved.
#  See the LICENSE file accompanying the distribution your rights to use
#  this software.
# 

export PATH := /toolkit/go/bin:$(PATH)

all:
	go mod tidy
	go generate ./...
	go build

fmt:
	go fmt gescript.go
	go fmt internal/parser/*.go
	go fmt internal/engine/*.go

test:
	go mod tidy
	go generate ./...
	go test -v ./... -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html
