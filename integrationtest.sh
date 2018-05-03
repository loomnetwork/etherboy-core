#!/bin/bash

set -ex

echo "_______________________IN PROCESS PLUGIN__________________________________"

echo "Using Loom Build ${loom_build}"

echo "Using Etherboy Build ${etherboy_build}"

mkdir -p /tmp/loom/contracts

echo "Downloading loom sdk"

gsutil cp gs://private.delegatecall.com/loom/linux/build-${loom_build}/loom /tmp/loom/loom-linux

echo "Downloading etherboy plugin"

echo "Downloading etherboy plugin"

gsutil cp gs://private.delegatecall.com/etherboy/linux/build-${etherboy_build}/etherboycore.so /tmp/loom/contracts/etherboycore.so

echo "Downloading etherboy cli"

gsutil cp gs://private.delegatecall.com/etherboy/linux/build-${etherboy_build}/etherboycli /tmp/loom/etherboycli

cd /tmp/loom

chmod +x loom-linux
chmod +x etherboycli

./loom-linux init --force

rm genesis.json

echo "{
    \"contracts\": [
        {
            \"vm\": \"plugin\",
            \"format\": \"plugin\",
            \"name\": \"etherboycore\",
            \"location\": \"etherboycore:0.0.1\",
            \"init\": {

            }
        }
    ]
}
" >> genesis.json


./loom-linux run > loom_run_${etherboy_build}_${loom_build}.log 2>&1 &

sleep 5

./etherboycli genkey -k key

./etherboycli create-acct -k key

pkill -f loom-linux

cat loom_run_${etherboy_build}_${loom_build}.log


echo "_______________________OUT PROCESS PLUGIN__________________________________"
