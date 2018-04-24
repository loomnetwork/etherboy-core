package main

import (
	"encoding/json"
	"fmt"
	"testing"

	proto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/loomnetwork/etherboy-core/txmsg"
	plugin "github.com/loomnetwork/loom-plugin/plugin"
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
	meta, err := e.Meta()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	msg := &plugin.ContractMethodCall{
		Method: fmt.Sprintf("%s.CreateAccount", meta.Name),
		Data:   any,
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
	plugin.RequestDispatcher
}

func (e *fakeStaticCallContract) Meta() (plugin.Meta, error) {
	return plugin.Meta{
		Name:    "fakeContract",
		Version: "0.0.1",
	}, nil
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

	meta, err := c.Meta()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	msg := &plugin.ContractMethodCallJSON{
		Method: fmt.Sprintf("%s.QueryMethod", meta.Name),
		Data:   data,
	}
	msgBytes, err := json.Marshal(msg)

	if err != nil {
		t.Errorf("Error: %v", err)
	}

	ctx := CreateFakeContext()
	req := &plugin.Request{
		ContentType: plugin.EncodingType_JSON,
		Accept:      plugin.EncodingType_JSON,
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
	if resp.ContentType != plugin.EncodingType_JSON {
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
