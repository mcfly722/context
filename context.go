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
type Context[M any] interface {

	// creates a new child context, for instance, what implements ContextedInstance interface
	NewContextFor(instance ContextedInstance[M]) (ChildContext[M], error)

	// Channel that transmits control state messages from parent to child. When it closes, it means that the child context should exit from the Go function.
	Controller() chan M

	// Close the current context and all children in reverse order.
	Close()
}

type contextState int

const (
	working   contextState = 0
	freezed   contextState = 1
	disposing contextState = 2
)

type context[M any] struct {
	parent     *context[M]
	childs     map[*context[M]]*context[M]
	instance   ContextedInstance[M]
	state      contextState
	controller chan M
	root       *root

	//controllerReady sync.Mutex // this mutex is essential to resolve runtime.closechan() and runtime.chansend() race. Unfortunately this two runtime functions are not thread safe (issue #30372)
}

type root struct {
	ready sync.Mutex
}

func newEmptyContext[M any]() *context[M] {

	return &context[M]{
		parent:     &context[M]{},
		childs:     map[*context[M]]*context[M]{},
		instance:   nil,
		state:      working,
		controller: make(chan M),
		root:       &root{},
	}
}

// NewContextFor ...
func (parent *context[M]) NewContextFor(instance ContextedInstance[M]) (ChildContext[M], error) {

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

func newContextFor[M any](parent *context[M], instance ContextedInstance[M]) (*context[M], error) {

	newContext := &context[M]{
		parent:     parent,
		childs:     map[*context[M]]*context[M]{},
		instance:   instance,
		state:      working,
		controller: make(chan M),
		root:       parent.root,
	}

	parent.childs[newContext] = newContext

	// Start new Context
	go func(current *context[M]) {

		// start user context select {...} loop
		current.instance.Go(current)

		if current.state != disposing {
			panic(ExitFromContextWithoutClosePanic)
		}

		// remove the node from parent children, and if the parent is frozen and empty, initiate its disposal
		current.root.ready.Lock()
		if current.parent != nil {
			delete(current.parent.childs, current)
			if current.parent.state == freezed && len(current.parent.childs) == 0 {
				current.parent.state = disposing
				current.parent.closeControllerChannel() // instead close(current.parent.controller)
			}
		}
		current.root.ready.Unlock()

	}(newContext)

	return newContext, nil
}

// Context ...
func (context *context[M]) Controller() chan M {
	return context.controller
}

func (context *context[M]) closeControllerChannel() {
	//context.controllerReady.Lock()
	close(context.controller)
	//context.controllerReady.Unlock()
}

// Close ...
func (current *context[M]) Close() {
	current.root.ready.Lock()
	defer current.root.ready.Unlock()

	current.freezeAllChildsAndSubchilds()
}

func (current *context[M]) freezeAllChildsAndSubchilds() {

	if current.state == working {
		current.state = freezed
		for child := range current.childs {
			child.freezeAllChildsAndSubchilds()
		}
	}

	if current.state == freezed && len(current.childs) == 0 {
		current.state = disposing
		current.closeControllerChannel() // instead close(current.controller)

	}
}

// Send a message to the context.
func (current *context[M]) Send(message M) (err error) {

	//current.controllerReady.Lock()
	//defer current.controllerReady.Unlock()

	defer func() {
		if r := recover(); r != nil {
			err = &ClosingIsInProcessForSendError{}
		}
	}()

	current.controller <- message

	return nil
}
