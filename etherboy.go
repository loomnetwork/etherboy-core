package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/loomnetwork/etherboy-core/txmsg"
	"github.com/loomnetwork/go-loom/plugin"
	contract "github.com/loomnetwork/go-loom/plugin/contractpb"
	"github.com/pkg/errors"
)

func main() {
	plugin.Serve(Contract)
}

type EtherBoy struct {
}

func (e *EtherBoy) Meta() (plugin.Meta, error) {
	return plugin.Meta{
		Name:    "etherboycore",
		Version: "0.0.1",
	}, nil
}

func (e *EtherBoy) Init(ctx contract.Context, req *plugin.Request) error {
	return nil
}

func (e *EtherBoy) CreateAccount(ctx contract.Context, accTx *txmsg.EtherboyCreateAccountTx) error {
	owner := strings.TrimSpace(accTx.Owner)
	// confirm owner doesnt exist already
	if ctx.Has(e.ownerKey(owner)) {
		return errors.New("Owner already exists")
	}
	addr := []byte(ctx.Message().Sender.Local)
	state := txmsg.EtherboyAppState{
		Address: addr,
	}
	if err := ctx.Set(e.ownerKey(owner), &state); err != nil {
		return errors.Wrap(err, "Error setting state")
	}
	ctx.GrantPermission([]byte(owner), []string{"owner"})
	ctx.Logger().Info("Created account", "owner", owner, "address", addr)
	emitMsg := struct {
		Owner  string
		Method string
		Addr   []byte
	}{owner, "createacct", addr}
	emitMsgJSON, err := json.Marshal(emitMsg)
	if err != nil {
		log.Println("Error marshalling emit message")
	}
	ctx.EmitTopics(emitMsgJSON, "etherboy:createaccount")
	return nil
}

func (e *EtherBoy) SaveState(ctx contract.Context, tx *txmsg.EtherboyStateTx) error {
	ctx.Logger().Debug(fmt.Sprintf("ctx: %v", ctx))
	owner := strings.TrimSpace(tx.Owner)
	var curState txmsg.EtherboyAppState
	if err := ctx.Get(e.ownerKey(owner), &curState); err != nil {
		return err
	}
	if ok, _ := ctx.HasPermission([]byte(owner), []string{"owner"}); !ok {
		return errors.New("Owner unverified")
	}
	curState.Blob = tx.Data
	if err := ctx.Set(e.ownerKey(owner), &curState); err != nil {
		return errors.Wrap(err, "Error marshaling state node")
	}
	emitMsg := struct {
		Owner  string
		Method string
		Addr   []byte
		Value  int64
	}{Owner: owner, Method: "savestate", Addr: curState.Address}
	json.Unmarshal(tx.Data, &emitMsg)
	ctx.Logger().Debug("Set state", "owner", owner, "value", emitMsg.Value)
	emitMsgJSON, err := json.Marshal(&emitMsg)
	if err != nil {
		ctx.Logger().Error("Error marshalling emit message", "error", err)
	}
	ctx.EmitTopics(emitMsgJSON, "etherboy:savestate")

	return nil
}

func (e *EtherBoy) GetState(ctx contract.StaticContext, params *txmsg.StateQueryParams) (*txmsg.StateQueryResult, error) {
	if ctx.Has(e.ownerKey(params.Owner)) {
		var curState txmsg.EtherboyAppState
		if err := ctx.Get(e.ownerKey(params.Owner), &curState); err != nil {
			return nil, err
		}
		return &txmsg.StateQueryResult{State: curState.Blob}, nil
	}
	return &txmsg.StateQueryResult{}, nil
}

func (s *EtherBoy) ownerKey(owner string) []byte {
	return []byte("owner:" + owner)
}

var Contract plugin.Contract = contract.MakePluginContract(&EtherBoy{})
