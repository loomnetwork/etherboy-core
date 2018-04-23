## Building

Ensure `github.com/loomnetwork/loom` is in your `GOPATH`, then:

```shell
cd $GOPATH/src/github.com/loomnetwork
# clone into gopath
git clone git@github.com:loomnetwork/etherboy-core.git

go build -buildmode=plugin -o out/etherboycore.so etherboy.go

cp out/etherboycore.so ../loom/contracts

cd ../loom

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
