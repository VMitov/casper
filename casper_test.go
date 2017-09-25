package casper

import "flag"

// It is defined in each package so you can run `go test ./...`
var full = flag.Bool("full", false, "Run all tests including integration")
