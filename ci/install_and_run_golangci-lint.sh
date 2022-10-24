#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

version=v1.50.1
curl -sfSL -o install.sh "https://raw.githubusercontent.com/golangci/golangci-lint/$version/install.sh"
echo "2078ae446b6fc0d50e847685102982af3cdffdcf7152079ef102bc552cc5c46645bef51d3aca587fb9864e91c1334f1aa01a540b10daea95c779455ebdeb0db4 install.sh" | sha512sum -c
sh ./install.sh -b "$(go env GOPATH)/bin" "$version"
~/go/bin/golangci-lint run
