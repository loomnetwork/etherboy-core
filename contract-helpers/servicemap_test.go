package contracthelpers

import (
	"fmt"
	"testing"

	"github.com/loomnetwork/loom/plugin"
)

type FakeTx struct{}

type fakeTx struct{}

type fakeResponse struct{}

type fakeContract struct{}

// These methods SHOULD NOT be auto-registered:
func (c *fakeContract) IgnoredMethod1()                               {}
func (c *fakeContract) ignoredMethod2()                               {}
func (c *fakeContract) IgnoredMethod3(ctx plugin.Context)             {}
func (c *fakeContract) IgnoredMethod4(ctx plugin.Context, tx *FakeTx) {}
func (c *fakeContract) IgnoredMethod5(ctx plugin.Context, tx *FakeTx) int {
	return 0
}

// This method will be ignored because the type of the second argument is not exported
func (c *fakeContract) IgnoredMethod6(ctx plugin.Context, tx *fakeTx) error {
	return nil
}

// func (c *fakeContract) InvalidMethod7(ctx plugin.Context, tx *FakeTx) (*fakeResponse, error) {}

// These methods SHOULD be auto-registered
func (c *fakeContract) Method1(ctx plugin.Context, tx *FakeTx) error {
	return nil
}

// func (c *fakeContract) ValidMethod2(ctx plugin.Context, tx *FakeTx) (*FakeResponse, error) {}

func TestServiceMapDuplicateServices(t *testing.T) {
	srvMap := new(serviceMap)
	if err := srvMap.Register(&fakeContract{}, "fakeContract"); err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	if err := srvMap.Register(&fakeContract{}, "fakeContract"); err == nil {
		t.Errorf("Error: duplicate service names should not be allowed")
	}
}

func TestServiceMapAutoDiscovery(t *testing.T) {
	srvMap := new(serviceMap)
	if err := srvMap.Register(&fakeContract{}, "fakeContract"); err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	for i := 1; i < 7; i++ {
		methodName := fmt.Sprintf("fakeContract.IgnoredMethod%d", i)
		if _, _, err := srvMap.Get(methodName); err == nil {
			t.Errorf("Error: %s should not be registered", methodName)
		}
	}

	for i := 1; i < 2; i++ {
		methodName := fmt.Sprintf("fakeContract.Method%d", i)
		if _, _, err := srvMap.Get(methodName); err != nil {
			t.Errorf("Error: %s should be registered", methodName)
		}
	}
}