package main

import (
	"fmt"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/loomnetwork/etherboy-core/txmsg"
	"github.com/loomnetwork/loom/plugin"
)

func TestCreateAccount(t *testing.T) {
	e := NewEtherBoyContract()
	tx := &txmsg.EtherboyCreateAccountTx{
		Version: 0,
		Owner:   "aditya",
		Data:    []byte("dummy"),
	}

	any, err := types.MarshalAny(tx)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	msg := &txmsg.SimpleContractMethod{
		Version: 0,
		Method:  fmt.Sprintf("%s.CreateAccount", e.Meta().Name),
		Data:    any,
	}

	msgBytes, err := proto.Marshal(msg)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	ctx := CreateFakeContext()
	req := &plugin.Request{}

	fmt.Printf("Data: msgBytes-%v\n", msgBytes)

	req.Body = msgBytes
	resp, err := e.Call(ctx, req)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	fmt.Printf("resp -%v\n", resp)
}
