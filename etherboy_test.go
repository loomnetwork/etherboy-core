package main

import (
	"encoding/json"
	"fmt"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	contract "github.com/loomnetwork/etherboy-core/contract-helpers"
	"github.com/loomnetwork/etherboy-core/txmsg"
	"github.com/loomnetwork/loom/plugin"
	"github.com/pkg/errors"
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

type FakeQueryParams struct {
	Key   string
	Value int
}
type FakeQueryResult struct {
	Value string
}

type fakeStaticCallContract struct {
	contract.RequestDispatcher
}

func (e *fakeStaticCallContract) Meta() plugin.Meta {
	return plugin.Meta{
		Name:    "fakeContract",
		Version: "0.0.1",
	}
}

func (c *fakeStaticCallContract) Init(ctx plugin.Context, req *plugin.Request) error {
	return nil
}

func (c *fakeStaticCallContract) QueryMethod(ctx plugin.Context, params *FakeQueryParams) (*FakeQueryResult, error) {
	return &FakeQueryResult{
		Value: "alice",
	}, nil
}

func (c *fakeStaticCallContract) FailingQueryMethod(ctx plugin.Context, params *FakeQueryParams) (*FakeQueryResult, error) {
	return nil, errors.New("query failed")
}

func newFakeStaticCallContract() plugin.Contract {
	c := &fakeStaticCallContract{}
	c.RequestDispatcher.Init(c)
	return c
}

func TestStaticCallDispatch_JSON(t *testing.T) {
	c := newFakeStaticCallContract()
	queryParams := &FakeQueryParams{
		Key:   "bob",
		Value: 10,
	}
	data, err := json.Marshal(queryParams)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	msg := &txmsg.SimpleContractMethodJSON{
		Version: 0,
		Method:  fmt.Sprintf("%s.QueryMethod", c.Meta().Name),
		Data:    data,
	}
	msgBytes, err := json.Marshal(msg)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	ctx := CreateFakeContext()
	req := &plugin.Request{
		ContentType: plugin.ContentType_JSON,
		Accept:      plugin.ContentType_JSON,
		Body:        msgBytes,
	}

	resp, err := c.StaticCall(ctx, req)

	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}
	if resp == nil {
		t.Errorf("Error: expected response != nil")
	}
	if resp.ContentType != plugin.ContentType_JSON {
		t.Errorf("Error: wrong response content type")
	}
	var queryResult FakeQueryResult
	if err := json.Unmarshal(resp.Body, &queryResult); err != nil {
		t.Errorf("Error: %v", err)
		return
	}
	if queryResult.Value != "alice" {
		t.Errorf("Error: unexpected query result %s", queryResult.Value)
	}
}

// TODO: TestStaticCallDispatch_PROTOBUF
