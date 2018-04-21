package main

import (
	//	"go/types"

	"fmt"

	"github.com/pkg/errors"

	proto "github.com/golang/protobuf/proto"
	"github.com/loomnetwork/etherboy-core/txmsg"
	"github.com/loomnetwork/loom"
	"github.com/loomnetwork/loom/plugin"
)

func main() {}

type EtherBoy struct {
	SimpleContract
}

func (e *EtherBoy) Meta() plugin.Meta {
	return plugin.Meta{
		Name:    "etherboycore",
		Version: "0.0.1",
	}
}

func (e *EtherBoy) Init(ctx plugin.Context, req *plugin.Request) error {
	err := e.SInit(ctx, req)
	if err != nil {
		return err
	}

	return e.RegisterService(e)
}

func (e *EtherBoy) CreateAccount(ctx plugin.Context, owner string, accTx *txmsg.EtherboyCreateAccountTx) error {
	// confirm owner doesnt exist already
	//	if ctx.Has(e.ownerKey(owner)) {
	//		return errors.New("Owner already exists")
	//	}

	fmt.Printf("accTx---%v\n", accTx)
	fmt.Printf("accTx---%T\n", accTx)
	fmt.Printf("accTx---%v\n", accTx.Data)
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

func (e *EtherBoy) SaveState(ctx plugin.Context, owner string, tx *txmsg.EtherboyStateTx) error {
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
