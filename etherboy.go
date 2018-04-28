package main

import (
	"encoding/json"
	"log"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	"github.com/loomnetwork/etherboy-core/txmsg"
	loom "github.com/loomnetwork/go-loom"
	"github.com/loomnetwork/go-loom/plugin"
	"github.com/pkg/errors"
)

func main() {}

type EtherBoy struct {
	plugin.RequestDispatcher
}

func (e *EtherBoy) Meta() (plugin.Meta, error) {
	return plugin.Meta{
		Name:    "etherboycore",
		Version: "0.0.1",
	}, nil
}

func (e *EtherBoy) Init(ctx plugin.Context, req *plugin.Request) error {
	return nil
}

func (e *EtherBoy) CreateAccount(ctx plugin.Context, accTx *txmsg.EtherboyCreateAccountTx) error {
	owner := strings.TrimSpace(accTx.Owner)
	// confirm owner doesnt exist already
	if ctx.Has(e.ownerKey(owner)) {
		return errors.New("Owner already exists")
	}
	addr := []byte(ctx.Message().Sender.Local)
	state := &txmsg.EtherboyAppState{
		Address: addr,
	}
	statebytes, err := proto.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "Error marshaling state node")
	}
	ctx.Set(e.ownerKey(owner), statebytes)
	emitMsg := &struct {
		Owner  string
		Method string
		Addr   []byte
	}{owner, "createacct", addr}
	emitMsgJSON, err := json.Marshal(emitMsg)
	if err != nil {
		log.Println("Error marshalling emit message")
	}
	ctx.Emit(emitMsgJSON)
	return nil
}

func (e *EtherBoy) SaveState(ctx plugin.Context, tx *txmsg.EtherboyStateTx) error {
	owner := strings.TrimSpace(tx.Owner)
	var curState txmsg.EtherboyAppState
	if err := proto.Unmarshal(ctx.Get(e.ownerKey(owner)), &curState); err != nil {
		return err
	}
	if loom.LocalAddress(curState.Address).Compare(ctx.Message().Sender.Local) != 0 {
		return errors.New("Owner unverified")
	}
	curState.Blob = tx.Data
	statebytes, err := proto.Marshal(&curState)
	if err != nil {
		return errors.Wrap(err, "Error marshaling state node")
	}
	ctx.Set(e.ownerKey(owner), statebytes)
	emitMsg := &struct {
		Owner  string
		Method string
		Addr   []byte
		Value  int64
	}{Owner: owner, Method: "savestate", Addr: curState.Address}
	json.Unmarshal(tx.Data, emitMsg)
	emitMsgJSON, err := json.Marshal(emitMsg)
	if err != nil {
		log.Println("Error marshalling emit message")
	}
	log.Printf("Emitting: %s\n", string(emitMsgJSON))
	ctx.Emit(emitMsgJSON)

	return nil
}

func (e *EtherBoy) GetState(ctx plugin.Context, params *txmsg.StateQueryParams) (*txmsg.StateQueryResult, error) {
	if ctx.Has(e.ownerKey(params.Owner)) {
		statebytes := ctx.Get(e.ownerKey(params.Owner))
		var curState txmsg.EtherboyAppState
		if err := proto.Unmarshal(statebytes, &curState); err != nil {
			return nil, err
		}
		return &txmsg.StateQueryResult{State: curState.Blob}, nil
	}
	return &txmsg.StateQueryResult{}, nil
}

func (s *EtherBoy) ownerKey(owner string) []byte {
	return []byte("owner:" + owner)
}

func NewEtherBoyContract() plugin.Contract {
	e := &EtherBoy{}
	e.RequestDispatcher.Init(e)
	return e
}

var Contract = NewEtherBoyContract()
