package main

import (
	"encoding/json"
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
	// Verify Tokens Transfer Status
	if ctx.Has(e.transferTokenKey(owner)) {
		return errors.New("Tokens already transferred")
	}

	// Transfer Tokens if not already
	localAddr := ctx.Message().Sender.Local.String()
	toAddr := loom.MustParseAddress(ctx.Message().Sender.ChainID + ":" + localAddr)
	coinAddr, err := ctx.Resolve("coin")
	if err != nil {
		ctx.Logger().Info("Cannot load coin contract", err)
		return err
	}
	approveMsg := &ctypes.ApproveRequest{
		Spender: toAddr.MarshalPB(),
		Amount: &types.BigUInt{
			Value: *loom.NewBigUIntFromInt(1),
		},
	}
	appResp := &ctypes.ApproveResponse{}
	err1 := contract.CallMethod(ctx, coinAddr, "Approve", approveMsg, appResp)
	ctx.Logger().Info("Approve Response", err1)
	amount := loom.NewBigUIntFromInt(1)
	msg := &ctypes.TransferFromRequest{
		From:   toAddr.MarshalPB(),
		To:     loom.MustParseAddress("default:0x06D313A35B77B0Ef70d3741022f79E3f5A56A971").MarshalPB(),
		Amount: &types.BigUInt{Value: *amount},
	}
	resp := &ctypes.TransferFromResponse{}
	err2 := contract.CallMethod(ctx, coinAddr, "TransferFrom", msg, resp)

	// Mark State with tokens transferre

	if err := ctx.Set(e.transferTokenKey(owner), tx); err != nil {
		return errors.Wrap(err, "Error setting state")
	}
	ctx.Logger().Info("Transfer From Response", err2)
	return err2
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

var Contract plugin.Contract = contract.MakePluginContract(&EtherBoy{})
