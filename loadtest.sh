#!/bin/bash

set -ex

echo "_______________________IN PROCESS PLUGIN__________________________________"

echo ${BUILDNUM}
echo ${JOBNAME}

LOADTMP=/tmp/loomloadtest
BASEURL="https://storage.googleapis.com/private.delegatecall.com"

echo "Using Loom Build ${loom_build}"

echo "Using Etherboy Build ${etherboy_build}"

echo "Cleaning up tmp files"

rm -rf ${LOADTMP}

mkdir -p ${LOADTMP}/internal/contracts

echo "Downloading loom sdk"

wget -q ${BASEURL}/loom/linux/build-${loom_build}/loom -O ${LOADTMP}/internal/loom-linux

echo "Downloading etherboy plugin"

wget -q ${BASEURL}/etherboy/linux/build-${etherboy_build}/etherboycore.so -O ${LOADTMP}/internal/contracts/etherboycore.so

echo "Downloading etherboy cli"

wget -q ${BASEURL}/etherboy/linux/build-${etherboy_build}/etherboycli -O ${LOADTMP}/internal/etherboycli

cd ${LOADTMP}/internal

chmod +x loom-linux
chmod +x etherboycli

./loom-linux init --force

rm -f genesis.json

cat << EOF > genesis.json
{
    "contracts": [
        {
            "vm": "plugin",
            "name": "etherboycore",
            "format": "plugin",
            "location": "etherboycore:0.0.1",
            "init": {

            }
        }
    ]
}
EOF

./loom-linux run > loom_run_${etherboy_build}_${loom_build}.log 2>&1 &

sleep 5

./etherboycli genkey -k key

./etherboycli loadtest-create -k key -i 100 -m 100 -r "http://127.0.0.1:9999"

./etherboycli loadtest-set -k key -i 100 -m 100 -r "http://127.0.0.1:9999"

./etherboycli loadtest-get -k key -i 10000 -m 100 -c 2 -r "http://127.0.0.1:9999"

pkill -f loom-linux

cat loom_run_${etherboy_build}_${loom_build}.log

rm *.log

echo "_______________________OUT PROCESS PLUGIN__________________________________"


mkdir -p ${LOADTMP}/external/contracts

cd ${LOADTMP}/external

wget -q ${BASEURL}/loom/linux/build-${loom_build}/loom -O ${LOADTMP}/external/loom-linux

chmod +x loom-linux

./loom-linux init --force

rm -f genesis.json

cat << EOF > genesis.json
{
    "contracts": [
        {
            "vm": "plugin",
            "name": "etherboycore",
            "format": "plugin",
            "location": "etherboycore:0.0.1",
            "init": {

            }
        }
    ]
}
EOF

wget -q ${BASEURL}/etherboy/linux/build-${etherboy_build}/etherboycore.0.0.1 -O ${LOADTMP}/external/contracts/etherboycore.0.0.1

wget -q ${BASEURL}/etherboy/linux/build-${etherboy_build}/etherboycli -O ${LOADTMP}/external/etherboycli
chmod +x etherboycli
chmod +x contracts/etherboycore.0.0.1

./loom-linux run > loom_run_${etherboy_build}_${loom_build}.log 2>&1 &

sleep 5

./etherboycli genkey -k key

./etherboycli loadtest-create -k key -i 100 -m 100 -r "http://127.0.0.1:9999"

./etherboycli loadtest-set -k key -i 100 -m 100 -r "http://127.0.0.1:9999"

./etherboycli loadtest-get -k key -i 10000 -m 100 -c 2 -r "http://127.0.0.1:9999"

pkill -f loom-linux
pkill -f etherboycore

cat loom_run_${etherboy_build}_${loom_build}.log

rm *.log