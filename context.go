package context

import (
	"errors"
	"sync"
	"time"

	"github.com/mcfly722/goPackages/scheduler"
)

// Reason ...
type Reason error

// Disposer ...
type Disposer func(reason Reason)

// Finalizer ...
type Finalizer func(reason Reason)

// ReasonCanceled ...
var ReasonCanceled = (Reason)(errors.New("context canceled"))

// ReasonOutdated ...
var ReasonOutdated = (Reason)(errors.New("context outdated"))

// ErrCreatingOnCanceled ...
var ErrCreatingOnCanceled = errors.New("Context already canceled. Creating new childs for canceled context are prohibited")

// Context ...
type Context interface {
	NewChildContext(Disposer, Finalizer) (Context, error)

	SetDeadline(time.Time)

	Cancel(Reason)
}

type ctx struct {
	id          int64
	parent      *ctx
	childs      map[int64]*ctx
	nextChildID int64
	tree        *tree
	disposer    Disposer
	finalizer   Finalizer
	canceled    Reason
	ready       sync.Mutex
}

type tree struct {
	scheduler      scheduler.Scheduler
	onDestroy      chan bool
	onCancel       chan cancel
	changesAllowed sync.Mutex
}

type cancel struct {
	context *ctx
	reason  Reason
}

func (context *ctx) SetDeadline(deadline time.Time) {
	context.ready.Lock()
	defer context.ready.Unlock()

	context.tree.scheduler.RegisterNewTimer(deadline, context)
}

func (context *ctx) cancel(reason Reason) {

	if context.canceled != nil {
		panic("trying to cancel already cancelled context")
	}

	// dispose context
	if context.disposer != nil {
		context.disposer(reason)
	}

	// delete all timers for current context from tree
	context.tree.scheduler.CancelTimerFor(context)

	// unbind context from it parent or destroy tree if it is root context
	if context.parent != nil {
		delete(context.parent.childs, context.id)
	}

	context.canceled = reason

	// finish context
	if context.finalizer != nil {
		context.finalizer(reason)
	}
}

func (context *ctx) cancelRecursively(reason Reason) {
	childs := context.childs // this copy are required because parent context we will unlink from childs map

	for _, child := range childs {
		child.cancelRecursively(reason)
	}

	context.cancel(reason)
}

// NewContextTree ...
func NewContextTree(disposer Disposer, finalizer Finalizer) Context {

	tree := &tree{
		scheduler: scheduler.NewScheduler(),
		onCancel:  make(chan cancel),
		onDestroy: make(chan bool),
	}

	newContext := &ctx{
		id:          0,
		childs:      make(map[int64]*ctx),
		nextChildID: 1,
		tree:        tree,
		disposer:    disposer,
		finalizer:   finalizer,
		canceled:    nil,
	}

	go func() {
		for {
			select {
			case cancel := <-tree.onCancel:
				cancel.context.tree.changesAllowed.Lock()
				defer cancel.context.tree.changesAllowed.Unlock()
				cancel.context.cancelRecursively(cancel.reason)
				if cancel.context.parent == nil { // cancelling root node
					return
				}
			default:
				outdatedContext := tree.scheduler.TakeFirstOutdatedOrNil()
				if outdatedContext != nil {
					tree.changesAllowed.Lock()
					outdatedContext.(*ctx).cancelRecursively(ReasonOutdated)
					tree.changesAllowed.Unlock()
				}
			}
		}
	}()

	return newContext
}

// NewChildContext ...
func (context *ctx) NewChildContext(disposer Disposer, finalizer Finalizer) (Context, error) {
	context.tree.changesAllowed.Lock()
	defer context.tree.changesAllowed.Unlock()

	if context.canceled != nil {
		return nil, ErrCreatingOnCanceled
	}

	newContext := &ctx{
		id:          context.nextChildID,
		parent:      context,
		childs:      make(map[int64]*ctx),
		nextChildID: 0,
		tree:        context.tree,
		disposer:    disposer,
		finalizer:   finalizer,
		canceled:    nil,
	}

	context.childs[context.nextChildID] = newContext
	context.nextChildID++

	return newContext, nil
}

// Cancel with reason Canceled(done)/Outdated/etc...
func (context *ctx) Cancel(reason Reason) {
	go func() {
		context.tree.onCancel <- cancel{
			context: context,
			reason:  reason,
		}
	}()
}
