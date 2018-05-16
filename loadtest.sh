#!/bin/bash

set -ex

echo "_______________________IN PROCESS PLUGIN__________________________________"

echo ${BUILDNUM}
echo ${JOBNAME}

LOADTMP=/tmp/loomloadtest



echo "Using Loom Build ${loom_build}"

echo "Using Etherboy Build ${etherboy_build}"

echo "Cleaning up tmp files"
rm -rf ${LOADTMP}
mkdir -p ${LOADTMP}/contracts

echo "Downloading loom sdk"

gsutil cp gs://private.delegatecall.com/loom/linux/build-${loom_build}/loom ${LOADTMP}/loom-linux

echo "Downloading etherboy plugin"

echo "Downloading etherboy plugin"

gsutil cp gs://private.delegatecall.com/etherboy/linux/build-${etherboy_build}/etherboycore.so ${LOADTMP}/contracts/etherboycore.so

echo "Downloading etherboy cli"

gsutil cp gs://private.delegatecall.com/etherboy/linux/build-${etherboy_build}/etherboycli ${LOADTMP}/etherboycli

cd ${LOADTMP}

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

./etherboycli loadtest-create -k key -i 100 -m 100

./etherboycli loadtest-set -k key -i 100 -m 100

./etherboycli loadtest-get -k key -i 10000 -m 100 -c 2

pkill -f loom-linux

cat loom_run_${etherboy_build}_${loom_build}.log

rm *.log