#!/bin/bash
version=v1.45.2
curl -sfSL -o install.sh "https://raw.githubusercontent.com/golangci/golangci-lint/$version/install.sh"
echo "91b7e19e60c36194ea02cedb90af5f4d93b5725dc0bfa4505bdfabbc79e809ffd2f04f783e309bde0c7c383d5a1d5719244c3e4306f6cff19c3c67748388b0db install.sh" | sha512sum -c
sh ./install.sh -b "$(go env GOPATH)/bin" "$version"
~/go/bin/golangci-lint run
