package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gogo/protobuf/jsonpb"
	proto "github.com/gogo/protobuf/proto"
	"github.com/loomnetwork/etherboy-core/testdata"
	"github.com/loomnetwork/etherboy-core/txmsg"
	plugin "github.com/loomnetwork/go-loom/plugin"
	"github.com/loomnetwork/go-loom/types"
	"github.com/pkg/errors"
)

func TestCreateAccount(t *testing.T) {
	e := NewEtherBoyContract()
	meta, err := e.Meta()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	tx := &txmsg.EtherboyCreateAccountTx{
		Version: 0,
		Owner:   "aditya",
		Data:    []byte("dummy"),
	}

	txBytes, err := proto.Marshal(tx)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	msg := &types.ContractMethodCall{
		Method: fmt.Sprintf("%s.CreateAccount", meta.Name),
		Args:   txBytes,
	}

	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	fmt.Printf("Data: msgBytes-%v\n", msgBytes)

	ctx := CreateFakeContext()
	req := &plugin.Request{
		ContentType: plugin.EncodingType_PROTOBUF3,
		Accept:      plugin.EncodingType_PROTOBUF3,
		Body:        msgBytes,
	}

	resp, err := e.Call(ctx, req)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	fmt.Printf("resp -%v\n", resp)
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

func (c *fakeStaticCallContract) QueryMethod(ctx plugin.Context, params *testdata.FakeQueryParams) (*testdata.FakeQueryResult, error) {
	return &testdata.FakeQueryResult{
		Value: "alice",
	}, nil
}

func (c *fakeStaticCallContract) FailingQueryMethod(ctx plugin.Context, params *testdata.FakeQueryParams) (*testdata.FakeQueryResult, error) {
	return nil, errors.New("query failed")
}

func newFakeStaticCallContract() plugin.Contract {
	c := &fakeStaticCallContract{}
	c.RequestDispatcher.Init(c)
	return c
}

func TestStaticCallDispatch_JSON(t *testing.T) {
	c := newFakeStaticCallContract()
	queryParams := &testdata.FakeQueryParams{
		Key:   "bob",
		Value: 10,
	}

	var data bytes.Buffer
	marsh := jsonpb.Marshaler{}
	err := marsh.Marshal(&data, queryParams)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	meta, err := c.Meta()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	msg := &types.ContractMethodCall{
		Method: fmt.Sprintf("%s.QueryMethod", meta.Name),
		Args:   data.Bytes(),
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
	var queryResult testdata.FakeQueryResult
	unmarsh := jsonpb.Unmarshaler{}
	if err := unmarsh.Unmarshal(bytes.NewBuffer(resp.Body), &queryResult); err != nil {
		t.Errorf("Error: %v", err)
		return
	}
	if queryResult.Value != "alice" {
		t.Errorf("Error: unexpected query result %s", queryResult.Value)
	}
}

// TODO: TestStaticCallDispatch_PROTOBUF
