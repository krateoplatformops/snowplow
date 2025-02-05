#!/bin/bash


go test -tags=unit -count=1 -cover -coverprofile=coverage.out ./...
grep -v "apis|docs|internal/resolvers/props|internal/resolvers/definitions" coverage.out > coverage_filtered.out
go tool cover -func=coverage_filtered.out

