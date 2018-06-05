package main

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/loomnetwork/etherboy-core/txmsg"
	loom "github.com/loomnetwork/go-loom"
	ctypes "github.com/loomnetwork/go-loom/builtin/types/coin"
	"github.com/loomnetwork/go-loom/plugin"
	contract "github.com/loomnetwork/go-loom/plugin/contractpb"
	types "github.com/loomnetwork/go-loom/types"
	"github.com/pkg/errors"
	"log"
	"strings"
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
	ctx.Emit(emitMsgJSON)
	return nil
}

func (e *EtherBoy) SaveState(ctx contract.Context, tx *txmsg.EtherboyStateTx) error {
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
	ctx.Emit(emitMsgJSON)

	return nil
}

func (e *EtherBoy) EndGame(ctx contract.Context, tx *txmsg.EtherboyEndGameTx) error {
	owner := strings.TrimSpace(tx.Owner)
	if ok, _ := ctx.HasPermission([]byte(owner), []string{"owner"}); !ok {
		return errors.New("Owner unverified")
	}

	addr := []byte(ctx.Message().Sender.Local)
	localAddr := ctx.Message().Sender.Local.String()
	if ctx.Has(e.endGameKey(owner)) {
		return errors.New("Game already completed")
	}

	amount := loom.NewBigUIntFromInt(1)
	toAddr := loom.MustParseAddress(ctx.Message().Sender.ChainID + ":" + localAddr)
	coinAddr, err := ctx.Resolve("coin")
	if err != nil {
		ctx.Logger().Info("Cannot load coin contract", err)
		return err
	}
	msg := &ctypes.TransferRequest{
		To:     toAddr.MarshalPB(),
		Amount: &types.BigUInt{Value: *amount},
	}
	resp := &ctypes.TransferResponse{}
	err1 := contract.CallMethod(ctx, coinAddr, "Transfer", msg, resp)
	state := txmsg.EtherboyAppState{
		Address: addr,
	}

	if err := ctx.Set(e.endGameKey(owner), &state); err != nil {
		return errors.Wrap(err, "Error setting state")
	}

	return err1
}

func (e *EtherBoy) TransferToken(ctx contract.Context, tx *txmsg.EtherboyTransferTokenTx) error {
	// Verify Owner
	owner := strings.TrimSpace(tx.Owner)
	if ok, _ := ctx.HasPermission([]byte(owner), []string{"owner"}); !ok {
		return errors.New("Owner unverified")
	}
	// Verify Tokens Transfer Transaction Status
	if ctx.Has(e.transferTokenKey(owner)) {
		return errors.New("Tokens already transferred")
	}
	h := sha256.New()
	txReceipt, err := proto.Marshal(tx)
	h.Write(txReceipt)
	txHash := h.Sum(nil)

	emitMsg := struct {
		Owner   string
		ToChain string
		ToAddr  string
		Hash    []byte
	}{tx.Owner, tx.ToAddr.ChainId, string(tx.ToAddr.Local), txHash}

	emitMsgJSON, err := json.Marshal(emitMsg)
	if err != nil {
		log.Println("Error marshalling emit message")
	}
	ctx.Emit(emitMsgJSON)

	// Mark State with tokens transfered
	addr := []byte(ctx.Message().Sender.Local)
	state := txmsg.EtherboyAppState{
		Address: addr,
		Blob:    txHash,
	}
	err = ctx.Set(e.transferTokenKey(owner), tx)
	if err != nil {
		return errors.Wrap(err, "Error setting state")
	}
	err = ctx.Set(e.transferTokenHashKey(owner), &state)
	if err != nil {
		return errors.Wrap(err, "Error setting state")
	}

	return nil
}

func (e *EtherBoy) GetTransferTokenStatus(ctx contract.StaticContext, params *txmsg.StateQueryParams) (*txmsg.StateQueryResult, error) {
	if ctx.Has(e.transferTokenHashKey(params.Owner)) {
		var curState txmsg.EtherboyAppState
		if err := ctx.Get(e.transferTokenHashKey(params.Owner), &curState); err != nil {
			return nil, err
		}
		return &txmsg.StateQueryResult{State: curState.Blob}, nil
	}
	return &txmsg.StateQueryResult{}, nil
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

func (s *EtherBoy) endGameKey(owner string) []byte {
	return []byte("endGame:" + owner)
}

func (s *EtherBoy) transferTokenKey(owner string) []byte {
	return []byte("transferToken:" + owner)
}

func (s *EtherBoy) transferTokenHashKey(owner string) []byte {
	return []byte("transferTokenHash:" + owner)
}

var Contract plugin.Contract = contract.MakePluginContract(&EtherBoy{})
