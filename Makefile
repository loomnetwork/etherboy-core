PKG = github.com/loomnetwork/etherboy-core
PROTOC = protoc --plugin=./protoc-gen-gogo -Ivendor -I$(GOPATH)/src -I/usr/local/include

.PHONY: all clean test lint deps proto

all: internal-plugin external-plugin etherboy-cli

external-plugin: etherboycore

internal-plugin: etherboycore.so

etherboycore: proto
	go build -o contracts/$@ $(PKG)/etherboy.go

etherboycore.so: proto
	go build -buildmode=plugin -o contracts/$@ ./etherboy.go

etherboy-cli: proto
	go build ./tools/cli

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
		github.com/spf13/cobra

clean:
	go clean
	rm -f \
		protoc-gen-gogo \
		txmsg/txmsg.pb.go \
		testdata/test.pb.go \
		contracts/etherboy \
		contracts/etherboy.so \
