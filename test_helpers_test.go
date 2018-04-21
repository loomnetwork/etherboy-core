package main

import (
	"time"

	"github.com/loomnetwork/loom"
	"github.com/loomnetwork/loom/plugin"
	"github.com/loomnetwork/loom/vm"
)

type FakeCTX struct {
}
type FakeContext struct {
	caller  loom.Address
	address loom.Address
	loom.State
	vm.VM
	data map[string][]byte
}

func CreateFakeContext() plugin.Context {
	return &FakeContext{data: make(map[string][]byte)}
}

func (c FakeContext) Call(addr loom.Address, input []byte) ([]byte, error) {
	return c.VM.Call(c.address, addr, input)
}

func (c FakeContext) StaticCall(addr loom.Address, input []byte) ([]byte, error) {
	return c.VM.StaticCall(c.address, addr, input)
}

func (c FakeContext) Message() plugin.Message {
	return plugin.Message{
		//		Sender: c.caller,
	}
}

func (c FakeContext) ContractAddress() loom.Address {
	return c.address
}

func (c FakeContext) Now() time.Time {
	return time.Unix(0, 0)
}

func (c FakeContext) Emit(event []byte) {
}
func (c FakeContext) Has(key []byte) bool {
	_, ok := c.data[string(key)]
	return ok
}
func (c FakeContext) Set(key []byte, value []byte) {
	c.data[string(key)] = value
}
