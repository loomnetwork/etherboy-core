#!/bin/bash

set -ex

echo "_______________________IN PROCESS PLUGIN__________________________________"

echo ${BUILDNUM}
echo ${JOBNAME}



echo "Using Loom Build ${loom_build}"

echo "Using Etherboy Build ${etherboy_build}"

echo "Cleaning up tmp files"
rm -rf /tmp/loom
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

rm -f genesis.json

echo "{
    \"contracts\": [
        {
            \"vm\": \"plugin\",
            \"name\": \"etherboycore\",
            \"format\": \"plugin\",
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

./etherboycli set -k key

./etherboycli get -k key

pkill -f loom-linux

cat loom_run_${etherboy_build}_${loom_build}.log
rm *.log

echo "_______________________OUT PROCESS PLUGIN__________________________________"


mkdir -p external_test
cd external_test
gsutil cp gs://private.delegatecall.com/loom/linux/build-${loom_build}/loom loom
chmod +x loom

./loom init --force
mkdir -p contracts
rm -f genesis.json

echo "{
    \"contracts\": [
        {
            \"vm\": \"plugin\",
            \"name\": \"etherboycore\",
            \"format\": \"plugin\",
            \"location\": \"etherboycore:0.0.1\",
            \"init\": {

            }
        }
    ]
}
" >> genesis.json


gsutil cp gs://private.delegatecall.com/etherboy/linux/build-${etherboy_build}/etherboycore.0.0.1 contracts/etherboycore.0.0.1
gsutil cp gs://private.delegatecall.com/etherboy/linux/build-${etherboy_build}/etherboycli etherboycli
chmod +x etherboycli
chmod +x contracts/etherboycore.0.0.1

./loom run > loom_run_${etherboy_build}_${loom_build}.log 2>&1 &

sleep 5
./etherboycli genkey -k key
./etherboycli create-acct -k key
./etherboycli set -k key
./etherboycli get -k key

pkill -f loom
pkill -f etherboycore
cat loom_run_${etherboy_build}_${loom_build}.log

rm *.log
