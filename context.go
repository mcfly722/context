// Package implements a graceful shutdown context tree for your goroutines.
//
// It means that parent context wouldn't close until all its children's worked.
//
// # example:
//
// You are creating a context tree:
//
// root => child1 => child2 => child3
//
// and trying to close the root.
//
// All subchilds will close in reverse order (first - child3, then child2, child1, root).
// This closing order is absolutely essential because the child context could use some parent resources or send some signals to the parent. If a parent closes before the child, it will cause undefined behavior or goroutine locking.
//
// Unfortunately, context from the standard Go library does not guarantee this closing order.
// See issue: https://github.com/golang/go/issues/51075
//
// This module resolves this problem and guarantees a correct closing order.
package context

import (
	"sync"
)

// Instances of this interface are sent to your node through the Go() method.
//
// (see [ContextedInstance])
type Context interface {

	// creates a new child context, for instance, what implements ContextedInstance interface
	NewContextFor(instance ContextedInstance) (ChildContext, error)

	// When this channel closes, it means that the child context should exit from the Go function.
	Context() chan struct{}

	// Close the current context and all children in reverse order.
	Close()
}

type contextState int

const (
	working   contextState = 0
	freezed   contextState = 1
	disposing contextState = 2
)

type context struct {
	parent   *context
	childs   map[*context]*context
	instance ContextedInstance
	state    contextState
	isOpened chan struct{}
	root     *root
}

type root struct {
	ready sync.Mutex
}

func newEmptyContext() *context {

	return &context{
		parent:   &context{},
		childs:   map[*context]*context{},
		instance: nil,
		state:    working,
		isOpened: make(chan struct{}),
		root:     &root{},
	}
}

// NewContextFor ...
func (parent *context) NewContextFor(instance ContextedInstance) (ChildContext, error) {

	parent.root.ready.Lock()
	defer parent.root.ready.Unlock()

	switch parent.state {
	case freezed:
		return nil, &ClosingIsInProcessForFreezeError{}
	case disposing:
		return nil, &ClosingIsInProcessForDisposingError{}
	}

	return newContextFor(parent, instance)
}

func newContextFor(parent *context, instance ContextedInstance) (*context, error) {

	newContext := &context{
		parent:   parent,
		childs:   map[*context]*context{},
		instance: instance,
		state:    working,
		isOpened: make(chan struct{}),
		root:     parent.root,
	}

	parent.childs[newContext] = newContext

	// Start new Context
	go func(current *context) {

		// execure user context select {...}
		current.instance.Go(current)

		if current.state != disposing {
			panic(ExitFromContextWithoutClosePanic)
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

// Close ...
func (current *context) Close() {
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
