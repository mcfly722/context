// Package implements graceful shutdown context tree for your goroutines.
//
// It means that parent context wouldn't finish until all its child's worked.
//
// # example:
//
// You creating context tree:
//
// root => child1 => child2 => child3
//
// and trying to finish root.
//
// All subchilds will finish in reverse order (first - child3, then child2, child1, root).
// This closing order is absolutely essential, because child context could use some parent resources or send some signals to parent. If parent will finishes before child, it will cause undefined behaviour or goroutine locking.
//
// Unfortunately, context from standard Go library does not guarantee this closing order.
//
// See issue: https://github.com/golang/go/issues/51075
//
// This module resolves this problem and guarantee correct closing order.
package context

import (
	"sync"
)

// Instances of this interfaces sends to your node through Go() method.
//
// (see [ContextedInstance])
type Context[M any] interface {

	// Creates new child context for instance what implements ContextedInstance interface
	NewContextFor(instance ContextedInstance[M]) (ChildContext[M], error)

	// Channel that transmits control state messages from parent to child. When it closes, it means that child context should exit from Go function
	Controller() chan M

	// Finish current context and all childs according reverse order.
	Finish()

	// Send control message
	Send(message M) (err error)
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
		return nil, &FinishInProcessForFreezeError{}
	case disposing:
		return nil, &FinishInProcessForDisposingError{}
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

		// execure user context select {...}
		current.instance.Go(current)

		if current.state != disposing {
			panic(ExitFromContextWithoutFinishPanic)
		}

		// Remove node from parent childs and if parent is freezed and empty, initiate it disposing
		current.root.ready.Lock()
		if current.parent != nil {
			delete(current.parent.childs, current)
			if current.parent.state == freezed && len(current.parent.childs) == 0 {
				current.parent.state = disposing
				close(current.parent.controller)
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

// Finish ...
func (current *context[M]) Finish() {
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
		close(current.controller)
	}
}

// Send controller message to context
func (current *context[M]) Send(message M) (err error) {

	defer func() {
		if r := recover(); r != nil {
			err = &FinishInProcessForSendError{}
		}
	}()

	current.controller <- message

	return nil
}
