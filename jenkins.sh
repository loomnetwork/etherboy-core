#!/bin/bash

set -ex

export GOPATH=~/gopath
ln -sfn `pwd` $GOPATH/src/github.com/loomnetwork/etherboy-core
cd $GOPATH/src/github.com/loomnetwork

if cd go-loom; then
  ## Always build with latest go-loom
  git pull origin master
else
  git clone git@github.com:loomnetwork/go-loom.git $GOPATH/src/github.com/loomnetwork/go-loom 
  cd go-loom
fi

## Building go-loom
make clean
make deps
make 
cd ..

## Building etherboy-core
cd etherboy-core
make clean
make deps
make 
