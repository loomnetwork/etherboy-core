#!/bin/bash

set -ex

glide install
mkdir -p run/contracts
rm -f run/contracts/etherboycore.so
go build -buildmode=plugin -o run/contracts/etherboycore.so etherboy.go
