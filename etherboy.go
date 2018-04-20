// go build -buildmode=plugin -o contracts/helloworld.so plugin/examples/helloworld.go
package main

import (
	"github.com/pkg/errors"

	proto "github.com/golang/protobuf/proto"
	"github.com/loomnetwork/etherboy-core/gen"
	"github.com/loomnetwork/loom"
	"github.com/loomnetwork/loom/plugin"
	"strings"
)

func main() {
}

type EtherBoy struct {
}

func (e *EtherBoy) Meta() plugin.Meta {
	return plugin.Meta{
		Name:    "etherboycore",
		Version: "0.0.1",
	}
}

func (e *EtherBoy) Init(ctx plugin.Context, req *plugin.Request) (*plugin.Response, error) {
	println("init contract")
	return &plugin.Response{}, nil
}

func (e *EtherBoy) Call(ctx plugin.Context, req *plugin.Request) (*plugin.Response, error) {
	var tx txmsg.EtherboyAppTx
	proto.Unmarshal(req.Body, &tx)
	owner := strings.TrimSpace(tx.Owner)
	switch tx.Data.(type) {
	case *txmsg.EtherboyAppTx_CreateAccount:
		createAccTx := tx.GetCreateAccount()
		if err := e.createAccount(ctx, owner, createAccTx); err != nil {
			return e.jsonResponse(), err
		}
	case *txmsg.EtherboyAppTx_State:
		saveStateTx := tx.GetState()
		if err := e.saveState(ctx, owner, saveStateTx); err != nil {
			return e.jsonResponse(), err
		}
	}
	return &plugin.Response{}, nil
}

func (e *EtherBoy) jsonResponse() *plugin.Response {
	return &plugin.Response{
		ContentType: plugin.ContentType_JSON,
		Body:        []byte("{}"),
	}
}

func (e *EtherBoy) StaticCall(ctx plugin.StaticContext, req *plugin.Request) (*plugin.Response, error) {
	return &plugin.Response{}, nil
}

func (e *EtherBoy) ownerKey(owner string) []byte {
	return []byte("owner:" + owner)
}

func (e *EtherBoy) createAccount(ctx plugin.Context, owner string, accTx *txmsg.EtherboyCreateAccountTx) error {
	// confirm owner doesnt exist already
	if ctx.Has(e.ownerKey(owner)) {
		return errors.New("Owner already exists")
	}
	state := &txmsg.EtherboyAppState{
		Address: []byte(ctx.Message().Sender.Local),
	}
	statebytes, err := proto.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "Error marshaling state node")
	}
	ctx.Set(e.ownerKey(owner), statebytes)
	return nil
}

func (e *EtherBoy) saveState(ctx plugin.Context, owner string, tx *txmsg.EtherboyStateTx) error {
	var curState txmsg.EtherboyAppState
	proto.Unmarshal(ctx.Get(e.ownerKey(owner)), &curState)
	if loom.LocalAddress(curState.Address).Compare(ctx.Message().Sender.Local) != 0 {
		return errors.New("Owner unverified")
	}
	curState.Blob = tx.Data
	statebytes, err := proto.Marshal(&curState)
	if err != nil {
		return errors.Wrap(err, "Error marshaling state node")
	}
	ctx.Set(e.ownerKey(owner), statebytes)
	return nil
}

var Contract plugin.Contract = &EtherBoy{}
