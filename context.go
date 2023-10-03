// Package implements graceful shutdown context tree for your goroutines.
//
// It means that parent context wouldn't close until all its child's doesn't close.
//
// # example:
//
// You creating context tree:
//
// root -> child1 -> child2 -> child3
//
// and trying to Close() root.
//
// All subchilds would be closed in reverse order (first - child3, then child2, child1, root).
// This closing order is absolutely essential, because child context could use some parent resources or send some signals to parent. If parent would be closed before it child, it will cause undefined behaviour or goroutine locking.
//
// Unfortunately, context from standard Go library does not guarantee this close order.
//
// See issue: https://github.com/golang/go/issues/51075
//
// This module resolves this problem and guarantee correct closing.
package context

import (
	"sync"
)

// Instances of this interfaces sends to your node through Go() method.
//
// (see [ContextedInstance])
type Context interface {

	// Method creates new child context for instance what implements ContextedInstance interface
	NewContextFor(instance ContextedInstance) (Context, error)

	// Method receives channel what could be used to understand when you can close your current context (all childs are already served and terminated).
	Context() chan struct{}

	// Method cancel current context and all childs according reverse order.
	Cancel()

	// Method used to set defer handler function to recover from panics inside Go(...) method of your [ContextedInstance]
	SetDefer(func(interface{}))
}

type contextState int

const (
	working   contextState = 0
	freezed   contextState = 1
	disposing contextState = 2
)

type context struct {
	parent       *context
	childs       map[*context]*context
	instance     ContextedInstance
	state        contextState
	isOpened     chan struct{}
	root         *root
	deferHandler *func(interface{})
}

type root struct {
	ready sync.Mutex
}

func newEmptyContext() *context {

	return &context{
		parent:       &context{},
		childs:       map[*context]*context{},
		instance:     nil,
		state:        working,
		isOpened:     make(chan struct{}),
		root:         &root{},
		deferHandler: nil,
	}
}

// NewContextFor ...
func (parent *context) NewContextFor(instance ContextedInstance) (Context, error) {

	parent.root.ready.Lock()
	defer parent.root.ready.Unlock()

	switch parent.state {
	case freezed:
		return nil, &CancelInProcessForFreezeError{}
	case disposing:
		return nil, &CancelInProcessForDisposingError{}
	}

	return newContextFor(parent, instance)
}

func newContextFor(parent *context, instance ContextedInstance) (Context, error) {

	newContext := &context{
		parent:       parent,
		childs:       map[*context]*context{},
		instance:     instance,
		state:        working,
		isOpened:     make(chan struct{}),
		root:         parent.root,
		deferHandler: nil,
	}

	parent.childs[newContext] = newContext

	// Start new Context
	go func(current *context) {

		// set defer handler
		defer func() {
			current.root.ready.Lock()
			defer current.root.ready.Unlock()
			if current.deferHandler != nil {
				deferHandler := *(current.deferHandler)
				deferHandler(recover()) // unfortunatelly recover() don't catch panic if you try to call it from function handler, so you need catch it here
			}
		}()

		// execure user context select {...}
		current.instance.Go(current)

		if current.state != disposing {
			panic(ExitFromContextWithoutCancelPanic)
		}

		// Remove node from parent childs and if parent is freezed and empty, initiate it disposing
		current.root.ready.Lock()
		if current.parent != nil {
			delete(current.parent.childs, current)
			if current.parent.state == freezed && len(current.parent.childs) == 0 {
				current.parent.state = disposing
				close(current.parent.isOpened)
			}
		}
		current.root.ready.Unlock()

	}(newContext)

	return newContext, nil
}

// Context ...
func (context *context) Context() chan struct{} {
	return context.isOpened
}

// Cancel ...
func (current *context) Cancel() {
	current.root.ready.Lock()
	defer current.root.ready.Unlock()

	current.freezeAllChildsAndSubchilds()
}

func (current *context) freezeAllChildsAndSubchilds() {

	if current.state == working {
		current.state = freezed
		for child := range current.childs {
			child.freezeAllChildsAndSubchilds()
		}
	}

	if current.state == freezed && len(current.childs) == 0 {
		current.state = disposing
		close(current.isOpened)
	}
}

func (current *context) SetDefer(deferFunction func(interface{})) {
	current.root.ready.Lock()
	current.deferHandler = &deferFunction
	current.root.ready.Unlock()
}
