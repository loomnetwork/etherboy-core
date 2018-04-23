package contracthelpers

import (
	"testing"

	"github.com/loomnetwork/loom/plugin"
	"github.com/stretchr/testify/require"
)

type fakeRequestDispatcherContract struct {
	RequestDispatcher
}

func (c *fakeRequestDispatcherContract) Meta() plugin.Meta {
	return plugin.Meta{
		Name:    "fakecontract",
		Version: "0.0.1",
	}
}

func (c *fakeRequestDispatcherContract) Init(ctx plugin.Context, req *plugin.Request) error {
	return nil
}

func (c *fakeRequestDispatcherContract) HandleTx(ctx plugin.Context, req *plugin.Request) error {
	return nil
}

func (c *fakeRequestDispatcherContract) HandleQuery(ctx plugin.Context, req *plugin.Request) (*plugin.Response, error) {
	return nil, nil
}

func newFakeRequestDispatcherContract() *fakeRequestDispatcherContract {
	c := &fakeRequestDispatcherContract{}
	c.RequestDispatcher.Init(c)
	return c
}

func TestEmbeddedRequestDispatcherDoesNotRegisterOwnMethods(t *testing.T) {
	c := newFakeRequestDispatcherContract()
	var err error
	_, _, err = c.RequestDispatcher.callbacks.Get("fakecontract.Call", false)
	require.NotNil(t, err)
	_, _, err = c.RequestDispatcher.callbacks.Get("fakecontract.StaticCall", true)
	require.NotNil(t, err)
}
