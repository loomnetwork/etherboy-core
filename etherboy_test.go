package main

import (
	"fmt"
	"testing"

	"github.com/loomnetwork/etherboy-core/txmsg"
	loom "github.com/loomnetwork/go-loom"
	"github.com/loomnetwork/go-loom/plugin"
	"github.com/loomnetwork/go-loom/plugin/contractpb"
)

func TestCreateAccount(t *testing.T) {
	e := &EtherBoy{}
	meta, err := e.Meta()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	fmt.Printf("got meta -%s\n", meta)

	tx := &txmsg.EtherboyCreateAccountTx{
		Version: 0,
		Owner:   "aditya",
		Data:    []byte("dummy"),
	}
	addr1 := loom.MustParseAddress("chain:0xb16a379ec18d4093666f8f38b11a3071c920207d")
	ctx := contractpb.WrapPluginContext(plugin.CreateFakeContext(addr1, addr1))

	err = e.CreateAccount(ctx, tx)
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
