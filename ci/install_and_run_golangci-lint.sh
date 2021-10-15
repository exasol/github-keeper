#!/bin/bash
curl -sfSL -o install.sh https://raw.githubusercontent.com/golangci/golangci-lint/v1.42.1/install.sh
echo "0762a15cd7ac4ef439cdfb63bf12865b53c03543c0981b5d46fb542497238b829a33334a288c09d356ec8f797804dfed56f687cbcffcf853ff93da6da4106736 install.sh" | sha512sum -c
sh ./install.sh -b "$(go env GOPATH)/bin" v1.42.1
~/go/bin/golangci-lint run