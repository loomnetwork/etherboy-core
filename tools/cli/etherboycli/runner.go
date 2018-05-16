package main

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/loadimpact/k6/lib"
	"github.com/loadimpact/k6/stats"
)

// LoomRunner wraps a function in a runner whose VUs will simply call that function.
type LoomRunner struct {
	Fn         func(ctx context.Context, uid string) ([]stats.SampleContainer, error)
	SetupFn    func(ctx context.Context) error
	TeardownFn func(ctx context.Context) error

	Group   *lib.Group
	Options lib.Options

	nextVUID int64
	MaxUID   int64
}

var _ lib.Runner = &LoomRunner{}

func (r *LoomRunner) VU() *LoomRunnerVU {
	atomic.AddInt64(&r.nextVUID, 1)
	id := atomic.LoadInt64(&r.nextVUID)
	return &LoomRunnerVU{R: r, ID: id, MaxUID: r.MaxUID}
}

func (r *LoomRunner) MakeArchive() *lib.Archive {
	return nil
}

func (r *LoomRunner) NewVU() (lib.VU, error) {
	return r.VU(), nil
}

func (r LoomRunner) Setup(ctx context.Context) error {
	if fn := r.SetupFn; fn != nil {
		return fn(ctx)
	}
	return nil
}

func (r LoomRunner) Teardown(ctx context.Context) error {
	if fn := r.TeardownFn; fn != nil {
		return fn(ctx)
	}
	return nil
}

func (r LoomRunner) GetDefaultGroup() *lib.Group {
	if r.Group == nil {
		r.Group = &lib.Group{}
	}
	return r.Group
}

func (r LoomRunner) GetOptions() lib.Options {
	return r.Options
}

func (r *LoomRunner) SetOptions(opts lib.Options) {
	r.Options = opts
}

// A VU spawned by a LoomRunner.
type LoomRunnerVU struct {
	R  *LoomRunner
	ID int64

	// user id for etherboy
	uid    int64
	MaxUID int64
}

func (vu *LoomRunnerVU) newUID() string {
	id := fmt.Sprintf("%d", vu.uid%vu.MaxUID)
	atomic.AddInt64(&vu.uid, 1)
	return id
}

func (vu *LoomRunnerVU) RunOnce(ctx context.Context) ([]stats.SampleContainer, error) {
	if vu.R.Fn == nil {
		return []stats.SampleContainer{}, nil
	}

	return vu.R.Fn(ctx, vu.newUID())
}

func (vu *LoomRunnerVU) Reconfigure(id int64) error {
	vu.ID = id
	return nil
}
