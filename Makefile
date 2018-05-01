PKG = github.com/loomnetwork/etherboy-core
PROTOC = protoc --plugin=./protoc-gen-gogo -Ivendor -I$(GOPATH)/src -I/usr/local/include

.PHONY: all clean test lint deps proto

all: internal-plugin etherboy-cli etherboy-indexer

internal-plugin: etherboycore.so

etherboycore.so: proto
	mkdir -p run/contracts
	go build -buildmode=plugin -o run/contracts/$@ ./etherboy.go

etherboy-cli: proto
	mkdir -p run/cmds
	go build -o run/cmds/etherboycli tools/cli/etherboycli/etherboycli.go

etherboy-indexer:
	go build ./tools/cli/indexer

protoc-gen-gogo:
	go build github.com/gogo/protobuf/protoc-gen-gogo

%.pb.go: %.proto protoc-gen-gogo
	$(PROTOC) --gogo_out=$(GOPATH)/src \
$(PKG)/$<

proto: txmsg/txmsg.pb.go testdata/test.pb.go

test: proto
	go test $(PKG)/...

lint:
	golint ./...

deps:
	go get \
		github.com/gogo/protobuf/jsonpb \
		github.com/gogo/protobuf/proto \
		github.com/spf13/cobra \
		github.com/gomodule/redigo/redis

clean:
	go clean
	rm -f \
		protoc-gen-gogo \
		txmsg/txmsg.pb.go \
		testdata/test.pb.go \
		run/contracts/etherboy.so \
		run/cmds/etherboyclu.so
