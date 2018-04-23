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

Install the [`protoc` compiler](https://github.com/google/protobuf/releases), then:

```shell
go get github.com/gogo/protobuf
cd $GOPATH/src/github.com/loomnetwork/etherboy-core
go build github.com/gogo/protobuf/protoc-gen-gogo
```

Once everything is installed run (note `Mgoogle` is not a typo!):

```shell
cd $GOPATH/src/github.com/loomnetwork/etherboy-core
protoc --plugin=./protoc-gen-gogo \
--gogo_out=Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types:./txmsg \
-I$GOPATH/src -I$GOPATH/src/github.com/gogo/protobuf/protobuf -I. ./txmsg.proto
```
