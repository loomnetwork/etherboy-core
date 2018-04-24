package main

import (
	"encoding/json"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	"github.com/loomnetwork/etherboy-core/txmsg"
	loom "github.com/loomnetwork/loom-plugin"
	"github.com/loomnetwork/loom-plugin/plugin"
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
	return nil
}

// NOTE: These could be defined as protobufs instead.
type StateQueryParams struct {
	Owner string `json:"owner"`
}

type StateQueryResult struct {
	State json.RawMessage `json:"state"`
}

func (e *EtherBoy) GetState(ctx plugin.Context, params *StateQueryParams) (*StateQueryResult, error) {
	if ctx.Has(e.ownerKey(params.Owner)) {
		statebytes := ctx.Get(e.ownerKey(params.Owner))
		var curState txmsg.EtherboyAppState
		if err := proto.Unmarshal(statebytes, &curState); err != nil {
			return nil, err
		}
		return &StateQueryResult{State: curState.Blob}, nil
	}
	return &StateQueryResult{}, nil
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
