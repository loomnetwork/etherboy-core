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
