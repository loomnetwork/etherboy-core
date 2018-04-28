package main

import (
	"time"

	loom "github.com/loomnetwork/go-loom"
	plugin "github.com/loomnetwork/go-loom/plugin"
	"github.com/loomnetwork/go-loom/types"
)

type FakeContext struct {
	caller  loom.Address
	address loom.Address
	data    map[string][]byte
}

func CreateFakeContext() plugin.Context {
	return &FakeContext{data: make(map[string][]byte)}
}

func (c FakeContext) Call(addr loom.Address, input []byte) ([]byte, error) {
	return nil, nil
}

func (c FakeContext) StaticCall(addr loom.Address, input []byte) ([]byte, error) {
	return nil, nil
}

func (c FakeContext) Message() plugin.Message {
	return plugin.Message{
		Sender: loom.Address{},
	}
}

func (c FakeContext) Block() types.BlockHeader {
	return types.BlockHeader{}
}

func (c FakeContext) ContractAddress() loom.Address {
	return c.address
}

func (c FakeContext) Now() time.Time {
	return time.Unix(0, 0)
}

func (c FakeContext) Emit(event []byte) {
}

func (c FakeContext) Get(key []byte) []byte {
	v, _ := c.data[string(key)]
	return v
}

func (c FakeContext) Has(key []byte) bool {
	_, ok := c.data[string(key)]
	return ok
}

func (c FakeContext) Set(key []byte, value []byte) {
	c.data[string(key)] = value
}

func (c FakeContext) Delete(key []byte) {
}
