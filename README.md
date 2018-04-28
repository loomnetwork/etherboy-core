## Building

Ensure `github.com/loomnetwork/loom` is in your `GOPATH`, then:

```shell
cd $GOPATH/src/github.com/loomnetwork
# clone into gopath
git clone git@github.com:loomnetwork/etherboy-core.git
# switch to the loom repo
cd $GOPATH/src/github.com/loomnetwork/loom
# build the contract plugin
go build -buildmode=plugin -o ./contracts/etherboycore.so ../etherboy-core/etherboy.go
# start the node
./loom run
```

## Regenerating Protobufs

```shell
make proto
```
