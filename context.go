package context

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mcfly722/goPackages/scheduler"
)

// ErrCanceled ...
var ErrCanceled = errors.New("context canceled")

// ErrOutdated ...
var ErrOutdated = errors.New("context outdated")

// Disposer ...
type Disposer func(err error)

// Context ...
type Context interface {
	NewChildContext(name string) Context

	GetChild(id int64) Context

	GetName() string

	SetDisposer(Disposer)

	SetDeadline(time.Time)

	Cancel(err error)

	OnDone() chan error
}

type ctx struct {
	id              int64
	name            string
	parent          *ctx
	childs          map[int64]*ctx
	nextChildID     int64
	onDone          chan error
	tree            *tree
	disposer        Disposer
	disposedWithErr error
}

type tree struct {
	scheduler      scheduler.Scheduler
	onDestroy      chan bool
	changesAllowed sync.Mutex
}

func (context *ctx) SetDisposer(disposer Disposer) {

	context.tree.changesAllowed.Lock()
	fmt.Printf("1+")
	context.disposer = disposer
	fmt.Printf("1-")
	context.tree.changesAllowed.Unlock()
}

func (context *ctx) SetDeadline(deadline time.Time) {

	context.tree.changesAllowed.Lock()
	fmt.Printf("2+")
	context.tree.scheduler.RegisterNewTimer(deadline, context)
	fmt.Printf("2-")
	context.tree.changesAllowed.Unlock()
}

func (context *ctx) cancel(err error) {
	// calling disposer to dispose context resources
	if context.disposer != nil {
		context.disposer(err)
	}
	// delete all timers for current context from tree
	context.tree.scheduler.CancelTimerFor(context)

	// unbind context from it parent or destroy tree
	if context.parent == nil {
		context.tree.onDestroy <- true
	} else {
		delete(context.parent.childs, context.id)
	}

	context.onDone <- err

	context.disposedWithErr = err
}

func (context *ctx) cancelRecursively(err error) {
	childs := context.childs // this copy are required because parent context we will unlink from childs map

	for _, child := range childs {
		child.cancelRecursively(err)
	}

	context.cancel(err)
}

// NewContextTree ...
func NewContextTree(name string) Context {

	tree := &tree{
		scheduler: scheduler.NewScheduler(),
		onDestroy: make(chan bool),
	}

	newContext := &ctx{
		id:          0,
		name:        name,
		childs:      make(map[int64]*ctx),
		nextChildID: 1,
		tree:        tree,
		onDone:      make(chan error),
	}

	go func() {
		for {
			select {
			case <-tree.onDestroy:
				return
			default:
				outdatedContext := tree.scheduler.TakeFirstOutdatedOrNil()
				if outdatedContext != nil {
					tree.changesAllowed.Lock()
					outdatedContext.(*ctx).cancelRecursively(ErrOutdated)
					tree.changesAllowed.Unlock()
				}
			}
		}
	}()

	return newContext
}

// NewChildContext ...
func (context *ctx) NewChildContext(name string) Context {

	context.tree.changesAllowed.Lock()
	//	fmt.Printf("4+")

	defer context.tree.changesAllowed.Unlock()

	newContext := &ctx{
		id:          context.nextChildID,
		name:        name,
		parent:      context,
		childs:      make(map[int64]*ctx),
		nextChildID: 0,
		tree:        context.tree,
		onDone:      make(chan error),
	}

	context.childs[context.nextChildID] = newContext
	context.nextChildID++

	if context.disposedWithErr != nil {
		go func() {
			newContext.cancel(context.disposedWithErr)
		}()
	}

	//	fmt.Printf("4-")

	return newContext
}

// Cancel with reason Canceled(done)/Outdated
func (context *ctx) Cancel(err error) {

	context.tree.changesAllowed.Lock()
	//	fmt.Printf("5+")
	context.cancelRecursively(err)
	//	fmt.Printf("5-")
	context.tree.changesAllowed.Unlock()
}

func (context *ctx) OnDone() chan error {
	return context.onDone
}

func (context *ctx) GetName() string {
	return context.name
}

// GetChildContext ...
func (context *ctx) GetChild(id int64) Context {

	context.tree.changesAllowed.Lock()

	defer func() {
		context.tree.changesAllowed.Unlock()
	}()

	if child, ok := context.childs[id]; ok {
		return child
	}

	return nil
}
