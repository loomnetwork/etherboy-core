## Etherboy Smart contracts

This repo contains the smart contracts for the Etherboy Game.

For the Unity frontend see this [game repo](https://github.com/loomnetwork/Etherboy)

![Animation](https://loomx.io/developers/img/etherboy-clip.gif)


## Docs

[Docs Site](https://loomx.io/developers/docs/en/etherboy-game.html)

## Building

```shell
cd $GOPATH/src/github.com/loomnetwork
# clone into gopath
git clone git@github.com:loomnetwork/etherboy-core.git
# switch to the loom repo
cd $GOPATH/src/github.com/loomnetwork/etherboy-core
# build the contract plugin, cmd plugin and indexer
make
```

## Running the node

```
# start the node
cd run
export LOOM_EXE="path/to/loom_executable'
$LOOM_EXE init
# modify genesis.json similar to below
./loom run 2>&1 | tee -a etherboy.log
```

## Creating an account and running transactions
```
export ETHERBOY_CLI="path/to/etherbodycli"

# create a key pair
LOOM_CMDPLUGINDIR=cmds/ $ETHERBOY_CLI genkey -k priv

# send a create account tx
LOOM_CMDPLUGINDIR=cmds/ $ETHERBOY_CLI create-acct -k priv -u loom

# send a set stage tx
LOOM_CMDPLUGINDIR=cmds/ $ETHERBOY_CLI set -v 1010 -k priv -u loom
```

## Regenerating Protobufs

```shell
make proto
```
