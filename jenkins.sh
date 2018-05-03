#!/bin/bash

set -ex

export GOPATH=/tmp/gopath-jenkins-loom-sdk-pipeline-test-${loom_sdk_pipeline_build_number}
ln -sfn `pwd` $GOPATH/src/github.com/loomnetwork/etherboy-core
cd $GOPATH/src/github.com/loomnetwork

# Building Go Loom

echo "Building Go Loom"

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
make test
cd ..

echo "Building Etherboy"

## Building etherboy-core
cd etherboy-core
go get github.com/pkg/errors
make clean
make deps
make


echo ${BUILD_NUMBER}
echo ${JOB_NAME}

gsutil cp run/contracts/etherboycore.so gs://private.delegatecall.com/etherboy/linux/build-$BUILD_NUMBER/etherboycore.so
gsutil cp run/cmds/etherboycli gs://private.delegatecall.com/etherboy/linux/build-$BUILD_NUMBER/etherboycli
