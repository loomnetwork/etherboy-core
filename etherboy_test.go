package main

import (
	"fmt"
	"testing"
	"time"

	proto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/loomnetwork/etherboy-core/txmsg"
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
func TestCreateAccount(t *testing.T) {
	tx := &txmsg.EtherboyCreateAccountTx{
		Data: []byte("dummy"),
	}

	any, err := ptypes.MarshalAny(tx)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	msg := &txmsg.SimpleContractMethod{
		Version: 0,
		Owner:   "aditya",
		Method:  "simple.CreateAccount",
		Data:    any,
	}

	msgBytes, err := proto.Marshal(msg)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	e := &EtherBoy{}

	ctx := CreateFakeContext()
	req := &plugin.Request{}

	fmt.Printf("Data: msgBytes-%v\n", msgBytes)
	err = e.Init(ctx, req)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	req.Body = msgBytes
	resp, err := e.Call(ctx, req)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	fmt.Printf("resp -%v\n", resp)
}
